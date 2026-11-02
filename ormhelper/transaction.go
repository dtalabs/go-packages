package ormhelper

import "context"

type TransactionManager interface {
	Begin(context.Context) (Transaction, error)
}

type Transaction interface {
	Context(context.Context) context.Context
	Ping(context.Context) error
	AfterCommit(func())
	Commit() error
	Rollback() error
}

const transactionContextKey = "ormhelper/transaction"

func WithTransaction(ctx context.Context, tx Transaction) context.Context {
	return context.WithValue(ctx, transactionContextKey, tx)
}

func TransactionFromContext(ctx context.Context) Transaction {
	tx, _ := ctx.Value(transactionContextKey).(Transaction)
	return tx
}

func RunInTransaction[Output any](ctx context.Context, manager TransactionManager, run func(context.Context) (Output, error)) (Output, error) {
	if TransactionFromContext(ctx) != nil {
		return run(ctx)
	}
	var zero Output
	tx, err := manager.Begin(ctx)
	if err != nil {
		return zero, err
	}
	defer tx.Rollback()
	output, err := run(tx.Context(ctx))
	if err != nil {
		return zero, err
	}
	if err := tx.Commit(); err != nil {
		return zero, err
	}
	return output, nil
}
