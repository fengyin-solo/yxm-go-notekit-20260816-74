package service

import (
	"context"
	"errors"
	"testing"

	"github.com/example/notekit/internal/model"
)

func TestNotebookDeleteRemovesNotesPermanently(t *testing.T) {
	nbSvc, noteSvc := notekitTestService(t)
	ctx := context.Background()
	nb, err := nbSvc.Create(ctx, &model.Notebook{Name: "Ops"})
	if err != nil {
		t.Fatalf("create notebook: %v", err)
	}
	note, err := noteSvc.Create(ctx, &model.Note{NotebookID: nb.ID, Title: "rotation plan"})
	if err != nil {
		t.Fatalf("create note: %v", err)
	}

	if err := nbSvc.Delete(ctx, nb.ID); err != nil {
		t.Fatalf("delete notebook: %v", err)
	}
	if _, err := noteSvc.GetByID(ctx, note.ID); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("deleted notebook left note readable: %v", err)
	}
	left, err := noteSvc.List(ctx, model.NoteFilter{NotebookID: nb.ID, IncludeDeleted: true})
	if err != nil {
		t.Fatalf("list notes: %v", err)
	}
	if len(left) != 0 {
		t.Fatalf("deleted notebook left %d notes behind", len(left))
	}
}
