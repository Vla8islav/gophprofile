package main

import (
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("migrate: %v", err)
	}
}

func run() error {
	db, err := goose.OpenDBWithDriver("pgx", os.Getenv("DATABASE_URI"))
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	// New pods race NetworkPolicy ipset sync (and, on a fresh cluster,
	// postgres init) — wait for the DB rather than dying on first refusal
	deadline := time.Now().Add(60 * time.Second)
	for {
		err = db.Ping()
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("database not reachable after 60s: %w", err)
		}
		log.Printf("waiting for database: %v", err)
		time.Sleep(2 * time.Second)
	}

	return goose.Up(db, os.Getenv("MIGRATIONS_FOLDER"))
}
