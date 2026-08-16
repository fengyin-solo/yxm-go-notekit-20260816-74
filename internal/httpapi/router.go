package httpapi

import (
	"net/http"
	"strings"
)

// Router dispatches HTTP requests to the appropriate handlers.
type Router struct {
	notebooks *NotebookHandler
	notes     *NoteHandler
}

func NewRouter(notebooks *NotebookHandler, notes *NoteHandler) *Router {
	return &Router{notebooks: notebooks, notes: notes}
}

func (rt *Router) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", rt.healthz)
	mux.HandleFunc("/api/v1/notebooks", rt.dispatchNotebooks)
	mux.HandleFunc("/api/v1/notebooks/", rt.dispatchNotebookItem)
	mux.HandleFunc("/api/v1/notes", rt.dispatchNotes)
	mux.HandleFunc("/api/v1/notes/", rt.dispatchNoteItem)
	return mux
}

func (rt *Router) healthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (rt *Router) dispatchNotebooks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rt.notebooks.List(w, r)
	case http.MethodPost:
		rt.notebooks.Create(w, r)
	default:
		methodNotAllowed(w)
	}
}

func (rt *Router) dispatchNotebookItem(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/api/v1/notebooks/")
	rest = strings.Trim(rest, "/")
	if rest == "" {
		rt.dispatchNotebooks(w, r)
		return
	}
	parts := strings.Split(rest, "/")
	switch {
	case len(parts) == 1:
		switch r.Method {
		case http.MethodGet:
			rt.notebooks.Get(w, r)
		case http.MethodPut:
			rt.notebooks.Update(w, r)
		case http.MethodDelete:
			rt.notebooks.Delete(w, r)
		default:
			methodNotAllowed(w)
		}
	default:
		notFound(w)
	}
}

func (rt *Router) dispatchNotes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rt.notes.List(w, r)
	case http.MethodPost:
		rt.notes.Create(w, r)
	default:
		methodNotAllowed(w)
	}
}

func (rt *Router) dispatchNoteItem(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/api/v1/notes/")
	rest = strings.Trim(rest, "/")
	if rest == "" {
		rt.dispatchNotes(w, r)
		return
	}
	parts := strings.Split(rest, "/")
	switch {
	case len(parts) == 1:
		switch r.Method {
		case http.MethodGet:
			rt.notes.Get(w, r)
		case http.MethodPut:
			rt.notes.Update(w, r)
		case http.MethodDelete:
			rt.notes.Delete(w, r)
		case http.MethodPost:
			// POST /api/v1/notes/{id} with body {"action":"soft_delete"} or {"action":"restore"}
			rt.notes.SoftDelete(w, r)
		default:
			methodNotAllowed(w)
		}
	case len(parts) == 2 && parts[1] == "restore":
		if r.Method == http.MethodPost {
			rt.notes.Restore(w, r)
		} else {
			methodNotAllowed(w)
		}
	case len(parts) == 2 && parts[1] == "trash":
		if r.Method == http.MethodPost {
			rt.notes.SoftDelete(w, r)
		} else {
			methodNotAllowed(w)
		}
	default:
		notFound(w)
	}
}

func methodNotAllowed(w http.ResponseWriter) {
	writeJSON(w, http.StatusMethodNotAllowed, apiError{Error: "method not allowed"})
}

func notFound(w http.ResponseWriter) {
	writeJSON(w, http.StatusNotFound, apiError{Error: "not found"})
}
