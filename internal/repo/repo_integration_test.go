package repo

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func testDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5433/rooms?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("pgxpool.New() error = %v", err)
	}

	resetTestDB(t, pool)
	applyTestMigrations(t, pool)

	t.Cleanup(func() {
		pool.Close()
	})

	return pool
}

func resetTestDB(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	_, err := pool.Exec(context.Background(), `
		DROP SCHEMA public CASCADE;
		CREATE SCHEMA public;
	`)
	if err != nil {
		t.Fatalf("reset test db error = %v", err)
	}
}

func applyTestMigrations(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	patterns := []string{
		filepath.Join("db", "migrations", "*.up.sql"),
		filepath.Join("..", "..", "db", "migrations", "*.up.sql"),
	}

	var files []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatalf("glob migrations error = %v", err)
		}
		if len(matches) > 0 {
			files = matches
			break
		}
	}

	if len(files) == 0 {
		t.Fatalf("no migration files found")
	}

	sort.Strings(files)

	for _, file := range files {
		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read migration %s error = %v", file, err)
		}

		sql := strings.TrimSpace(string(sqlBytes))
		if sql == "" {
			continue
		}

		if _, err := pool.Exec(context.Background(), sql); err != nil {
			t.Fatalf("apply migration %s error = %v", file, err)
		}
	}
}