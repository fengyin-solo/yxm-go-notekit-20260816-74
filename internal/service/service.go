package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/example/notekit/internal/config"
	"github.com/example/notekit/internal/model"
	"github.com/example/notekit/internal/store"
)

// NotebookService handles business logic for notebooks.
type NotebookService struct {
	store     store.NotebookStore
	noteStore store.NoteStore
	now       func() time.Time
}

// NewNotebookService creates a new NotebookService.
func NewNotebookService(store store.NotebookStore, noteStore store.NoteStore) *NotebookService {
	return &NotebookService{store: store, noteStore: noteStore, now: time.Now}
}

// Create creates a new notebook.
func (s *NotebookService) Create(ctx context.Context, nb *model.Notebook) (*model.Notebook, error) {
	nb.Normalize()
	if nb.Name == "" {
		return nil, fmt.Errorf("%w: name is required", model.ErrInvalidInput)
	}
	nb.ID = newID()
	nb.CreatedAt = s.now().UTC()
	nb.UpdatedAt = nb.CreatedAt
	nb.NoteCount = 0
	if err := s.store.Create(ctx, nb); err != nil {
		return nil, err
	}
	return nb.Clone(), nil
}

// GetByID returns a notebook by ID.
func (s *NotebookService) GetByID(ctx context.Context, id string) (*model.Notebook, error) {
	nb, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// Update live note count
	count, _ := s.noteStore.Count(ctx, model.NoteFilter{NotebookID: id, IncludeDeleted: false})
	nb.NoteCount = count
	return nb, nil
}

// Update updates an existing notebook.
func (s *NotebookService) Update(ctx context.Context, id string, req *UpdateNotebookRequest) (*model.Notebook, error) {
	nb, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	req.Apply(nb)
	nb.Normalize()
	if nb.Name == "" {
		return nil, fmt.Errorf("%w: name is required", model.ErrInvalidInput)
	}
	nb.UpdatedAt = s.now().UTC()
	if err := s.store.Update(ctx, nb); err != nil {
		return nil, err
	}
	return nb.Clone(), nil
}

// Delete deletes a notebook (cascades to notes).
func (s *NotebookService) Delete(ctx context.Context, id string) error {
	// Cascade delete all notes in this notebook
	notes, _ := s.noteStore.List(ctx, model.NoteFilter{NotebookID: id, IncludeDeleted: true})
	for _, n := range notes {
		s.noteStore.Delete(ctx, n.ID)
	}
	return s.store.Delete(ctx, id)
}

// List returns all notebooks.
func (s *NotebookService) List(ctx context.Context) ([]*model.Notebook, error) {
	notebooks, err := s.store.List(ctx)
	if err != nil {
		return nil, err
	}
	// Populate live note counts
	for _, nb := range notebooks {
		count, _ := s.noteStore.Count(ctx, model.NoteFilter{NotebookID: nb.ID, IncludeDeleted: false})
		nb.NoteCount = count
	}
	return notebooks, nil
}

// UpdateNotebookRequest describes fields that can be updated on a notebook.
type UpdateNotebookRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// Apply applies the update request to the notebook.
func (r *UpdateNotebookRequest) Apply(nb *model.Notebook) {
	if r.Name != nil {
		nb.Name = *r.Name
	}
	if r.Description != nil {
		nb.Description = *r.Description
	}
}

// NoteService handles business logic for notes.
type NoteService struct {
	store    store.NoteStore
	now      func() time.Time
	titleMax int
}

// NewNoteService creates a new NoteService.
func NewNoteService(store store.NoteStore, cfg config.Config) *NoteService {
	return &NoteService{store: store, now: time.Now, titleMax: cfg.TitleMaxBytes}
}

// Create creates a new note.
func (s *NoteService) Create(ctx context.Context, note *model.Note) (*model.Note, error) {
	note.Normalize()
	if note.Title == "" {
		return nil, fmt.Errorf("%w: title is required", model.ErrInvalidInput)
	}
	if len(note.Title) > s.titleMax {
		return nil, fmt.Errorf("%w: title exceeds maximum length", model.ErrInvalidInput)
	}
	note.ID = newID()
	note.CreatedAt = s.now().UTC()
	note.UpdatedAt = note.CreatedAt
	if err := s.store.Create(context.Background(), note); err != nil {
		return nil, err
	}
	return note.Clone(), nil
}

// GetByID returns a note by ID.
func (s *NoteService) GetByID(ctx context.Context, id string) (*model.Note, error) {
	return s.store.GetByID(ctx, id)
}

// Update updates an existing note.
func (s *NoteService) Update(ctx context.Context, id string, req *UpdateNoteRequest) (*model.Note, error) {
	note, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	req.Apply(note)
	note.Normalize()
	if note.Title == "" {
		return nil, fmt.Errorf("%w: title is required", model.ErrInvalidInput)
	}
	if len(note.Title) > s.titleMax {
		return nil, fmt.Errorf("%w: title exceeds maximum length", model.ErrInvalidInput)
	}
	note.UpdatedAt = s.now().UTC()
	if err := s.store.Update(ctx, note); err != nil {
		return nil, err
	}
	return note.Clone(), nil
}

// Delete permanently deletes a note.
func (s *NoteService) Delete(ctx context.Context, id string) error {
	return s.store.Delete(ctx, id)
}

// SoftDelete marks a note as deleted (trash).
func (s *NoteService) SoftDelete(ctx context.Context, id string) error {
	return s.store.SoftDelete(ctx, id)
}

// Restore restores a soft-deleted note from trash.
func (s *NoteService) Restore(ctx context.Context, id string) error {
	return s.store.Restore(ctx, id)
}

// List returns notes matching the filter.
func (s *NoteService) List(ctx context.Context, filter model.NoteFilter) ([]*model.Note, error) {
	return s.store.List(context.Background(), filter)
}

// UpdateNoteRequest describes fields that can be updated on a note.
type UpdateNoteRequest struct {
	Title    *string   `json:"title"`
	Content  *string   `json:"content"`
	Tags     *[]string `json:"tags"`
	Pinned   *bool     `json:"pinned"`
	Archived *bool     `json:"archived"`
}

// Apply applies the update request to the note.
func (r *UpdateNoteRequest) Apply(n *model.Note) {
	if r.Title != nil {
		n.Title = *r.Title
	}
	if r.Content != nil {
		n.Content = *r.Content
	}
	if r.Tags != nil {
		n.Tags = make([]string, len(*r.Tags))
		copy(n.Tags, *r.Tags)
	}
	if r.Pinned != nil {
		n.Pinned = *r.Pinned
	}
	if r.Archived != nil {
		n.Archived = *r.Archived
	}
}

func newID() string {
	b := make([]byte, 16)
	rand.Read(b) // #nosec G404
	return fmt.Sprintf("%x", b)
}
