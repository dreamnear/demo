package course

import (
	"database/sql"
	"fmt"
)

// Store provides database operations for courses.
type Store struct {
	db     *sql.DB
	driver string
}

// Open creates a new Store by opening a database connection.
// Supported driverName values: "sqlite", "mysql", "postgres".
func Open(driverName, dsn string) (*Store, error) {
	var actualDriver string
	switch driverName {
	case "sqlite":
		actualDriver = "sqlite"
	case "mysql":
		actualDriver = "mysql"
	case "postgres":
		actualDriver = "pgx"
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driverName)
	}

	db, err := sql.Open(actualDriver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Store{db: db, driver: driverName}, nil
}

// InitDB creates the courses table if it does not exist.
// The DDL is compatible with SQLite, MySQL, and PostgreSQL.
func (s *Store) InitDB() error {
	var ddl string
	switch s.driver {
	case "sqlite":
		ddl = `CREATE TABLE IF NOT EXISTS courses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			credits INTEGER NOT NULL
		)`
	case "mysql":
		ddl = `CREATE TABLE IF NOT EXISTS courses (
			id INT AUTO_INCREMENT PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			description TEXT NOT NULL,
			credits INT NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`
	case "postgres":
		ddl = `CREATE TABLE IF NOT EXISTS courses (
			id SERIAL PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			description TEXT NOT NULL,
			credits INT NOT NULL
		)`
	default:
		return fmt.Errorf("unsupported driver for DDL: %s", s.driver)
	}

	_, err := s.db.Exec(ddl)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}

// Create inserts a new course into the database and backfills c.ID.
func (s *Store) Create(c *Course) error {
	var err error
	switch s.driver {
	case "postgres":
		err = s.db.QueryRow(
			`INSERT INTO courses (title, description, credits) VALUES ($1, $2, $3) RETURNING id`,
			c.Title, c.Description, c.Credits,
		).Scan(&c.ID)
	default:
		// sqlite and mysql support LastInsertId
		result, execErr := s.db.Exec(
			`INSERT INTO courses (title, description, credits) VALUES (?, ?, ?)`,
			c.Title, c.Description, c.Credits,
		)
		if execErr != nil {
			return fmt.Errorf("failed to insert course: %w", execErr)
		}
		id, idErr := result.LastInsertId()
		if idErr != nil {
			return fmt.Errorf("failed to get last insert id: %w", idErr)
		}
		c.ID = int(id)
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to insert course: %w", err)
	}
	return nil
}

// GetByID retrieves a course by its ID.
// Returns sql.ErrNoRows if the course is not found.
func (s *Store) GetByID(id int) (*Course, error) {
	c := &Course{}
	var row *sql.Row
	switch s.driver {
	case "postgres":
		row = s.db.QueryRow(`SELECT id, title, description, credits FROM courses WHERE id = $1`, id)
	default:
		row = s.db.QueryRow(`SELECT id, title, description, credits FROM courses WHERE id = ?`, id)
	}
	err := row.Scan(&c.ID, &c.Title, &c.Description, &c.Credits)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// List returns all courses from the database.
func (s *Store) List() ([]Course, error) {
	rows, err := s.db.Query(`SELECT id, title, description, credits FROM courses`)
	if err != nil {
		return nil, fmt.Errorf("failed to query courses: %w", err)
	}
	defer rows.Close()

	var courses []Course
	for rows.Next() {
		var c Course
		if err := rows.Scan(&c.ID, &c.Title, &c.Description, &c.Credits); err != nil {
			return nil, fmt.Errorf("failed to scan course: %w", err)
		}
		courses = append(courses, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating courses: %w", err)
	}
	return courses, nil
}

// Update updates an existing course's Title, Description, and Credits by ID.
func (s *Store) Update(c *Course) error {
	var err error
	switch s.driver {
	case "postgres":
		_, err = s.db.Exec(
			`UPDATE courses SET title = $1, description = $2, credits = $3 WHERE id = $4`,
			c.Title, c.Description, c.Credits, c.ID,
		)
	default:
		_, err = s.db.Exec(
			`UPDATE courses SET title = ?, description = ?, credits = ? WHERE id = ?`,
			c.Title, c.Description, c.Credits, c.ID,
		)
	}
	if err != nil {
		return fmt.Errorf("failed to update course: %w", err)
	}
	return nil
}

// Delete removes a course by its ID.
func (s *Store) Delete(id int) error {
	var err error
	switch s.driver {
	case "postgres":
		_, err = s.db.Exec(`DELETE FROM courses WHERE id = $1`, id)
	default:
		_, err = s.db.Exec(`DELETE FROM courses WHERE id = ?`, id)
	}
	if err != nil {
		return fmt.Errorf("failed to delete course: %w", err)
	}
	return nil
}

// Close closes the underlying database connection.
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
