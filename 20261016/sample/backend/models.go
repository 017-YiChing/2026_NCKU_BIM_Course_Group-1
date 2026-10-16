package main

// User mirrors the users table.
type User struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Role      string `json:"role"` // owner / client / admin
	AvatarURL string `json:"avatar_url"`
}

// AuthUser is the minimal user info carried in session / JWT.
type AuthUser struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// Issue mirrors SPEC Ch3 issues 表 (extended with user relations).
type Issue struct {
	ID            int64  `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Status        string `json:"status"` // open / in_progress / closed
	Reporter      string `json:"reporter"`
	ReporterID    int64  `json:"reporter_id"`
	ReporterRole  string `json:"reporter_role"`
	Assignee      string `json:"assignee"`
	AssigneeID    int64  `json:"assignee_id"`
	AssigneeRole  string `json:"assignee_role"`
	Source        string `json:"source"` // revit / ifc / manual
	RevitElementID string `json:"revit_element_id"`
	IFCGlobalID   string `json:"ifc_global_id"`
	ViewName      string `json:"view_name"`
	RevitViewID   string `json:"revit_view_id"`
	ModelName     string `json:"model_name"` // 議題所屬檔案/模型名（Revit 專案名或 IFC 檔名），樹狀分組用
	Archived      bool   `json:"archived"`     // 封存：預設列表隱藏，紀錄保留
	ModelID       *int64 `json:"model_id"`
	Screenshot    string `json:"screenshot"`
	Position      string `json:"position"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// Model mirrors SPEC Ch3 models 表.
type Model struct {
	ID           int64  `json:"id"`
	Filename     string `json:"filename"`
	UploadedAt   string `json:"uploaded_at"`
	ElementCount int    `json:"element_count"`
}

// Element mirrors SPEC Ch3 elements 表（只存身分，不存幾何）.
type Element struct {
	ModelID  int64  `json:"model_id"`
	GlobalID string `json:"global_id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Storey   string `json:"storey"`
}

var validStatus = map[string]bool{
	"open":        true,
	"in_progress": true,
	"closed":      true,
}

var validRole = map[string]bool{
	"owner":  true,
	"client": true,
	"admin":  true,
}

// gravatar-like avatar from name initials (demo, no external service)
func initialsAvatar(name string) string {
	runes := []rune(name)
	if len(runes) == 0 {
		return "?"
	}
	if len(runes) == 1 {
		return string(runes[0])
	}
	return string(runes[0]) + string(runes[len(runes)-1])
}
