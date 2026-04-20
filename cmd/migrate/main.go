package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/asg017/sqlite-vec-go-bindings/ncruces"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

func main() {
	dbPath := "data/radio.db"
	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatalf("Failed to create data dir: %v", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open db: %v", err)
	}
	defer db.Close()

	// 1. Create migrations table
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS _migrations (name TEXT PRIMARY KEY)")
	if err != nil {
		log.Fatalf("Failed to create migrations table: %v", err)
	}

	// 2. Read migration files
	files, err := filepath.Glob("migrations/*.sql")
	if err != nil {
		log.Fatalf("Failed to list migrations: %v", err)
	}
	sort.Strings(files)

	// 3. Execute new migrations
	for _, file := range files {
		name := filepath.Base(file)
		
		var exists int
		err := db.QueryRow("SELECT COUNT(*) FROM _migrations WHERE name = ?", name).Scan(&exists)
		if err != nil {
			log.Fatalf("Query failed: %v", err)
		}
		
		if exists > 0 {
			log.Printf("[SKIP] %s", name)
			continue
		}

		log.Printf("[RUN]  %s", name)
		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Read failed: %v", err)
		}

		// Split by semicolon and run (simple parser)
		queries := strings.Split(string(content), ";")
		for _, q := range queries {
			q = strings.TrimSpace(q)
			if q == "" {
				continue
			}
			if _, err := db.Exec(q); err != nil {
				log.Fatalf("Execution failed for %s: %v\nQuery: %s", name, err, q)
			}
		}

		_, err = db.Exec("INSERT INTO _migrations (name) VALUES (?)", name)
		if err != nil {
			log.Fatalf("Failed to record migration: %v", err)
		}
	}

	log.Println("Migrations complete.")
}
