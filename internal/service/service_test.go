package service

import (
	"context"
	"testing"
	"time"

	"github.com/example/notekit/internal/config"
	"github.com/example/notekit/internal/model"
	"github.com/example/notekit/internal/store"
)

func notekitTestService(t *testing.T) (*NotebookService, *NoteService) {
	t.Helper()
	nbStore, _ := store.NewMemoryNotebookStore("", nil, 30*time.Second)
	noteStore, _ := store.NewMemoryNoteStore("", nil, 30*time.Second)
	t.Cleanup(func() {
		nbStore.Close()
		noteStore.Close()
	})
	cfg := config.Default()
	cfg.TitleMaxBytes = 200
	nbSvc := NewNotebookService(nbStore, noteStore)
	noteSvc := NewNoteService(noteStore, cfg)
	return nbSvc, noteSvc
}

func TestNotebookCreate(t *testing.T) {
	nbSvc, _ := notekitTestService(t)
	nb, err := nbSvc.Create(context.Background(), &model.Notebook{Name: "  My Notebook  "})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if nb.ID == "" {
		t.Error("ID should be generated")
	}
	if nb.Name != "My Notebook" {
		t.Errorf("Name = %q, want normalized", nb.Name)
	}
}

func TestNotebookCreateWithoutName(t *testing.T) {
	nbSvc, _ := notekitTestService(t)
	if _, err := nbSvc.Create(context.Background(), &model.Notebook{Name: ""}); err == nil {
		t.Error("empty name should error")
	}
}

func TestNotebookDeleteCascade(t *testing.T) {
	nbSvc, noteSvc := notekitTestService(t)
	nb, _ := nbSvc.Create(context.Background(), &model.Notebook{Name: "Work"})
	noteSvc.Create(context.Background(), &model.Note{NotebookID: nb.ID, Title: "Note 1"})
	noteSvc.Create(context.Background(), &model.Note{NotebookID: nb.ID, Title: "Note 2"})

	if err := nbSvc.Delete(context.Background(), nb.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := noteSvc.GetByID(context.Background(), "note1id"); err == nil {
		t.Error("notes should be cascade-deleted")
	}
}

func TestNoteCreate(t *testing.T) {
	_, noteSvc := notekitTestService(t)
	note, err := noteSvc.Create(context.Background(), &model.Note{
		NotebookID: "nb1",
		Title:      "  Hello  ",
		Content:    "  world  ",
		Tags:       []string{"  Go  ", "  NOTES  "},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if note.ID == "" {
		t.Error("ID should be generated")
	}
	if note.Title != "Hello" {
		t.Errorf("Title = %q", note.Title)
	}
	if len(note.Tags) != 2 || note.Tags[0] != "go" || note.Tags[1] != "notes" {
		t.Errorf("Tags = %v", note.Tags)
	}
}

func TestNoteCreateEmptyTitle(t *testing.T) {
	_, noteSvc := notekitTestService(t)
	if _, err := noteSvc.Create(context.Background(), &model.Note{NotebookID: "nb1", Title: "   "}); err == nil {
		t.Error("empty title should error")
	}
}

func TestNoteSoftDelete(t *testing.T) {
	_, noteSvc := notekitTestService(t)
	note, _ := noteSvc.Create(context.Background(), &model.Note{NotebookID: "nb1", Title: "To trash"})
	if err := noteSvc.SoftDelete(context.Background(), note.ID); err != nil {
		t.Fatalf("SoftDelete: %v", err)
	}
	list, _ := noteSvc.List(context.Background(), model.NoteFilter{})
	if len(list) != 0 {
		t.Errorf("soft-deleted note in list, got %d", len(list))
	}
	if err := noteSvc.Restore(context.Background(), note.ID); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	list, _ = noteSvc.List(context.Background(), model.NoteFilter{})
	if len(list) != 1 {
		t.Errorf("after restore: len = %d, want 1", len(list))
	}
}

func TestNoteFilter(t *testing.T) {
	_, noteSvc := notekitTestService(t)
	noteSvc.Create(context.Background(), &model.Note{NotebookID: "nb1", Title: "Go tutorial", Content: "Learn Go", Tags: []string{"go", "tutorial"}})
	noteSvc.Create(context.Background(), &model.Note{NotebookID: "nb1", Title: "Python guide", Content: "Learn Python", Tags: []string{"python"}})
	noteSvc.Create(context.Background(), &model.Note{NotebookID: "nb2", Title: "Go notes", Content: "Notes", Tags: []string{"go"}})

	// Filter by tag
	list, _ := noteSvc.List(context.Background(), model.NoteFilter{Tags: []string{"go"}})
	if len(list) != 2 {
		t.Errorf("tag=go: got %d, want 2", len(list))
	}

	// Filter by query
	list, _ = noteSvc.List(context.Background(), model.NoteFilter{Query: "tutorial"})
	if len(list) != 1 {
		t.Errorf("query=tutorial: got %d, want 1", len(list))
	}

	// Filter by notebook
	list, _ = noteSvc.List(context.Background(), model.NoteFilter{NotebookID: "nb1"})
	if len(list) != 2 {
		t.Errorf("notebook=nb1: got %d, want 2", len(list))
	}
}
