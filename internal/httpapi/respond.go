package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/example/notekit/internal/model"
)

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error response, mapping model errors to HTTP status codes.
func writeError(w http.ResponseWriter, err error) {
	writeJSON(w, mapError(err), apiError{Error: err.Error()})
}

// writeValidationErrors writes a 400 response with field-level validation errors.
func writeValidationErrors(w http.ResponseWriter, ve model.ValidationErrors) {
	writeJSON(w, http.StatusBadRequest, ve)
}

func mapError(err error) int {
	switch err {
	case model.ErrNotFound:
		return http.StatusNotFound
	case model.ErrAlreadyExists:
		return http.StatusConflict
	case model.ErrInvalidInput:
		return http.StatusBadRequest
	case model.ErrUnauthorized:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

// decodeJSON decodes a JSON request body into dst, respecting maxBytes.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any, maxBytes int64) error {
	if maxBytes > 0 {
		r.Body = http.MaxBytesReader(nil, r.Body, maxBytes)
	}
	return json.NewDecoder(r.Body).Decode(dst)
}

type apiError struct {
	Error string `json:"error"`
}
