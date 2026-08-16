package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/notekit/internal/config"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	cfg := config.Default()
	cfg.DataFile = ""
	cfg.AuthToken = ""
	cfg.RateLimit = 0
	srv, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	t.Cleanup(func() {
		srv.Shutdown(t.Context())
	})
	return srv
}

func doJSON(t *testing.T, h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		payload, _ := json.Marshal(body)
		reader = bytes.NewReader(payload)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHealthz(t *testing.T) {
	cfg := config.Default()
	cfg.DataFile = ""
	cfg.AuthToken = ""
	cfg.RateLimit = 0
	srv, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	defer srv.Shutdown(t.Context())

	mux := srv.httpServer.Handler.(http.Handler)
	rec := doJSON(t, mux, http.MethodGet, "/healthz", nil)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestNotebookCRUD(t *testing.T) {
	cfg := config.Default()
	cfg.DataFile = ""
	cfg.AuthToken = ""
	cfg.RateLimit = 0
	srv, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	defer srv.Shutdown(t.Context())
	mux := srv.httpServer.Handler.(http.Handler)

	// Create
	rec := doJSON(t, mux, http.MethodPost, "/api/v1/notebooks", map[string]string{"name": "Work"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d body=%s", rec.Code, rec.Body.String())
	}
	var nb map[string]any
	json.Unmarshal(rec.Body.Bytes(), &nb)
	id := nb["id"].(string)

	// Get
	rec = doJSON(t, mux, http.MethodGet, "/api/v1/notebooks/"+id, nil)
	if rec.Code != http.StatusOK {
		t.Errorf("get status = %d", rec.Code)
	}

	// List
	rec = doJSON(t, mux, http.MethodGet, "/api/v1/notebooks", nil)
	if rec.Code != http.StatusOK {
		t.Errorf("list status = %d", rec.Code)
	}

	// Update
	rec = doJSON(t, mux, http.MethodPut, "/api/v1/notebooks/"+id, map[string]string{"name": "Personal"})
	if rec.Code != http.StatusOK {
		t.Errorf("update status = %d", rec.Code)
	}

	// Delete
	rec = doJSON(t, mux, http.MethodDelete, "/api/v1/notebooks/"+id, nil)
	if rec.Code != http.StatusNoContent {
		t.Errorf("delete status = %d", rec.Code)
	}
}

func TestNoteCRUD(t *testing.T) {
	cfg := config.Default()
	cfg.DataFile = ""
	cfg.AuthToken = ""
	cfg.RateLimit = 0
	srv, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	defer srv.Shutdown(t.Context())
	mux := srv.httpServer.Handler.(http.Handler)

	// Create notebook first
	rec := doJSON(t, mux, http.MethodPost, "/api/v1/notebooks", map[string]string{"name": "Work"})
	var nb map[string]any
	json.Unmarshal(rec.Body.Bytes(), &nb)
	nbID := nb["id"].(string)

	// Create note
	rec = doJSON(t, mux, http.MethodPost, "/api/v1/notes", map[string]any{
		"notebook_id": nbID,
		"title":       "My first note",
		"content":     "Hello, world!",
		"tags":        []string{"intro", "hello"},
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create note status = %d body=%s", rec.Code, rec.Body.String())
	}
	var note map[string]any
	json.Unmarshal(rec.Body.Bytes(), &note)
	noteID := note["id"].(string)

	// Get
	rec = doJSON(t, mux, http.MethodGet, "/api/v1/notes/"+noteID, nil)
	if rec.Code != http.StatusOK {
		t.Errorf("get note status = %d", rec.Code)
	}

	// Update
	rec = doJSON(t, mux, http.MethodPut, "/api/v1/notes/"+noteID, map[string]string{"title": "Updated title"})
	if rec.Code != http.StatusOK {
		t.Errorf("update note status = %d", rec.Code)
	}

	// Soft delete
	rec = doJSON(t, mux, http.MethodPost, "/api/v1/notes/"+noteID+"/trash", nil)
	if rec.Code != http.StatusNoContent {
		t.Errorf("soft delete status = %d", rec.Code)
	}

	// Should be gone from default list
	rec = doJSON(t, mux, http.MethodGet, "/api/v1/notes", nil)
	var notes []any
	json.Unmarshal(rec.Body.Bytes(), &notes)
	if len(notes) != 0 {
		t.Errorf("soft-deleted note in list, got %d", len(notes))
	}

	// Restore
	rec = doJSON(t, mux, http.MethodPost, "/api/v1/notes/"+noteID+"/restore", nil)
	if rec.Code != http.StatusNoContent {
		t.Errorf("restore status = %d", rec.Code)
	}

	// Permanent delete
	rec = doJSON(t, mux, http.MethodDelete, "/api/v1/notes/"+noteID, nil)
	if rec.Code != http.StatusNoContent {
		t.Errorf("delete note status = %d", rec.Code)
	}
}
