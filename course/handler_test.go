package course

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// setupTestHandler creates a handler backed by an in-memory SQLite database.
func setupTestHandler(t *testing.T) *Handler {
	t.Helper()
	store := setupTestStore(t)
	return NewHandler(store)
}

// setupTestMux creates a ServeMux with all routes registered.
func setupTestMux(t *testing.T, h *Handler) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux
}

func TestListCoursesEmpty(t *testing.T) {
	h := setupTestHandler(t)
	mux := setupTestMux(t, h)

	req := httptest.NewRequest("GET", "/courses", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
	var courses []Course
	if err := json.NewDecoder(rr.Body).Decode(&courses); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if len(courses) != 0 {
		t.Errorf("expected empty list, got %d courses", len(courses))
	}
}

func TestCreateAndGetCourse(t *testing.T) {
	h := setupTestHandler(t)
	mux := setupTestMux(t, h)

	body := `{"title":"Math 101","description":"Intro to Math","credits":3}`
	req := httptest.NewRequest("POST", "/courses", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d; body: %s", rr.Code, rr.Body.String())
	}
	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	var created Course
	if err := json.NewDecoder(rr.Body).Decode(&created); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if created.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if created.Title != "Math 101" {
		t.Errorf("expected title %q, got %q", "Math 101", created.Title)
	}

	// GET by id
	req = httptest.NewRequest("GET", "/courses/1", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var got Course
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, got.ID)
	}
	if got.Title != "Math 101" {
		t.Errorf("expected title %q, got %q", "Math 101", got.Title)
	}
}

func TestGetCourseNotFound(t *testing.T) {
	h := setupTestHandler(t)
	mux := setupTestMux(t, h)

	req := httptest.NewRequest("GET", "/courses/999", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}

func TestGetCourseInvalidID(t *testing.T) {
	h := setupTestHandler(t)
	mux := setupTestMux(t, h)

	req := httptest.NewRequest("GET", "/courses/abc", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCreateCourseInvalidBody(t *testing.T) {
	h := setupTestHandler(t)
	mux := setupTestMux(t, h)

	req := httptest.NewRequest("POST", "/courses", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestUpdateCourse(t *testing.T) {
	h := setupTestHandler(t)
	mux := setupTestMux(t, h)

	// Create first
	createBody := `{"title":"Physics","description":"Intro","credits":4}`
	req := httptest.NewRequest("POST", "/courses", bytes.NewBufferString(createBody))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d", rr.Code)
	}

	// Update
	updateBody := `{"title":"Advanced Physics","description":"Advanced","credits":5}`
	req = httptest.NewRequest("PUT", "/courses/1", bytes.NewBufferString(updateBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
	var updated Course
	if err := json.NewDecoder(rr.Body).Decode(&updated); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if updated.Title != "Advanced Physics" {
		t.Errorf("expected title %q, got %q", "Advanced Physics", updated.Title)
	}
}

func TestUpdateCourseInvalidID(t *testing.T) {
	h := setupTestHandler(t)
	mux := setupTestMux(t, h)

	body := `{"title":"X","description":"Y","credits":1}`
	req := httptest.NewRequest("PUT", "/courses/abc", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestDeleteCourse(t *testing.T) {
	h := setupTestHandler(t)
	mux := setupTestMux(t, h)

	// Create first
	createBody := `{"title":"ToDelete","description":"Temp","credits":1}`
	req := httptest.NewRequest("POST", "/courses", bytes.NewBufferString(createBody))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d", rr.Code)
	}

	// Delete
	req = httptest.NewRequest("DELETE", "/courses/1", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	// Verify gone
	req = httptest.NewRequest("GET", "/courses/1", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", rr.Code)
	}
}

func TestDeleteCourseInvalidID(t *testing.T) {
	h := setupTestHandler(t)
	mux := setupTestMux(t, h)

	req := httptest.NewRequest("DELETE", "/courses/abc", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestListCoursesWithData(t *testing.T) {
	h := setupTestHandler(t)
	mux := setupTestMux(t, h)

	// Create two courses
	for _, title := range []string{"Course A", "Course B"} {
		body := `{"title":"` + title + `","description":"Desc","credits":2}`
		req := httptest.NewRequest("POST", "/courses", bytes.NewBufferString(body))
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		if rr.Code != http.StatusCreated {
			t.Fatalf("create: expected 201, got %d", rr.Code)
		}
	}

	// List
	req := httptest.NewRequest("GET", "/courses", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var courses []Course
	if err := json.NewDecoder(rr.Body).Decode(&courses); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(courses) != 2 {
		t.Errorf("expected 2 courses, got %d", len(courses))
	}
}
