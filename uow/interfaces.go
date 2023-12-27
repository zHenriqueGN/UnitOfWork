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

type Repository func(dbtx DBTX) interface{}

type UowInterface interface {
	Register(eventName string, repository Repository)
	Unregister(eventName string)
	GetRepository(ctx context.Context, eventName string) (interface{}, error)
	Do(ctx context.Context, fn func() error) error
	Commit() error
	Rollback() error
}
