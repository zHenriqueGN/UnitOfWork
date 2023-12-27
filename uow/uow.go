package uow

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrNoTransaction             = errors.New("no transaction")
	ErrTransactionAlreadyStarted = errors.New("transaction already started")
	ErrRowback                   = errors.New("error on rollback")
	ErrRepositoryNotRegistered   = errors.New("repository not registered")
)

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

func (u *UnitOfWork) Register(eventName string, repository Repository) {
	u.Repositories[eventName] = repository
}

func (u *UnitOfWork) Unregister(eventName string) {
	delete(u.Repositories, eventName)
}

func (u *UnitOfWork) GetRepository(ctx context.Context, eventName string) (interface{}, error) {
	if u.Tx == nil {
		tx, err := u.DB.BeginTx(ctx, nil)
		if err != nil {
			return nil, err
		}
		u.Tx = tx
	}
	if _, repositoryRegistered := u.Repositories[eventName]; !repositoryRegistered {
		return nil, fmt.Errorf("%w: %s", ErrRepositoryNotRegistered, eventName)
	}
	repository := u.Repositories[eventName](u.Tx)
	return repository, nil
}

func (u *UnitOfWork) Do(ctx context.Context, fn func() error) error {
	if u.Tx != nil {
		return ErrTransactionAlreadyStarted
	}

	tx, err := u.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	u.Tx = tx
	err = fn()
	if err != nil {
		errRowback := u.Rollback()
		if errRowback != nil {
			return fmt.Errorf("%w: %w, original error: %w", ErrRowback, errRowback, err)
		}
		return err
	}
	return u.Commit()
}

func (u *UnitOfWork) Commit() error {
	if u.Tx == nil {
		return ErrNoTransaction
	}
	err := u.Tx.Commit()
	if err != nil {
		errRowback := u.Rollback()
		if errRowback != nil {
			return fmt.Errorf("%w: %w, original error: %w", ErrRowback, errRowback, err)
		}
	}
	u.Tx = nil
	return nil
}

func (u *UnitOfWork) Rollback() error {
	if u.Tx == nil {
		return ErrNoTransaction
	}
	err := u.Tx.Rollback()
	if err != nil {
		return err
	}
	u.Tx = nil
	return nil
}
