package store

import (
	"context"

	"github.com/example/notekit/internal/model"
)

// NotebookStore manages notebooks in the backing store.
type NotebookStore interface {
	Create(ctx context.Context, nb *model.Notebook) error
	GetByID(ctx context.Context, id string) (*model.Notebook, error)
	Update(ctx context.Context, nb *model.Notebook) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]*model.Notebook, error)
	Close() error
}

// NoteStore manages notes in the backing store.
type NoteStore interface {
	Create(ctx context.Context, note *model.Note) error
	GetByID(ctx context.Context, id string) (*model.Note, error)
	Update(ctx context.Context, note *model.Note) error
	Delete(ctx context.Context, id string) error // permanent delete
	SoftDelete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
	List(ctx context.Context, filter model.NoteFilter) ([]*model.Note, error)
	Count(ctx context.Context, filter model.NoteFilter) (int, error)
	Close() error
}
