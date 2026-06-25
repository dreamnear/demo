package main

import (
	"demo/course"
	"log"
	"net/http"
	"os"

	// Database drivers - registered via init()
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

// loadConfig reads database configuration from environment variables
// with sensible defaults.
func loadConfig() (driver, dsn string) {
	driver = os.Getenv("DB_DRIVER")
	if driver == "" {
		driver = "sqlite"
	}

	dsn = os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "file:demo.db?_pragma=journal_mode(WAL)"
	}
	return driver, dsn
}

// setupMux creates an HTTP ServeMux with the health endpoint and course routes.
func setupMux(store *course.Store) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := course.NewHandler(store)
	handler.RegisterRoutes(mux)

	return mux
}

func main() {
	driver, dsn := loadConfig()

	store, err := course.Open(driver, dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer store.Close()

	if err := store.InitDB(); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	mux := setupMux(store)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
