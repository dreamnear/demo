package course

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
)

// Handler provides HTTP handlers for course CRUD operations.
type Handler struct {
	Store *Store
}

// NewHandler creates a new Handler with the given Store.
func NewHandler(store *Store) *Handler {
	return &Handler{Store: store}
}

// RegisterRoutes registers all course routes on the given ServeMux
// using Go 1.22 method-aware path patterns.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /courses", h.listCourses)
	mux.HandleFunc("POST /courses", h.createCourse)
	mux.HandleFunc("GET /courses/{id}", h.getCourse)
	mux.HandleFunc("PUT /courses/{id}", h.updateCourse)
	mux.HandleFunc("DELETE /courses/{id}", h.deleteCourse)
}

// writeJSON marshals v as JSON and writes it to w with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		json.NewEncoder(w).Encode(v)
	}
}

// writeError writes a JSON error response with the given status code and message.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// parseID extracts and parses the "id" path parameter as an integer.
// Returns the parsed id and true on success, or writes a 400 error and returns false.
func (h *Handler) parseID(w http.ResponseWriter, r *http.Request) (int, bool) {
	raw := r.PathValue("id")
	id, err := strconv.Atoi(raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id parameter")
		return 0, false
	}
	return id, true
}

// listCourses handles GET /courses.
func (h *Handler) listCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := h.Store.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list courses")
		return
	}
	if courses == nil {
		courses = []Course{}
	}
	writeJSON(w, http.StatusOK, courses)
}

// createCourse handles POST /courses.
func (h *Handler) createCourse(w http.ResponseWriter, r *http.Request) {
	var c Course
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.Store.Create(&c); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create course")
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

// getCourse handles GET /courses/{id}.
func (h *Handler) getCourse(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parseID(w, r)
	if !ok {
		return
	}
	c, err := h.Store.GetByID(id)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "course not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get course")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

// updateCourse handles PUT /courses/{id}.
func (h *Handler) updateCourse(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parseID(w, r)
	if !ok {
		return
	}
	var c Course
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	c.ID = id
	if err := h.Store.Update(&c); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update course")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

// deleteCourse handles DELETE /courses/{id}.
func (h *Handler) deleteCourse(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parseID(w, r)
	if !ok {
		return
	}
	if err := h.Store.Delete(id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete course")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
