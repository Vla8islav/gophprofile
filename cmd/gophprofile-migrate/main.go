package main

import (
	"fmt"
	"log"
	"os"

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
	return goose.Up(db, os.Getenv("MIGRATIONS_FOLDER"))
}
