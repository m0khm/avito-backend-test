package tx

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type transaction interface {
	Queryer
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type Manager struct {
	beginTx func(ctx context.Context) (transaction, error)
}

func NewManager(pool *pgxpool.Pool) *Manager {
	return &Manager{
		beginTx: func(ctx context.Context) (transaction, error) {
			return pool.BeginTx(ctx, pgx.TxOptions{})
		},
	}
}

type ctxKey string

const txKey ctxKey = "tx"

func (m *Manager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.beginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	ctx = context.WithValue(ctx, txKey, tx)
	if err := fn(ctx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func DB(ctx context.Context, fallback Queryer) Queryer {
	if tx, ok := ctx.Value(txKey).(Queryer); ok {
		return tx
	}
	return fallback
}

type Queryer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}
