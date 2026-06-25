package course

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestStore(t *testing.T) *Store {
	t.Helper()
	store, err := Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := store.InitDB(); err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	t.Cleanup(func() {
		store.Close()
	})
	return store
}

func TestCreate(t *testing.T) {
	store := setupTestStore(t)

	c := &Course{
		Title:       "Mathematics 101",
		Description: "Introduction to Mathematics",
		Credits:     3,
	}

	if err := store.Create(c); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if c.ID == 0 {
		t.Error("expected ID to be backfilled, got 0")
	}
}

func TestGetByID(t *testing.T) {
	store := setupTestStore(t)

	// Create a course first
	c := &Course{
		Title:       "Physics 101",
		Description: "Introduction to Physics",
		Credits:     4,
	}
	if err := store.Create(c); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Retrieve by ID
	retrieved, err := store.GetByID(c.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved.ID != c.ID {
		t.Errorf("expected ID %d, got %d", c.ID, retrieved.ID)
	}
	if retrieved.Title != c.Title {
		t.Errorf("expected Title %q, got %q", c.Title, retrieved.Title)
	}
	if retrieved.Description != c.Description {
		t.Errorf("expected Description %q, got %q", c.Description, retrieved.Description)
	}
	if retrieved.Credits != c.Credits {
		t.Errorf("expected Credits %d, got %d", c.Credits, retrieved.Credits)
	}
}

func TestGetByIDNotFound(t *testing.T) {
	store := setupTestStore(t)

	_, err := store.GetByID(999)
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestList(t *testing.T) {
	store := setupTestStore(t)

	// Create multiple courses
	courses := []Course{
		{Title: "Course 1", Description: "Desc 1", Credits: 3},
		{Title: "Course 2", Description: "Desc 2", Credits: 4},
		{Title: "Course 3", Description: "Desc 3", Credits: 2},
	}

	for i := range courses {
		if err := store.Create(&courses[i]); err != nil {
			t.Fatalf("Create failed for course %d: %v", i, err)
		}
	}

	// List all courses
	list, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(list) != len(courses) {
		t.Errorf("expected %d courses, got %d", len(courses), len(list))
	}
}

func TestUpdate(t *testing.T) {
	store := setupTestStore(t)

	// Create a course
	c := &Course{
		Title:       "Original Title",
		Description: "Original Description",
		Credits:     3,
	}
	if err := store.Create(c); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Update the course
	c.Title = "Updated Title"
	c.Description = "Updated Description"
	c.Credits = 4

	if err := store.Update(c); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Retrieve and verify
	retrieved, err := store.GetByID(c.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved.Title != "Updated Title" {
		t.Errorf("expected Title %q, got %q", "Updated Title", retrieved.Title)
	}
	if retrieved.Description != "Updated Description" {
		t.Errorf("expected Description %q, got %q", "Updated Description", retrieved.Description)
	}
	if retrieved.Credits != 4 {
		t.Errorf("expected Credits %d, got %d", 4, retrieved.Credits)
	}
}

func TestDelete(t *testing.T) {
	store := setupTestStore(t)

	// Create a course
	c := &Course{
		Title:       "To Be Deleted",
		Description: "This will be deleted",
		Credits:     2,
	}
	if err := store.Create(c); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Delete the course
	if err := store.Delete(c.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's gone
	_, err := store.GetByID(c.ID)
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows after delete, got %v", err)
	}
}
