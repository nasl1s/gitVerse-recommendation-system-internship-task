package db

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v4/stdlib" 
	"github.com/pressly/goose/v3"
)

func RunMigrations(dsn string) error {
	fmt.Printf("%s", dsn)
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("failed to open sql connection: %w", err)
	}
	defer conn.Close()

	if err := goose.Up(conn, "./migrations"); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}