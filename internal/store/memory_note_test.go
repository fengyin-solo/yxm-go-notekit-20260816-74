package store

import (
	"context"
	"testing"
	"time"

	"github.com/example/notekit/internal/model"
)

func TestNoteStoreCRUD(t *testing.T) {
	s, _ := NewMemoryNoteStore("", nil, 30*time.Second)
	defer s.Close()

	note := &model.Note{ID: "n1", NotebookID: "nb1", Title: "My Note", Content: "Hello world"}
	if err := s.Create(context.Background(), note); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := s.GetByID(context.Background(), "n1")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Title != "My Note" {
		t.Errorf("Title = %q", got.Title)
	}

	// Clone independence
	got.Title = "Changed"
	if orig, _ := s.GetByID(context.Background(), "n1"); orig.Title != "My Note" {
		t.Error("clone leaked title")
	}

	// Tags deep copy
	note2 := &model.Note{ID: "n2", NotebookID: "nb1", Title: "Tagged", Tags: []string{"go", "notes"}}
	s.Create(context.Background(), note2)
	got2, _ := s.GetByID(context.Background(), "n2")
	got2.Tags[0] = "hacked"
	if orig, _ := s.GetByID(context.Background(), "n2"); orig.Tags[0] != "go" {
		t.Error("tags clone leaked")
	}
}

func TestNoteStoreSoftDelete(t *testing.T) {
	s, _ := NewMemoryNoteStore("", nil, 30*time.Second)
	defer s.Close()

	s.Create(context.Background(), &model.Note{ID: "n1", NotebookID: "nb1", Title: "Note"})
	if err := s.SoftDelete(context.Background(), "n1"); err != nil {
		t.Fatalf("SoftDelete: %v", err)
	}

	// Should not appear in default (non-deleted) list
	list, _ := s.List(context.Background(), model.NoteFilter{})
	if len(list) != 0 {
		t.Errorf("deleted note should not appear in list, got %d", len(list))
	}

	// Should appear in include-deleted list
	list, _ = s.List(context.Background(), model.NoteFilter{IncludeDeleted: true})
	if len(list) != 1 {
		t.Errorf("include-deleted list should have 1, got %d", len(list))
	}

	// Restore
	if err := s.Restore(context.Background(), "n1"); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	list, _ = s.List(context.Background(), model.NoteFilter{})
	if len(list) != 1 {
		t.Errorf("after restore: list len = %d, want 1", len(list))
	}
}

func TestNoteStoreNoteCount(t *testing.T) {
	s, _ := NewMemoryNoteStore("", nil, 30*time.Second)
	defer s.Close()

	s.Create(context.Background(), &model.Note{ID: "n1", NotebookID: "nb1", Title: "A"})
	s.Create(context.Background(), &model.Note{ID: "n2", NotebookID: "nb1", Title: "B"})
	s.Create(context.Background(), &model.Note{ID: "n3", NotebookID: "nb2", Title: "C"})

	count, _ := s.Count(context.Background(), model.NoteFilter{NotebookID: "nb1"})
	if count != 2 {
		t.Errorf("nb1 count = %d, want 2", count)
	}
}
