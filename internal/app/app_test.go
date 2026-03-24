package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

type fakeMigrationTx struct {
	executed    []string
	commitErr   error
	rollbackCnt int
	committed   bool
	inserted    map[string]bool
}

func (f *fakeMigrationTx) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	trimmed := strings.TrimSpace(sql)
	f.executed = append(f.executed, trimmed)
	if strings.Contains(trimmed, "INSERT INTO schema_migrations") {
		filename := arguments[0].(string)
		if f.inserted[filename] {
			return pgconn.NewCommandTag("INSERT 0 0"), nil
		}
		f.inserted[filename] = true
		return pgconn.NewCommandTag("INSERT 0 1"), nil
	}
	return pgconn.NewCommandTag("EXEC 1"), nil
}

func (f *fakeMigrationTx) Commit(ctx context.Context) error {
	f.committed = true
	return f.commitErr
}

func (f *fakeMigrationTx) Rollback(ctx context.Context) error {
	f.rollbackCnt++
	return nil
}

type fakeMigrationDB struct {
	tx       *fakeMigrationTx
	execSQL  []string
	beginErr error
}

func (f *fakeMigrationDB) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	f.execSQL = append(f.execSQL, strings.TrimSpace(sql))
	return pgconn.NewCommandTag("EXEC 1"), nil
}

func (f *fakeMigrationDB) BeginTx(ctx context.Context) (migrationTx, error) {
	if f.beginErr != nil {
		return nil, f.beginErr
	}
	return f.tx, nil
}

func TestMigrationFiles_SortsFiles(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"0002_b.up.sql", "0001_a.up.sql"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("SELECT 1;"), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}

	files, err := migrationFiles(dir)
	if err != nil {
		t.Fatalf("migrationFiles() error = %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	if !strings.HasSuffix(files[0], "0001_a.up.sql") || !strings.HasSuffix(files[1], "0002_b.up.sql") {
		t.Fatalf("files are not sorted: %v", files)
	}
}

func TestApplyMigrationsFromDir_AppliesOnlyOnce(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "0001_init.up.sql"), []byte("CREATE TABLE test(id int);"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	db := &fakeMigrationDB{tx: &fakeMigrationTx{inserted: map[string]bool{}}}
	if err := applyMigrationsFromDir(context.Background(), db, dir); err != nil {
		t.Fatalf("first applyMigrationsFromDir() error = %v", err)
	}
	if err := applyMigrationsFromDir(context.Background(), db, dir); err != nil {
		t.Fatalf("second applyMigrationsFromDir() error = %v", err)
	}

	var appliedSQL int
	for _, sql := range db.tx.executed {
		if strings.Contains(sql, "CREATE TABLE test") {
			appliedSQL++
		}
	}
	if appliedSQL != 1 {
		t.Fatalf("expected migration body to run once, got %d", appliedSQL)
	}
}

func TestApplyMigrationsFromDir_BeginError(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "0001_init.up.sql"), []byte("SELECT 1;"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := applyMigrationsFromDir(context.Background(), &fakeMigrationDB{beginErr: errors.New("no tx")}, dir)
	if err == nil {
		t.Fatal("expected error")
	}
}
