package http

import (
	"encoding/json"
	"net/http"
)

// Response represents a standard API response
type Response struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, statusCode int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

// WriteSuccess writes a successful JSON response
func WriteSuccess(w http.ResponseWriter, data any) error {
	return WriteJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// WriteError writes an error JSON response
func WriteError(w http.ResponseWriter, statusCode int, message string) error {
	return WriteJSON(w, statusCode, Response{
		Success: false,
		Error:   message,
	})
}

// WriteCreated writes a 201 Created response
func WriteCreated(w http.ResponseWriter, data any) error {
	return WriteJSON(w, http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

// WriteNoContent writes a 204 No Content response
func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// WriteBadRequest writes a 400 Bad Request response
func WriteBadRequest(w http.ResponseWriter, message string) error {
	return WriteError(w, http.StatusBadRequest, message)
}

// WriteUnauthorized writes a 401 Unauthorized response
func WriteUnauthorized(w http.ResponseWriter, message string) error {
	return WriteError(w, http.StatusUnauthorized, message)
}

// WriteForbidden writes a 403 Forbidden response
func WriteForbidden(w http.ResponseWriter, message string) error {
	return WriteError(w, http.StatusForbidden, message)
}

// WriteNotFound writes a 404 Not Found response
func WriteNotFound(w http.ResponseWriter, message string) error {
	return WriteError(w, http.StatusNotFound, message)
}

// WriteInternalError writes a 500 Internal Server Error response
func WriteInternalError(w http.ResponseWriter, message string) error {
	return WriteError(w, http.StatusInternalServerError, message)
}
