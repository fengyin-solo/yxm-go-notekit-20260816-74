package httpapi

import (
	"net/http"
	"net/url"

	"github.com/example/notekit/internal/model"
	"github.com/example/notekit/internal/service"
	"github.com/example/notekit/internal/validator"
)

type NoteHandler struct {
	svc     *service.NoteService
	maxBody int64
}

func NewNoteHandler(svc *service.NoteService, maxBody int64) *NoteHandler {
	return &NoteHandler{svc: svc, maxBody: maxBody}
}

func (h *NoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req validator.CreateNoteRequest
	if err := decodeJSON(w, r, &req, h.maxBody); err != nil {
		writeError(w, err)
		return
	}
	if errs := req.Validate(); errs.HasErrors() {
		writeValidationErrors(w, errs)
		return
	}
	note := req.ToNote()
	created, err := h.svc.Create(r.Context(), note)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *NoteHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := pathSegment(r.URL.Path, 4)
	note, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, note)
}

func (h *NoteHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := pathSegment(r.URL.Path, 4)
	var req service.UpdateNoteRequest
	if err := decodeJSON(w, r, &req, h.maxBody); err != nil {
		writeError(w, err)
		return
	}
	updated, err := h.svc.Update(r.Context(), id, &req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *NoteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := pathSegment(r.URL.Path, 4)
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *NoteHandler) SoftDelete(w http.ResponseWriter, r *http.Request) {
	id := pathSegment(r.URL.Path, 4)
	if err := h.svc.SoftDelete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *NoteHandler) Restore(w http.ResponseWriter, r *http.Request) {
	id := pathSegment(r.URL.Path, 4)
	if err := h.svc.Restore(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *NoteHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := parseNoteFilter(r.URL.Query())
	notes, err := h.svc.List(r.Context(), filter)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, notes)
}

// parseNoteFilter reads query parameters into a NoteFilter.
// r.URL.Query() returns url.Values which has .Get() method.
func parseNoteFilter(q url.Values) model.NoteFilter {
	f := model.NoteFilter{}
	if v := q.Get("notebook_id"); v != "" {
		f.NotebookID = v
	}
	if v := q.Get("q"); v != "" {
		f.Query = v
	}
	if v := q.Get("tag"); v != "" {
		f.Tags = []string{v}
	}
	if v := q.Get("pinned"); v == "true" {
		b := true
		f.Pinned = &b
	} else if v == "false" {
		b := false
		f.Pinned = &b
	}
	if v := q.Get("archived"); v == "true" {
		b := true
		f.Archived = &b
	} else if v == "false" {
		b := false
		f.Archived = &b
	}
	if v := q.Get("include_deleted"); v == "true" {
		f.IncludeDeleted = true
	}
	return f
}
