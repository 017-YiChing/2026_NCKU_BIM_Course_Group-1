package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// POST /api/models — multipart: file(.ifc)。存檔＋解析身分＋回 element_count。
func (s *Server) handleUploadModel(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		writeErr(w, http.StatusBadRequest, "parse form: "+err.Error())
		return
	}
	f, hdr, err := r.FormFile("file")
	if err != nil {
		writeErr(w, http.StatusBadRequest, "缺少 file 欄位（.ifc）")
		return
	}
	defer f.Close()
	if err := os.MkdirAll(s.ModelDir, 0o755); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	stored := fmt.Sprintf("%d_%s", time.Now().UnixNano(), hdr.Filename)
	dstPath := filepath.Join(s.ModelDir, stored)
	out, err := os.Create(dstPath)
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

	els, err := ParseIFCIdentity(dstPath)
	if err != nil {
		writeErr(w, 500, "IFC 解析失敗: "+err.Error())
		return
	}

	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	res, err := s.DB.Exec(`INSERT INTO models (filename, uploaded_at, element_count) VALUES (?,?,?)`,
		hdr.Filename, now, len(els))
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	mid, _ := res.LastInsertId()
	for _, e := range els {
		_, _ = s.DB.Exec(`INSERT OR IGNORE INTO elements (model_id, global_id, type, name, storey)
			VALUES (?,?,?,?,?)`, mid, e.GlobalID, e.Type, e.Name, e.Storey)
	}
	writeJSON(w, http.StatusCreated, Model{
		ID: mid, Filename: hdr.Filename, UploadedAt: now, ElementCount: len(els),
	})
}

// GET /api/models/{id}/file — 下載 IFC（前端 viewer 用；第一版 viewer 先不接，但端點留著）。
func (s *Server) handleDownloadModel(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "id 須為數字")
		return
	}
	var filename string
	err = s.DB.QueryRow(`SELECT filename FROM models WHERE id=?`, id).Scan(&filename)
	if err != nil {
		writeErr(w, http.StatusNotFound, "找不到模型")
		return
	}
	// 找出實體檔（檔名帶時間戳前綴，尾綴比對）
	matches, _ := filepath.Glob(filepath.Join(s.ModelDir, "*_"+filename))
	if len(matches) == 0 {
		// 退回：直接找最新同名（上傳時原名保存，實體帶前綴）
		writeErr(w, http.StatusNotFound, "模型檔案遺失")
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	http.ServeFile(w, r, matches[len(matches)-1])
}

// GET /api/models/{id}/elements — 元件身分列表。
func (s *Server) handleListElements(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "id 須為數字")
		return
	}
	rows, err := s.DB.Query(`SELECT model_id, global_id, type, name, storey
		FROM elements WHERE model_id=? ORDER BY type, name LIMIT 2000`, id)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	defer rows.Close()
	out := []Element{}
	for rows.Next() {
		var e Element
		if err := rows.Scan(&e.ModelID, &e.GlobalID, &e.Type, &e.Name, &e.Storey); err != nil {
			writeErr(w, 500, err.Error())
			return
		}
		out = append(out, e)
	}
	writeJSON(w, http.StatusOK, out)
}
