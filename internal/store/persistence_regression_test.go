package store

import (
	"context"
	"testing"
	"time"

	"github.com/example/notekit/internal/model"
)

func TestNoteStoreReloadPreservesTagsAndNotebookCount(t *testing.T) {
	path := t.TempDir() + "/notes.json"
	first, err := NewMemoryNoteStore(path, nil, time.Hour)
	if err != nil {
		t.Fatalf("open first store: %v", err)
	}
	if err := first.Create(context.Background(), &model.Note{ID: "n1", NotebookID: "nb1", Title: "persist me", Tags: []string{"go", "ops"}}); err != nil {
		t.Fatalf("create note: %v", err)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("close first store: %v", err)
	}

	second, err := NewMemoryNoteStore(path, nil, time.Hour)
	if err != nil {
		t.Fatalf("open second store: %v", err)
	}
	defer second.Close()
	got, err := second.GetByID(context.Background(), "n1")
	if err != nil {
		t.Fatalf("get reloaded note: %v", err)
	}
	if len(got.Tags) != 2 || got.Tags[0] != "go" || got.Tags[1] != "ops" {
		t.Fatalf("reloaded tags = %#v", got.Tags)
	}
	count, err := second.Count(context.Background(), model.NoteFilter{NotebookID: "nb1"})
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("reloaded notebook count = %d, want 1", count)
	}
}
