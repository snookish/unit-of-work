// Package uow provides a Unit of Work implementation for managing
// database transactions across multiple repositories.
package uow

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrNoTransaction     = errors.New("uow: no active transaction; call Begin first")
	ErrExistsTransaction = errors.New("uow: transaction already active; nested transactions are not supported")
)

type TxFunc func(ctx context.Context, unit *UnitOfWork) error

type UnitOfWork struct {
	db *sql.DB
	tx *sql.Tx
}

func NewUnitOfWork(db *sql.DB) *UnitOfWork {
	return &UnitOfWork{
		db: db,
	}
}

func (uow *UnitOfWork) Begin(ctx context.Context, options *sql.TxOptions) error {
	if uow.tx != nil {
		return ErrExistsTransaction
	}

	if options == nil {
		options = &sql.TxOptions{}
	}

	tx, err := uow.db.BeginTx(ctx, options)
	if err != nil {
		return fmt.Errorf("uow: begin transaction: %w", err)
	}

	uow.tx = tx
	return nil
}

func (uow *UnitOfWork) Commit() error {
	if uow.tx == nil {
		return ErrNoTransaction
	}

	if err := uow.tx.Commit(); err != nil {
		return fmt.Errorf("uow: commit: %w", err)
	}
	return nil
}

func (uow *UnitOfWork) Rollback() error {
	if uow.tx == nil {
		return ErrNoTransaction
	}

	err := uow.tx.Rollback()
	uow.tx = nil

	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		return fmt.Errorf("uow: rollback: %w", err)
	}
	return nil
}

func (uow *UnitOfWork) Tx() (*sql.Tx, error) {
	if uow.tx == nil {
		return nil, ErrNoTransaction
	}
	return uow.tx, nil
}

func (uow *UnitOfWork) DB() *sql.DB {
	return uow.db
}

func (uow *UnitOfWork) WithTx(ctx context.Context, options *sql.TxOptions, fn TxFunc) (err error) {
	if err := uow.Begin(ctx, options); err != nil {
		return err
	}

	defer func() {
		if rErr := uow.Rollback(); rErr != nil {
			if err == nil {
				err = rErr
			}
		}
	}()

	if err := fn(ctx, uow); err != nil {
		return err
	}
	return uow.Commit()
}
