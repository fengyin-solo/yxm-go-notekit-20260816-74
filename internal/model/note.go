// Package model defines the domain entities for the notekit note-taking service.
package model

import (
	"strings"
	"time"
)

// Notebook groups notes under a common namespace.
type Notebook struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	NoteCount   int       `json:"note_count"`
}

// Clone returns a deep copy of the notebook.
func (n *Notebook) Clone() *Notebook {
	if n == nil {
		return nil
	}
	cp := *n
	return &cp
}

// Normalize trims whitespace from name and description.
func (n *Notebook) Normalize() {
	n.Name = strings.TrimSpace(n.Name)
	n.Description = strings.TrimSpace(n.Description)
}

// Note represents a single note within a notebook.
type Note struct {
	ID         string    `json:"id"`
	NotebookID string    `json:"notebook_id"`
	Title      string    `json:"title"`
	Content    string    `json:"content,omitempty"`
	Tags       []string  `json:"tags,omitempty"`
	Pinned     bool      `json:"pinned"`
	Archived   bool      `json:"archived"`
	Deleted    bool      `json:"deleted"` // soft delete flag
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Clone returns a deep copy of the note, including a fresh copy of the tags slice.
func (n *Note) Clone() *Note {
	if n == nil {
		return nil
	}
	cp := *n
	return &cp
}

// Normalize trims whitespace from title, content, and tags.
func (n *Note) Normalize() {
	n.Title = strings.TrimSpace(n.Title)
	n.Content = strings.TrimSpace(n.Content)
	clean := make([]string, 0, len(n.Tags))
	for _, t := range n.Tags {
		t = strings.TrimSpace(t)
		if t != "" {
			clean = append(clean, strings.ToLower(t))
		}
	}
	n.Tags = clean
}

// NoteFilter constrains note listing queries.
type NoteFilter struct {
	NotebookID     string
	Query          string // full-text match against title+content
	Tags           []string
	Pinned         *bool
	Archived       *bool
	IncludeDeleted bool
}

// Matches reports whether the note satisfies every active filter constraint.
func (f *NoteFilter) Matches(n *Note) bool {
	if f.NotebookID != "" && n.NotebookID != f.NotebookID {
		return false
	}
	if !f.IncludeDeleted && n.Deleted {
		return false
	}
	if f.Archived != nil && n.Archived != *f.Archived {
		return false
	}
	if f.Pinned != nil && n.Pinned != *f.Pinned {
		return false
	}
	if len(f.Tags) > 0 {
		tagSet := make(map[string]bool, len(n.Tags))
		for _, t := range n.Tags {
			tagSet[t] = true
		}
		hasAll := true
		for _, t := range f.Tags {
			if !tagSet[t] {
				hasAll = false
				break
			}
		}
		if !hasAll {
			return false
		}
	}
	if f.Query != "" {
		q := strings.ToLower(strings.TrimSpace(f.Query))
		if q != "" {
			lower := strings.ToLower(n.Title + " " + n.Content)
			if !strings.Contains(lower, q) {
				return false
			}
		}
	}
	return true
}
