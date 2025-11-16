package service

import "context"

type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(context.Context) error) error
}