package gorm

import (
	"context"
	gormdb "gorm.io/gorm"
)

const transactionContextKey = "ormhelper/gorm/transaction"

// WithTransaction carries a transaction connection to repository operations.
func WithTransaction(ctx context.Context, tx *gormdb.DB) context.Context {
	return context.WithValue(ctx, transactionContextKey, tx)
}

func transactionFromContext(ctx context.Context) *gormdb.DB {
	tx, _ := ctx.Value(transactionContextKey).(*gormdb.DB)
	return tx
}
