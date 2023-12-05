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
	ErrRowback                   = "erron on rollback: %s; original error: %s"
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

func (u *UnitOfWork) Register(name string, repository Repository) {
	u.Repositories[name] = repository
}

func (u *UnitOfWork) Unregister(name string) {
	delete(u.Repositories, name)
}

func (u *UnitOfWork) GetRepository(ctx context.Context, name string) (interface{}, error) {
	if u.Tx == nil {
		tx, err := u.DB.BeginTx(ctx, nil)
		if err != nil {
			return nil, err
		}
		u.Tx = tx
	}
	repository := u.Repositories[name](u.Tx)
	return repository, nil
}

func (u *UnitOfWork) Do(ctx context.Context, fn func(uow *UnitOfWork) error) error {
	if u.Tx != nil {
		return ErrTransactionAlreadyStarted
	}

	tx, err := u.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	u.Tx = tx
	err = fn(u)
	if err != nil {
		errRowback := u.Rollback()
		if errRowback != nil {
			return errors.New(fmt.Sprintf(ErrRowback, errRowback, err))
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
			return errors.New(fmt.Sprintf(ErrRowback, errRowback, err))
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
