package tx

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type fakeTx struct {
	committed  bool
	rolledBack bool
	commitErr  error
}

func (f *fakeTx) Commit(ctx context.Context) error {
	f.committed = true
	return f.commitErr
}

func (f *fakeTx) Rollback(ctx context.Context) error {
	f.rolledBack = true
	return nil
}

func (f *fakeTx) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	return pgconn.NewCommandTag("EXEC 1"), nil
}

func (f *fakeTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return nil, nil
}

func (f *fakeTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return fakeRow{}
}

func (f *fakeTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}

type fakeRow struct{}

func (fakeRow) Scan(dest ...any) error { return nil }

type fakeQueryer struct{ id string }

func (f *fakeQueryer) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	return pgconn.NewCommandTag("EXEC 1"), nil
}
func (f *fakeQueryer) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return nil, nil
}
func (f *fakeQueryer) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return fakeRow{}
}
func (f *fakeQueryer) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}

func TestManager_RunInTx_Commits(t *testing.T) {
	tx := &fakeTx{}
	m := &Manager{beginTx: func(ctx context.Context) (transaction, error) { return tx, nil }}

	err := m.RunInTx(context.Background(), func(ctx context.Context) error {
		if DB(ctx, &fakeQueryer{id: "fallback"}) != tx {
			t.Fatal("expected tx from context")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("RunInTx() error = %v", err)
	}
	if !tx.committed {
		t.Fatal("expected commit")
	}
	if !tx.rolledBack {
		t.Fatal("expected deferred rollback")
	}
}

func TestManager_RunInTx_RollsBackOnFnError(t *testing.T) {
	tx := &fakeTx{}
	m := &Manager{beginTx: func(ctx context.Context) (transaction, error) { return tx, nil }}

	err := m.RunInTx(context.Background(), func(ctx context.Context) error {
		return errors.New("boom")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if tx.committed {
		t.Fatal("did not expect commit")
	}
	if !tx.rolledBack {
		t.Fatal("expected rollback")
	}
}

func TestManager_RunInTx_CommitError(t *testing.T) {
	tx := &fakeTx{commitErr: errors.New("commit failed")}
	m := &Manager{beginTx: func(ctx context.Context) (transaction, error) { return tx, nil }}

	err := m.RunInTx(context.Background(), func(ctx context.Context) error { return nil })
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDB_UsesFallbackWithoutTransaction(t *testing.T) {
	fallback := &fakeQueryer{id: "fallback"}
	if DB(context.Background(), fallback) != fallback {
		t.Fatal("expected fallback queryer")
	}
}

func TestNewManager(t *testing.T) {
	m := NewManager(nil)
	if m == nil {
		t.Fatal("expected manager")
	}
	if m.beginTx == nil {
		t.Fatal("expected beginTx func to be set")
	}
}
