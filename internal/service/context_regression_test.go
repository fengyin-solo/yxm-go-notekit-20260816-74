package service

import (
	"context"
	"errors"
	"testing"

	"github.com/example/notekit/internal/model"
)

func TestNoteCreateHonorsCanceledContext(t *testing.T) {
	_, noteSvc := notekitTestService(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := noteSvc.Create(ctx, &model.Note{NotebookID: "nb1", Title: "late write"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Create with canceled context returned %v, want context.Canceled", err)
	}
}
