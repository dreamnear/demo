package main

import (
	"database/sql"
	"demo/course"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// BUG-002: Tests for DB_DRIVER and DB_DSN environment variable reading and defaults.

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("DB_DRIVER", "")
	t.Setenv("DB_DSN", "")

	driver, dsn := loadConfig()
	if driver != "sqlite" {
		t.Errorf("expected default driver 'sqlite', got %q", driver)
	}
	if dsn != "file:demo.db?_pragma=journal_mode(WAL)" {
		t.Errorf("expected default dsn, got %q", dsn)
	}
}

func TestLoadConfigEnvOverride(t *testing.T) {
	t.Setenv("DB_DRIVER", "mysql")
	t.Setenv("DB_DSN", "user:pass@tcp(localhost:3306)/testdb")

	driver, dsn := loadConfig()
	if driver != "mysql" {
		t.Errorf("expected driver 'mysql', got %q", driver)
	}
	if dsn != "user:pass@tcp(localhost:3306)/testdb" {
		t.Errorf("expected custom dsn, got %q", dsn)
	}
}

func TestLoadConfigPostgresOverride(t *testing.T) {
	t.Setenv("DB_DRIVER", "postgres")
	t.Setenv("DB_DSN", "postgres://user:pass@localhost:5432/testdb?sslmode=disable")

	driver, dsn := loadConfig()
	if driver != "postgres" {
		t.Errorf("expected driver 'postgres', got %q", driver)
	}
	if dsn != "postgres://user:pass@localhost:5432/testdb?sslmode=disable" {
		t.Errorf("expected custom dsn, got %q", dsn)
	}
}

// BUG-003: Tests for database driver registration via _ imports.

func TestDatabaseDriversRegistered(t *testing.T) {
	drivers := sql.Drivers()

	expected := []string{"sqlite", "mysql", "pgx"}
	for _, d := range expected {
		if !slices.Contains(drivers, d) {
			t.Errorf("expected driver %q to be registered, available: %v", d, drivers)
		}
	}
}

// BUG-004: Tests for course.Open and InitDB flow.

func TestOpenAndInitDB(t *testing.T) {
	dir := t.TempDir()
	dsn := filepath.Join(dir, "test.db")

	store, err := course.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("course.Open failed: %v", err)
	}
	defer store.Close()

	if err := store.InitDB(); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	courses, err := store.List()
	if err != nil {
		t.Fatalf("List after InitDB failed: %v", err)
	}
	if len(courses) != 0 {
		t.Errorf("expected 0 courses after InitDB, got %d", len(courses))
	}
}

// BUG-005: Tests for Handler creation and RegisterRoutes.

func TestSetupMux_ListCourses(t *testing.T) {
	store := newTestStore(t)
	mux := setupMux(store)

	req := httptest.NewRequest("GET", "/courses", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /courses: expected 200, got %d", rr.Code)
	}

	var courses []course.Course
	if err := json.NewDecoder(rr.Body).Decode(&courses); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(courses) != 0 {
		t.Errorf("expected 0 courses, got %d", len(courses))
	}
}

func TestSetupMux_CreateAndGetCourse(t *testing.T) {
	store := newTestStore(t)
	mux := setupMux(store)

	// POST /courses
	body := `{"title":"Test Course","description":"A test","credits":3}`
	req := httptest.NewRequest("POST", "/courses", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("POST /courses: expected 201, got %d. body: %s", rr.Code, rr.Body.String())
	}

	var created course.Course
	if err := json.NewDecoder(rr.Body).Decode(&created); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}
	if created.ID == 0 {
		t.Error("expected non-zero ID after create")
	}
	if created.Title != "Test Course" {
		t.Errorf("expected title 'Test Course', got %q", created.Title)
	}

	// GET /courses/{id}
	req = httptest.NewRequest("GET", "/courses/1", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /courses/1: expected 200, got %d", rr.Code)
	}

	var got course.Course
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode get response: %v", err)
	}
	if got.Title != "Test Course" {
		t.Errorf("expected title 'Test Course', got %q", got.Title)
	}
}

// BUG-006: Tests for GET /health endpoint.

func TestHealthEndpoint(t *testing.T) {
	store := newTestStore(t)
	mux := setupMux(store)

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /health: expected 200, got %d", rr.Code)
	}
}

// newTestStore creates an in-memory SQLite store for testing.
func newTestStore(t *testing.T) *course.Store {
	t.Helper()
	dir := t.TempDir()
	dsn := filepath.Join(dir, "test.db")

	store, err := course.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("course.Open failed: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	if err := store.InitDB(); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	return store
}

