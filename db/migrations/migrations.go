package migrations

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

// Run выполняет все миграции из папки ./migrations
func Run() {
	db, err := sql.Open("postgres", os.Getenv("POSTGRES_CONN"))
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("failed to set dialect: %v", err)
	}

	migrationDir := "./migrations" // или полный путь к вашим миграциям

	fmt.Printf("Running migrations from %s\n", migrationDir)
	if err := goose.Up(db, migrationDir); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
}
