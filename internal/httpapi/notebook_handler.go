package httpapi

import (
	"net/http"
	"strings"

	"github.com/example/notekit/internal/model"
	"github.com/example/notekit/internal/service"
)

type NotebookHandler struct {
	svc     *service.NotebookService
	maxBody int64
}

func NewNotebookHandler(svc *service.NotebookService, maxBody int64) *NotebookHandler {
	return &NotebookHandler{svc: svc, maxBody: maxBody}
}

func (h *NotebookHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
	}
	if err := decodeJSON(w, r, &req, h.maxBody); err != nil {
		writeError(w, err)
		return
	}
	nb := &model.Notebook{Name: req.Name, Description: req.Description}
	created, err := h.svc.Create(r.Context(), nb)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *NotebookHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := pathSegment(r.URL.Path, 4)
	nb, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nb)
}

func (h *NotebookHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := pathSegment(r.URL.Path, 4)
	var req service.UpdateNotebookRequest
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

func (h *NotebookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := pathSegment(r.URL.Path, 4)
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *NotebookHandler) List(w http.ResponseWriter, r *http.Request) {
	notebooks, err := h.svc.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, notebooks)
}

func pathSegment(path string, index int) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if index < 1 || index > len(parts) {
		return ""
	}
	return parts[index-1]
}
