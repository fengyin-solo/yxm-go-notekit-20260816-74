package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/example/notekit/internal/logger"
	"github.com/example/notekit/internal/model"
)

type MemoryNoteStore struct {
	mu       sync.RWMutex
	items    map[string]*model.Note
	notebook map[string]int // notebookID -> noteCount
	path     string
	dirty    bool
	log      *logger.Logger
	stop     chan struct{}
	done     chan struct{}
}

func NewMemoryNoteStore(path string, log *logger.Logger, interval time.Duration) (*MemoryNoteStore, error) {
	if log == nil {
		log = logger.Default()
	}
	s := &MemoryNoteStore{
		items:    make(map[string]*model.Note),
		notebook: make(map[string]int),
		path:     path,
		log:      log,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
	if path != "" {
		if err := s.load(); err != nil {
			return nil, fmt.Errorf("load notes: %w", err)
		}
		go s.saveLoop(interval)
	} else {
		close(s.done)
	}
	return s, nil
}

func (s *MemoryNoteStore) Create(ctx context.Context, note *model.Note) error {
	if note == nil || note.ID == "" || note.NotebookID == "" {
		return model.ErrInvalidInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.items[note.ID]; exists {
		return model.ErrAlreadyExists
	}
	s.items[note.ID] = note.Clone()
	if !note.Deleted {
		s.notebook[note.NotebookID]++
	}
	s.dirty = true
	return nil
}

func (s *MemoryNoteStore) GetByID(ctx context.Context, id string) (*model.Note, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n, ok := s.items[id]
	if !ok {
		return nil, model.ErrNotFound
	}
	return n.Clone(), nil
}

func (s *MemoryNoteStore) Update(ctx context.Context, note *model.Note) error {
	if note == nil || note.ID == "" {
		return model.ErrInvalidInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	old, ok := s.items[note.ID]
	if !ok {
		return model.ErrNotFound
	}
	s.items[note.ID] = note.Clone()
	// Update notebook count if deleted flag changed
	if old.Deleted && !note.Deleted {
		s.notebook[note.NotebookID]++
	} else if !old.Deleted && note.Deleted {
		s.notebook[note.NotebookID]--
	}
	s.dirty = true
	return nil
}

func (s *MemoryNoteStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	n, ok := s.items[id]
	if !ok {
		return model.ErrNotFound
	}
	if !n.Deleted {
		s.notebook[n.NotebookID]--
	}
	delete(s.items, id)
	s.dirty = true
	return nil
}

func (s *MemoryNoteStore) SoftDelete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	n, ok := s.items[id]
	if !ok {
		return model.ErrNotFound
	}
	if !n.Deleted {
		n.Deleted = true
		n.UpdatedAt = time.Now().UTC()
		s.notebook[n.NotebookID]--
		s.dirty = true
	}
	return nil
}

func (s *MemoryNoteStore) Restore(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	n, ok := s.items[id]
	if !ok {
		return model.ErrNotFound
	}
	if n.Deleted {
		n.Deleted = false
		n.UpdatedAt = time.Now().UTC()
		s.notebook[n.NotebookID]++
		s.dirty = true
	}
	return nil
}

func (s *MemoryNoteStore) List(ctx context.Context, filter model.NoteFilter) ([]*model.Note, error) {
	s.mu.RLock()
	matched := make([]*model.Note, 0, len(s.items))
	for _, n := range s.items {
		if filter.Matches(n) {
			matched = append(matched, n.Clone())
		}
	}
	s.mu.RUnlock()
	sort.Slice(matched, func(i, j int) bool {
		if matched[i].Pinned != matched[j].Pinned {
			return matched[i].Pinned
		}
		if matched[i].UpdatedAt.Equal(matched[j].UpdatedAt) {
			return matched[i].ID < matched[j].ID
		}
		return matched[i].UpdatedAt.After(matched[j].UpdatedAt)
	})
	return matched, nil
}

func (s *MemoryNoteStore) Count(ctx context.Context, filter model.NoteFilter) (int, error) {
	s.mu.RLock()
	n := 0
	for _, note := range s.items {
		if filter.Matches(note) {
			n++
		}
	}
	s.mu.RUnlock()
	return n, nil
}

func (s *MemoryNoteStore) Close() error {
	if s.path == "" {
		return nil
	}
	s.mu.Lock()
	close(s.stop)
	s.mu.Unlock()
	<-s.done
	return s.flush()
}

func (s *MemoryNoteStore) saveLoop(interval time.Duration) {
	defer close(s.done)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			if err := s.flushIfDirty(); err != nil {
				s.log.Error("note store flush failed", "err", err.Error())
			}
		}
	}
}

func (s *MemoryNoteStore) flushIfDirty() error {
	s.mu.RLock()
	dirty := s.dirty
	s.mu.RUnlock()
	if !dirty {
		return nil
	}
	return s.flush()
}

func (s *MemoryNoteStore) flush() error {
	if s.path == "" {
		return nil
	}
	s.mu.Lock()
	snapshot := make([]*model.Note, 0, len(s.items))
	for _, n := range s.items {
		snapshot = append(snapshot, n.Clone())
	}
	s.dirty = false
	s.mu.Unlock()
	sort.Slice(snapshot, func(i, j int) bool { return snapshot[i].ID < snapshot[j].ID })
	payload, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal notes: %w", err)
	}
	dir := filepath.Dir(s.path)
	if dir != "" && dir != "." {
		os.MkdirAll(dir, 0o755) // #nosec G301
	}
	tmp, err := os.CreateTemp(dir, ".notekit-note-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(payload); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		return err
	}
	return nil
}

func (s *MemoryNoteStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if len(data) == 0 {
		return nil
	}
	var items []*model.Note
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("corrupt note data: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, n := range items {
		if n == nil || n.ID == "" {
			continue
		}
		s.items[n.ID] = n
		if !n.Deleted {
			s.notebook[n.NotebookID]++
		}
	}
	return nil
}
