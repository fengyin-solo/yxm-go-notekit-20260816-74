package validator

import (
	"github.com/example/notekit/internal/model"
)

// CreateNotebookRequest is the payload for creating a notebook.
type CreateNotebookRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// Validate checks the request and returns any errors.
func (r CreateNotebookRequest) Validate() model.ValidationErrors {
	var ve model.ValidationErrors
	if r.Name = trimSpace(r.Name); r.Name == "" {
		ve = append(ve, model.ValidationError{Field: "name", Message: "required"})
	}
	return ve
}

// ToNotebook converts the request to a notebook entity.
func (r CreateNotebookRequest) ToNotebook() *model.Notebook {
	return &model.Notebook{
		Name:        r.Name,
		Description: r.Description,
	}
}

// CreateNoteRequest is the payload for creating a note.
type CreateNoteRequest struct {
	NotebookID string   `json:"notebook_id"`
	Title      string   `json:"title"`
	Content    string   `json:"content,omitempty"`
	Tags       []string `json:"tags,omitempty"`
}

// Validate checks the request and returns any errors.
func (r CreateNoteRequest) Validate() model.ValidationErrors {
	var ve model.ValidationErrors
	if r.NotebookID = trimSpace(r.NotebookID); r.NotebookID == "" {
		ve = append(ve, model.ValidationError{Field: "notebook_id", Message: "required"})
	}
	if r.Title = trimSpace(r.Title); r.Title == "" {
		ve = append(ve, model.ValidationError{Field: "title", Message: "required"})
	}
	return ve
}

// ToNote converts the request to a note entity.
func (r CreateNoteRequest) ToNote() *model.Note {
	return &model.Note{
		NotebookID: r.NotebookID,
		Title:      r.Title,
		Content:    r.Content,
		Tags:       model.NormalizeTags(r.Tags),
	}
}

func trimSpace(s string) string {
	i := 0
	j := len(s) - 1
	for i <= j && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') {
		i++
	}
	for j >= i && (s[j] == ' ' || s[j] == '\t' || s[j] == '\n' || s[j] == '\r') {
		j--
	}
	if i > j {
		return ""
	}
	return s[i : j+1]
}
