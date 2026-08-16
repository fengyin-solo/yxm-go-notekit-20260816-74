package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/notekit/internal/config"
	"github.com/example/notekit/internal/model"
	"github.com/example/notekit/internal/service"
	"github.com/example/notekit/internal/store"
)

func TestWrappedValidationErrorReturnsBadRequest(t *testing.T) {
	nbStore, _ := store.NewMemoryNotebookStore("", nil, 30*time.Second)
	noteStore, _ := store.NewMemoryNoteStore("", nil, 30*time.Second)
	t.Cleanup(func() {
		nbStore.Close()
		noteStore.Close()
	})
	cfg := config.Default()
	cfg.TitleMaxBytes = 80
	noteSvc := service.NewNoteService(noteStore, cfg)
	note, err := noteSvc.Create(context.Background(), &model.Note{NotebookID: "nb1", Title: "draft"})
	if err != nil {
		t.Fatalf("create note: %v", err)
	}
	router := NewRouter(
		NewNotebookHandler(service.NewNotebookService(nbStore, noteStore), 4096),
		NewNoteHandler(noteSvc, 4096),
	).Handler()

	req := httptest.NewRequest(http.MethodPut, "/api/v1/notes/"+note.ID, bytes.NewBufferString(`{"title":"   "}`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}
