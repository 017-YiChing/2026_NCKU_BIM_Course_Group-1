package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	db, err := OpenDB(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	s := NewServer(db, t.TempDir())
	return s
}

func TestIssueCRUD(t *testing.T) {
	s := newTestServer(t)
	mux := http.NewServeMux()
	s.routes(mux)

	// create（multipart，不帶截圖）
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	_ = w.WriteField("title", "門打不開")
	_ = w.WriteField("description", "FL1 門 M1 被牆卡住")
	_ = w.WriteField("reporter", "學生A")
	_ = w.WriteField("source", "revit")
	_ = w.WriteField("revit_element_id", "12345")
	_ = w.WriteField("view_name", "FL1 平面")
	_ = w.WriteField("revit_view_id", "abc-123")
	w.Close()
	req := httptest.NewRequest("POST", "/api/issues", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create = %d, body=%s", rec.Code, rec.Body.String())
	}
	var created Issue
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.ID == 0 || created.Status != "open" {
		t.Fatalf("unexpected created: %+v", created)
	}
	if created.ViewName != "FL1 平面" || created.RevitViewID != "abc-123" {
		t.Fatalf("view fields not round-tripped: %+v", created)
	}

	// list + status 篩選
	req = httptest.NewRequest("GET", "/api/issues?status=open", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list = %d", rec.Code)
	}
	var list []Issue
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("list len = %d", len(list))
	}

	// patch 狀態
	patchBody := bytes.NewBufferString(`{"status":"in_progress","assignee":"學生B"}`)
	req = httptest.NewRequest("PATCH", "/api/issues/1", patchBody)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("patch = %d, body=%s", rec.Code, rec.Body.String())
	}
	var patched Issue
	_ = json.Unmarshal(rec.Body.Bytes(), &patched)
	if patched.Status != "in_progress" || patched.Assignee != "學生B" {
		t.Fatalf("unexpected patched: %+v", patched)
	}

	// 非法狀態要 400
	bad := bytes.NewBufferString(`{"status":"nope"}`)
	req = httptest.NewRequest("PATCH", "/api/issues/1", bad)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad status = %d, want 400", rec.Code)
	}
}

func TestParseIFCIdentity(t *testing.T) {
	ifc := `ISO-10303-21;
#1=IFCWALL('2aXyzAAA00001',#42,'Wall-001',$,'Basic Wall 15cm',$,$);
#2=IFCDOOR('2aXyzAAA00002',#42,'Door-M1',$,'Single Door',$,$);
#10=IFCBUILDINGSTOREY('storey111',#42,'Level 1',$,$,$,$,$,$);
#20=IFCRELCONTAINEDINSPATIALSTRUCTURE('rel1',#42,$,$,(#1,#2),#10);
`
	p := filepath.Join(t.TempDir(), "a.ifc")
	if err := os.WriteFile(p, []byte(ifc), 0o644); err != nil {
		t.Fatal(err)
	}
	els, err := ParseIFCIdentity(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(els) != 2 {
		t.Fatalf("elements = %d, want 2: %+v", len(els), els)
	}
}
