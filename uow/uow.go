package uow

import (
	"context"
	"database/sql"
)

type Repository func(tx *sql.Tx) interface{}

type UowInterface interface {
	Register(name string, repository Repository)
	Unregister(name string)
	GetRepository(ctx context.Context, name string) (interface{}, error)
	Do(ctx context.Context, fn func(uow UowInterface) error) error
	Commit() error
	Rollback() error
}

type UnitOfWork struct {
	DB           *sql.DB
	Tx           *sql.Tx
	Repositories map[string]Repository
}

func NewUnitOfWork(db *sql.DB) *UnitOfWork {
	return &UnitOfWork{
		DB:           db,
		Repositories: make(map[string]Repository),
	}
}

func (u *UnitOfWork) Register(name string, repository Repository) {
	u.Repositories[name] = repository
}

func (u *UnitOfWork) Unregister(name string) {
	delete(u.Repositories, name)
}
