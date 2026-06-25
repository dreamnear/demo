// Package course provides data models and database operations for courses.
package course

// Course represents a course entity in the system.
type Course struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Credits     int    `json:"credits"`
}
