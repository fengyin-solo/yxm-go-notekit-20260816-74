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

type MemoryNotebookStore struct {
	mu    sync.RWMutex
	items map[string]*model.Notebook
	path  string
	dirty bool
	log   *logger.Logger
	stop  chan struct{}
	done  chan struct{}
}

func NewMemoryNotebookStore(path string, log *logger.Logger, interval time.Duration) (*MemoryNotebookStore, error) {
	if log == nil {
		log = logger.Default()
	}
	s := &MemoryNotebookStore{
		items: make(map[string]*model.Notebook),
		path:  path,
		log:   log,
		stop:  make(chan struct{}),
		done:  make(chan struct{}),
	}
	if path != "" {
		if err := s.load(); err != nil {
			return nil, fmt.Errorf("load notebooks: %w", err)
		}
		go s.saveLoop(interval)
	} else {
		close(s.done)
	}
	return s, nil
}

func (s *MemoryNotebookStore) Create(ctx context.Context, nb *model.Notebook) error {
	if nb == nil || nb.ID == "" {
		return model.ErrInvalidInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.items[nb.ID]; exists {
		return model.ErrAlreadyExists
	}
	s.items[nb.ID] = nb.Clone()
	s.dirty = true
	return nil
}

func (s *MemoryNotebookStore) GetByID(ctx context.Context, id string) (*model.Notebook, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	nb, ok := s.items[id]
	if !ok {
		return nil, model.ErrNotFound
	}
	return nb.Clone(), nil
}

func (s *MemoryNotebookStore) Update(ctx context.Context, nb *model.Notebook) error {
	if nb == nil || nb.ID == "" {
		return model.ErrInvalidInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[nb.ID]; !ok {
		return model.ErrNotFound
	}
	s.items[nb.ID] = nb.Clone()
	s.dirty = true
	return nil
}

func (s *MemoryNotebookStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return model.ErrNotFound
	}
	delete(s.items, id)
	s.dirty = true
	return nil
}

func (s *MemoryNotebookStore) List(ctx context.Context) ([]*model.Notebook, error) {
	s.mu.RLock()
	out := make([]*model.Notebook, 0, len(s.items))
	for _, nb := range s.items {
		out = append(out, nb.Clone())
	}
	s.mu.RUnlock()
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func (s *MemoryNotebookStore) Close() error {
	if s.path == "" {
		return nil
	}
	s.mu.Lock()
	close(s.stop)
	s.mu.Unlock()
	<-s.done
	return s.flush()
}

func (s *MemoryNotebookStore) saveLoop(interval time.Duration) {
	defer close(s.done)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			if err := s.flushIfDirty(); err != nil {
				s.log.Error("notebook store flush failed", "err", err.Error())
			}
		}
	}
}

func (s *MemoryNotebookStore) flushIfDirty() error {
	s.mu.RLock()
	dirty := s.dirty
	s.mu.RUnlock()
	if !dirty {
		return nil
	}
	return s.flush()
}

func (s *MemoryNotebookStore) flush() error {
	if s.path == "" {
		return nil
	}
	s.mu.Lock()
	snapshot := make([]*model.Notebook, 0, len(s.items))
	for _, nb := range s.items {
		snapshot = append(snapshot, nb.Clone())
	}
	s.dirty = false
	s.mu.Unlock()
	sort.Slice(snapshot, func(i, j int) bool { return snapshot[i].ID < snapshot[j].ID })
	payload, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal notebooks: %w", err)
	}
	dir := filepath.Dir(s.path)
	if dir != "" && dir != "." {
		os.MkdirAll(dir, 0o755) // #nosec G301
	}
	tmp, err := os.CreateTemp(dir, ".notekit-nb-*.tmp")
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

func (s *MemoryNotebookStore) load() error {
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
	var items []*model.Notebook
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("corrupt notebook data: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, nb := range items {
		if nb == nil || nb.ID == "" {
			continue
		}
		s.items[nb.ID] = nb
	}
	return nil
}
