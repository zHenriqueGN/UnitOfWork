package uow

import (
	"context"
	"database/sql"
)

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type Repository func(tx *sql.Tx) interface{}

type UowInterface interface {
	Register(name string, repository Repository)
	Unregister(name string)
	GetRepository(ctx context.Context, name string) (interface{}, error)
	Do(ctx context.Context, fn func() error) error
	Commit() error
	Rollback() error
}
