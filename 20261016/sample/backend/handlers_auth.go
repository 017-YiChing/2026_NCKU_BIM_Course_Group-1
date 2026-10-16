package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

// generateToken creates a random session token.
func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// currentUserFromRequest returns the authenticated user from the request,
// or (nil, false) if anonymous.
func currentUserFromRequest(db *sql.DB, r *http.Request) (*User, bool, string) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return nil, false, "missing bearer token"
	}
	token := strings.TrimPrefix(auth, "Bearer ")
	var u User
	err := db.QueryRow(`
		SELECT u.id, u.email, u.name, u.role, u.avatar_url
		FROM sessions s JOIN users u ON u.id = s.user_id
		WHERE s.token=?`, token).Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.AvatarURL)
	if err != nil {
		log.Printf("AUTH DEBUG: token=%q scan err=%v", token, err)
		return nil, false, err.Error()
	}
	log.Printf("AUTH DEBUG: authenticated user=%d %s %s", u.ID, u.Email, u.Role)
	return &u, true, ""
}

// POST /api/auth/register
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "JSON 解析失敗: "+err.Error())
		return
	}
	if body.Email == "" || body.Password == "" || body.Name == "" {
		writeErr(w, http.StatusBadRequest, "email、password、name 皆必填")
		return
	}
	if body.Role == "" {
		body.Role = "client"
	}
	if !validRole[body.Role] {
		writeErr(w, http.StatusBadRequest, "role 須為 owner/client/admin")
		return
	}
	var existing int
	s.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE email=?`, body.Email).Scan(&existing)
	if existing > 0 {
		writeErr(w, http.StatusBadRequest, "該 Email 已註冊")
		return
	}
	res, err := s.DB.Exec(`INSERT INTO users (email, password_hash, name, role) VALUES (?,?,?,?)`,
		body.Email, simpleHash(body.Password), body.Name, body.Role)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	id, _ := res.LastInsertId()
	writeJSON(w, http.StatusCreated, User{ID: id, Email: body.Email, Name: body.Name, Role: body.Role})
}

// POST /api/auth/login
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "JSON 解析失敗: "+err.Error())
		return
	}
	var id int64
	var email, passwordHash, name, role string
	err := s.DB.QueryRow(`SELECT id, email, password_hash, name, role FROM users WHERE email=?`,
		body.Email).Scan(&id, &email, &passwordHash, &name, &role)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "Email 或密碼錯誤")
		return
	}
	if simpleHash(body.Password) != passwordHash {
		writeErr(w, http.StatusUnauthorized, "Email 或密碼錯誤")
		return
	}
	u := User{ID: id, Email: email, Name: name, Role: role}
	token := generateToken()
	expires := time.Now().Add(24 * 7 * time.Hour).Format("2006-01-02 15:04:05")
	_, err = s.DB.Exec(`INSERT OR REPLACE INTO sessions (token, user_id, expires_at) VALUES (?,?,?)`,
		token, u.ID, expires)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user":  User{ID: u.ID, Email: u.Email, Name: u.Name, Role: u.Role},
	})
}

// POST /api/auth/logout
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		token := strings.TrimPrefix(auth, "Bearer ")
		s.DB.Exec(`DELETE FROM sessions WHERE token=?`, token)
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true"})
}

// GET /api/auth/me
func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	u, ok, errMsg := currentUserFromRequest(s.DB, r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "未登入："+errMsg)
		return
	}
	writeJSON(w, http.StatusOK, *u)
}

// GET /api/users — list all users (for assignee dropdown)
func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.Query(`SELECT id, email, name, role, avatar_url FROM users ORDER BY name`)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	out := []User{}
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.AvatarURL); err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		out = append(out, u)
	}
	writeJSON(w, http.StatusOK, out)
}
