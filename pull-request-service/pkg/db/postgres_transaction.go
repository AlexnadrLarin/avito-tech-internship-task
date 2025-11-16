package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txKey struct{}



type TransactionManager struct {
	pool *pgxpool.Pool
}

func NewTransactionManager(pool *pgxpool.Pool) *TransactionManager {
	return &TransactionManager{pool: pool}
}

func (tm *TransactionManager) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	if tx := getTxFromContext(ctx); tx != nil {
		return fn(ctx)
	}

	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}

	ctxWithTx := context.WithValue(ctx, txKey{}, tx)

	err = fn(ctxWithTx)
	if err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return fmt.Errorf("rolling back transaction: %w, original error: %w", rollbackErr, err)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

type TxOrPool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
}


func GetTx(ctx context.Context, pool *pgxpool.Pool) TxOrPool {
	if tx := getTxFromContext(ctx); tx != nil {
		return tx
	}
	return pool
}

func getTxFromContext(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return nil
}
