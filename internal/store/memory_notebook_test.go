package store

import (
	"context"
	"testing"
	"time"

	"github.com/example/notekit/internal/model"
)

func TestNotebookStore(t *testing.T) {
	s, err := NewMemoryNotebookStore("", nil, 30*time.Second)
	if err != nil {
		t.Fatalf("NewMemoryNotebookStore: %v", err)
	}
	defer s.Close()

	nb := &model.Notebook{ID: "nb1", Name: "Work"}
	if err := s.Create(context.Background(), nb); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := s.GetByID(context.Background(), "nb1")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != "Work" {
		t.Errorf("Name = %q, want Work", got.Name)
	}

	// Clone independence
	got.Name = "Changed"
	if orig, _ := s.GetByID(context.Background(), "nb1"); orig.Name != "Work" {
		t.Error("clone leaked name")
	}

	list, err := s.List(context.Background())
	if err != nil || len(list) != 1 {
		t.Errorf("List len = %d", len(list))
	}

	if err := s.Delete(context.Background(), "nb1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := s.GetByID(context.Background(), "nb1"); err != model.ErrNotFound {
		t.Errorf("after delete: err = %v, want ErrNotFound", err)
	}
}

func TestNotebookStoreDuplicate(t *testing.T) {
	s, _ := NewMemoryNotebookStore("", nil, 30*time.Second)
	defer s.Close()
	s.Create(context.Background(), &model.Notebook{ID: "nb1", Name: "A"})
	if err := s.Create(context.Background(), &model.Notebook{ID: "nb1", Name: "B"}); err != model.ErrAlreadyExists {
		t.Errorf("duplicate ID: err = %v, want ErrAlreadyExists", err)
	}
}
