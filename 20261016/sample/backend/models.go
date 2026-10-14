package main

// Issue mirrors SPEC Ch3 issues 表.
type Issue struct {
	ID            int64  `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Status        string `json:"status"` // open / in_progress / closed
	Reporter      string `json:"reporter"`
	Assignee      string `json:"assignee"`
	Source        string `json:"source"` // revit / ifc / manual
	RevitElementID string `json:"revit_element_id"`
	IFCGlobalID   string `json:"ifc_global_id"`
	ViewName      string `json:"view_name"`
	RevitViewID   string `json:"revit_view_id"`
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
