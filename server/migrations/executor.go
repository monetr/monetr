package migrations

import (
	"context"

	"github.com/go-pg/pg/v10"
)

// Executor is the bit of a database driver that the migration runner actually
// uses. It's an interface so we can swap go-pg out for something else (or a
// fake in a test) without dragging the rest of the package along.
type Executor interface {
	Exec(ctx context.Context, query string, args ...any) error
	Get(ctx context.Context, dst any, query string, args ...any) error
	Query(ctx context.Context, dst any, query string, args ...any) error
	RunInTransaction(ctx context.Context, fn func(Executor) error) error
}

// pgInner is the subset of go-pg's API that pgExecutor reaches for. *pg.DB,
// *pg.Conn, and *pg.Tx all expose these via baseDB method promotion, so one
// adapter wraps all three.
type pgInner interface {
	ExecContext(ctx context.Context, query interface{}, params ...interface{}) (pg.Result, error)
	QueryOneContext(ctx context.Context, model, query interface{}, params ...interface{}) (pg.Result, error)
	QueryContext(ctx context.Context, model, query interface{}, params ...interface{}) (pg.Result, error)
	RunInTransaction(ctx context.Context, fn func(*pg.Tx) error) error
}

// Lock pgInner satisfaction in at compile time so a go-pg upgrade that drops
// one of these methods fails to build here rather than blowing up later when
// we actually try to migrate something.
var (
	_ pgInner = (*pg.DB)(nil)
	_ pgInner = (*pg.Conn)(nil)
	_ pgInner = (*pg.Tx)(nil)
)

type pgExecutor struct {
	db pgInner
}

func newPgExecutor(db pgInner) Executor {
	return &pgExecutor{db: db}
}

func (e *pgExecutor) Exec(ctx context.Context, query string, args ...any) error {
	_, err := e.db.ExecContext(ctx, query, args...)
	return err
}

func (e *pgExecutor) Get(ctx context.Context, dst any, query string, args ...any) error {
	_, err := e.db.QueryOneContext(ctx, pg.Scan(dst), query, args...)
	return err
}

func (e *pgExecutor) Query(ctx context.Context, dst any, query string, args ...any) error {
	_, err := e.db.QueryContext(ctx, dst, query, args...)
	return err
}

func (e *pgExecutor) RunInTransaction(ctx context.Context, fn func(Executor) error) error {
	return e.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		return fn(&pgExecutor{db: tx})
	})
}
