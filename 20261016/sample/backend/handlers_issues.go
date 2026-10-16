package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	DB            *sql.DB
	DataDir       string
	ScreenshotDir string
	ModelDir      string
}

func NewServer(db *sql.DB, dataDir string) *Server {
	return &Server{
		DB:            db,
		DataDir:       dataDir,
		ScreenshotDir: filepath.Join(dataDir, "screenshots"),
		ModelDir:      filepath.Join(dataDir, "models"),
	}
}

func (s *Server) routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/issues", s.handleCreateIssue)
	mux.HandleFunc("GET /api/issues", s.handleListIssues)
	mux.HandleFunc("GET /api/issues/{id}", s.handleGetIssue)
	mux.HandleFunc("PATCH /api/issues/{id}", s.handlePatchIssue)
	mux.HandleFunc("DELETE /api/issues/{id}", s.handleDeleteIssue)
	mux.HandleFunc("POST /api/models", s.handleUploadModel)
	mux.HandleFunc("GET /api/models/{id}/file", s.handleDownloadModel)
	mux.HandleFunc("GET /api/models/{id}/elements", s.handleListElements)
	mux.HandleFunc("GET /api/users", s.handleListUsers)
	mux.HandleFunc("POST /api/auth/register", s.handleRegister)
	mux.HandleFunc("POST /api/auth/login", s.handleLogin)
	mux.HandleFunc("/api/auth/login-diagnostic", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"method": r.Method, "path": r.URL.Path})
	})
	mux.HandleFunc("POST /api/auth/logout", s.handleLogout)
	mux.HandleFunc("GET /api/auth/me", s.handleMe)
	mux.Handle("GET /screenshots/", http.StripPrefix("/screenshots/", http.FileServer(http.Dir(s.ScreenshotDir))))
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"ok": "true"})
	})
}

// canModifyIssue returns true if the user can modify the issue:
// - owner / admin can modify everything
// - reporter or assignee can modify their own issues
// - anonymous cannot modify
func (s *Server) canModifyIssue(user *User, issueID int64) bool {
	if user == nil {
		return false
	}
	if user.Role == "owner" || user.Role == "admin" {
		return true
	}
	var reporterID, assigneeID sql.NullInt64
	s.DB.QueryRow(`SELECT reporter_id, assignee_id FROM issues WHERE id=?`, issueID).Scan(&reporterID, &assigneeID)
	if reporterID.Valid && reporterID.Int64 == user.ID {
		return true
	}
	if assigneeID.Valid && assigneeInt64(assigneeID) == user.ID {
		return true
	}
	return false
}

func assigneeInt64(n sql.NullInt64) int64 {
	if n.Valid {
		return n.Int64
	}
	return 0
}

// POST /api/issues — multipart: 欄位 + screenshot(可選)
func (s *Server) handleCreateIssue(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(20 << 20); err != nil && err != http.ErrNotMultipart {
		writeErr(w, http.StatusBadRequest, "parse form: "+err.Error())
		return
	}
	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		writeErr(w, http.StatusBadRequest, "title 必填")
		return
	}
	status := r.FormValue("status")
	if status == "" {
		status = "open"
	}
	if !validStatus[status] {
		writeErr(w, http.StatusBadRequest, "status 須為 open/in_progress/closed")
		return
	}
	source := r.FormValue("source")
	if source == "" {
		source = "manual"
	}
	reporter := strings.TrimSpace(r.FormValue("reporter"))

	var reporterID int64
	if u, ok, _ := currentUserFromRequest(s.DB, r); ok {
		reporterID = u.ID
		if reporter == "" {
			reporter = u.Name
		}
	}

	var modelID *int64
	if v := strings.TrimSpace(r.FormValue("model_id")); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "model_id 須為數字")
			return
		}
		modelID = &n
	}

	screenshotName := ""
	if f, hdr, err := r.FormFile("screenshot"); err == nil {
		defer f.Close()
		if err := os.MkdirAll(s.ScreenshotDir, 0o755); err != nil {
			writeErr(w, 500, err.Error())
			return
		}
		ext := filepath.Ext(hdr.Filename)
		if ext == "" {
			ext = ".png"
		}
		screenshotName = fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		out, err := os.Create(filepath.Join(s.ScreenshotDir, screenshotName))
		if err != nil {
			writeErr(w, 500, err.Error())
			return
		}
		if _, err := io.Copy(out, f); err != nil {
			out.Close()
			writeErr(w, 500, err.Error())
			return
		}
		out.Close()
	}

	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	res, err := s.DB.Exec(`INSERT INTO issues
		(title,description,status,reporter,reporter_id,assignee,assignee_id,source,revit_element_id,ifc_global_id,model_id,screenshot,position,view_name,revit_view_id,model_name,created_at,updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?, ?, ?)`,
		title, r.FormValue("description"), status, reporter, reporterID,
		r.FormValue("assignee"), 0, source,
		r.FormValue("revit_element_id"), r.FormValue("ifc_global_id"), modelID,
		screenshotName, r.FormValue("position"),
		r.FormValue("view_name"), r.FormValue("revit_view_id"), r.FormValue("model_name"), now, now)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	id, _ := res.LastInsertId()
	issue, err := getIssueByID(s.DB, id)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, issue)
}

// GET /api/issues?status=open&archived=1&me=1
func (s *Server) handleListIssues(w http.ResponseWriter, r *http.Request) {
	user, _, _ := currentUserFromRequest(s.DB, r)
	base := `SELECT i.id,i.title,i.description,i.status,i.reporter,i.reporter_id,i.assignee,i.assignee_id,i.source,
		COALESCE(i.revit_element_id,''),COALESCE(i.ifc_global_id,''),i.model_id,
		COALESCE(i.screenshot,''),COALESCE(i.position,''),
		COALESCE(i.view_name,''),COALESCE(i.revit_view_id,''),
		COALESCE(NULLIF(i.model_name,''), m.filename, ''),
		i.archived,i.created_at,i.updated_at
		FROM issues i LEFT JOIN models m ON m.id = i.model_id`
	var conds []string
	var args []any
	if st := r.URL.Query().Get("status"); st != "" {
		if !validStatus[st] {
			writeErr(w, http.StatusBadRequest, "status 須為 open/in_progress/closed")
			return
		}
		conds = append(conds, "i.status=?")
		args = append(args, st)
	}
	if r.URL.Query().Get("me") == "1" && user != nil {
		conds = append(conds, "(i.reporter_id=? OR i.assignee_id=?)")
		args = append(args, user.ID, user.ID)
	}
	switch r.URL.Query().Get("archived") {
	case "1":
		conds = append(conds, "i.archived=1")
	case "all":
		// 不篩
	default:
		conds = append(conds, "i.archived=0")
	}
	q := base
	if len(conds) > 0 {
		q += " WHERE " + strings.Join(conds, " AND ")
	}
	q += " ORDER BY i.id DESC"
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	defer rows.Close()
	out := []Issue{}
	for rows.Next() {
		var it Issue
		var repID, assID sql.NullInt64
		if err := rows.Scan(&it.ID, &it.Title, &it.Description, &it.Status,
			&it.Reporter, &repID, &it.Assignee, &assID, &it.Source,
			&it.RevitElementID, &it.IFCGlobalID, &it.ModelID,
			&it.Screenshot, &it.Position,
			&it.ViewName, &it.RevitViewID,
			&it.ModelName, &it.Archived,
			&it.CreatedAt, &it.UpdatedAt); err != nil {
			writeErr(w, 500, err.Error())
			return
		}
		if repID.Valid {
			it.ReporterID = repID.Int64
		}
		if assID.Valid {
			it.AssigneeID = assID.Int64
		}
		// 填補 reporter/assignee 名稱與角色（若 DB 欄位為空則從 users 表查）
		if it.Reporter == "" || it.ReporterRole == "" {
			if it.ReporterID != 0 {
				var name, role string
				s.DB.QueryRow(`SELECT name, role FROM users WHERE id=?`, it.ReporterID).Scan(&name, &role)
				if it.Reporter == "" {
					it.Reporter = name
				}
				it.ReporterRole = role
			}
		}
		if it.Assignee == "" || it.AssigneeRole == "" {
			if it.AssigneeID != 0 {
				var name, role string
				s.DB.QueryRow(`SELECT name, role FROM users WHERE id=?`, it.AssigneeID).Scan(&name, &role)
				if it.Assignee == "" {
					it.Assignee = name
				}
				it.AssigneeRole = role
			}
		}
		out = append(out, it)
	}
	writeJSON(w, http.StatusOK, out)
}

// GET /api/issues/{id}
func (s *Server) handleGetIssue(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "id 須為數字")
		return
	}
	it, err := getIssueByID(s.DB, id)
	if err == sql.ErrNoRows {
		writeErr(w, http.StatusNotFound, "找不到議題")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, it)
}

// PATCH /api/issues/{id} — JSON: {status?, assignee?, archived?}
func (s *Server) handlePatchIssue(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "id 須為數字")
		return
	}
	user, _, _ := currentUserFromRequest(s.DB, r)
	if !s.canModifyIssue(user, id) {
		writeErr(w, http.StatusForbidden, "無權限修改此議題（非建立者/指派人/管理員）")
		return
	}
	var body struct {
		Status   *string          `json:"status"`
		Assignee *json.RawMessage `json:"assignee"`
		Archived *bool            `json:"archived"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "JSON 解析失敗: "+err.Error())
		return
	}
	if body.Status != nil && !validStatus[*body.Status] {
		writeErr(w, http.StatusBadRequest, "status 須為 open/in_progress/closed")
		return
	}
	if body.Status == nil && body.Assignee == nil && body.Archived == nil {
		writeErr(w, http.StatusBadRequest, "至少要帶 status、assignee 或 archived")
		return
	}
	if body.Status != nil {
		if _, err := s.DB.Exec(`UPDATE issues SET status=?, updated_at=datetime('now') WHERE id=?`, *body.Status, id); err != nil {
			writeErr(w, 500, err.Error())
			return
		}
	}
	if body.Assignee != nil {
		// 解析指派人：可傳入數字（user id）或字串（姓名/email）
		var assigneeRaw interface{}
		if err := json.Unmarshal(*body.Assignee, &assigneeRaw); err != nil {
			writeErr(w, http.StatusBadRequest, "assignee 格式錯誤")
			return
		}
		assignee := ""
		assigneeID := int64(0)
		switch v := assigneeRaw.(type) {
		case float64:
			n := int64(v)
			var exists int
			s.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE id=?`, n).Scan(&exists)
			if exists > 0 {
				assigneeID = n
				var name string
				s.DB.QueryRow(`SELECT name FROM users WHERE id=?`, n).Scan(&name)
				assignee = name
			}
		case string:
			assignee = v
			var uid int64
			s.DB.QueryRow(`SELECT id FROM users WHERE name=? OR email=?`, v, v).Scan(&uid)
			if uid > 0 {
				assigneeID = uid
			}
		}
		if _, err := s.DB.Exec(`UPDATE issues SET assignee=?, assignee_id=?, updated_at=datetime('now') WHERE id=?`,
			assignee, assigneeID, id); err != nil {
			writeErr(w, 500, err.Error())
			return
		}
	}
	if body.Archived != nil {
		v := 0
		if *body.Archived {
			v = 1
		}
		if _, err := s.DB.Exec(`UPDATE issues SET archived=?, updated_at=datetime('now') WHERE id=?`, v, id); err != nil {
			writeErr(w, 500, err.Error())
			return
		}
	}
	it, err := getIssueByID(s.DB, id)
	if err == sql.ErrNoRows {
		writeErr(w, http.StatusNotFound, "找不到議題")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, it)
}

// DELETE /api/issues/{id} — only owner/admin or reporter
func (s *Server) handleDeleteIssue(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "id 須為數字")
		return
	}
	user, _, _ := currentUserFromRequest(s.DB, r)
	if user == nil {
		writeErr(w, http.StatusForbidden, "需登入才能刪除")
		return
	}
	if user.Role != "owner" && user.Role != "admin" {
		var reporterID sql.NullInt64
		s.DB.QueryRow(`SELECT reporter_id FROM issues WHERE id=?`, id).Scan(&reporterID)
		if !reporterID.Valid || reporterID.Int64 != user.ID {
			writeErr(w, http.StatusForbidden, "只能刪除自己建立的議題")
			return
		}
	}
	if _, err := s.DB.Exec(`DELETE FROM issues WHERE id=?`, id); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true"})
}

func getIssueByID(db *sql.DB, id int64) (Issue, error) {
	var it Issue
	var repID, assID sql.NullInt64
	err := db.QueryRow(`SELECT i.id,i.title,i.description,i.status,i.reporter,i.reporter_id,i.assignee,i.assignee_id,i.source,
		COALESCE(i.revit_element_id,''),COALESCE(i.ifc_global_id,''),i.model_id,
		COALESCE(i.screenshot,''),COALESCE(i.position,''),
		COALESCE(i.view_name,''),COALESCE(i.revit_view_id,''),
		COALESCE(NULLIF(i.model_name,''), m.filename, ''),
		i.archived,i.created_at,i.updated_at
		FROM issues i LEFT JOIN models m ON m.id = i.model_id WHERE i.id=?`, id).Scan(
		&it.ID, &it.Title, &it.Description, &it.Status, &it.Reporter, &repID, &it.Assignee, &assID, &it.Source,
		&it.RevitElementID, &it.IFCGlobalID, &it.ModelID,
		&it.Screenshot, &it.Position,
		&it.ViewName, &it.RevitViewID,
		&it.ModelName, &it.Archived, &it.CreatedAt, &it.UpdatedAt)
	if err != nil {
		return it, err
	}
	if repID.Valid {
		it.ReporterID = repID.Int64
	}
	if assID.Valid {
		it.AssigneeID = assID.Int64
	}
	if it.ReporterID != 0 {
		var name, role string
		db.QueryRow(`SELECT name, role FROM users WHERE id=?`, it.ReporterID).Scan(&name, &role)
		if it.Reporter == "" {
			it.Reporter = name
		}
		it.ReporterRole = role
	}
	if it.AssigneeID != 0 {
		var name, role string
		db.QueryRow(`SELECT name, role FROM users WHERE id=?`, it.AssigneeID).Scan(&name, &role)
		if it.Assignee == "" {
			it.Assignee = name
		}
		it.AssigneeRole = role
	}
	return it, err
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
