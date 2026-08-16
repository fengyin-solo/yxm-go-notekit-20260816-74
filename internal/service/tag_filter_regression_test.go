package service

import (
	"context"
	"testing"

	"github.com/example/notekit/internal/model"
)

func TestListRequiresEveryRequestedTag(t *testing.T) {
	_, noteSvc := notekitTestService(t)
	ctx := context.Background()
	if _, err := noteSvc.Create(ctx, &model.Note{NotebookID: "nb1", Title: "Release checklist", Tags: []string{"go", "api"}}); err != nil {
		t.Fatalf("create release note: %v", err)
	}
	if _, err := noteSvc.Create(ctx, &model.Note{NotebookID: "nb1", Title: "Go scratchpad", Tags: []string{"go"}}); err != nil {
		t.Fatalf("create scratch note: %v", err)
	}

	got, err := noteSvc.List(ctx, model.NoteFilter{NotebookID: "nb1", Tags: []string{"go", "api"}})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 1 || got[0].Title != "Release checklist" {
		t.Fatalf("multi-tag filter returned %d notes: %#v", len(got), got)
	}
}
