package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

// Executor is the bit of a database driver that the migration runner actually
// uses. It's an interface so we can swap the database library out for
// something else (or a fake in a test) without dragging the rest of the
// package along.
type Executor interface {
	Exec(ctx context.Context, query string, args ...any) error
	Get(ctx context.Context, dst any, query string, args ...any) error
	Query(ctx context.Context, dst any, query string, args ...any) error
	RunInTransaction(ctx context.Context, fn func(Executor) error) error
}

// bunExecutor adapts bun.IDB (implemented by *bun.DB, bun.Conn, and bun.Tx) to
// the Executor interface. Note that queries executed with no args are passed
// to the server verbatim; bun only rewrites `?` placeholders when args are
// present, which is what lets migration SQL containing literal `?` (e.g. jsonb
// operators) run unmodified.
type bunExecutor struct {
	db bun.IDB
}

func newBunExecutor(db bun.IDB) Executor {
	return &bunExecutor{db: db}
}

func (e *bunExecutor) Exec(ctx context.Context, query string, args ...any) error {
	_, err := e.db.ExecContext(ctx, query, args...)
	return err
}

func (e *bunExecutor) Get(ctx context.Context, dst any, query string, args ...any) error {
	return e.db.NewRaw(query, args...).Scan(ctx, dst)
}

func (e *bunExecutor) Query(ctx context.Context, dst any, query string, args ...any) error {
	return e.db.NewRaw(query, args...).Scan(ctx, dst)
}

func (e *bunExecutor) RunInTransaction(ctx context.Context, fn func(Executor) error) error {
	return e.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		return fn(&bunExecutor{db: tx})
	})
}
