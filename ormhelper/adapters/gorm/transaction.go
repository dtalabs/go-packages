package gorm

import (
	"context"

	gormdb "gorm.io/gorm"
	"github.com/dtalabs/go-packages/ormhelper"
)

type TransactionManager struct{ db *gormdb.DB }

func NewTransactionManager(connection *gormdb.DB) *TransactionManager {
	return &TransactionManager{db: connection}
}

func (m *TransactionManager) Ping(ctx context.Context) error {
	if tx := ormhelper.TransactionFromContext(ctx); tx != nil {
		return tx.Ping(ctx)
	}
	connection, err := m.db.DB()
	if err != nil {
		return err
	}
	return connection.PingContext(ctx)
}

func (m *TransactionManager) Begin(ctx context.Context) (ormhelper.Transaction, error) {
	tx := m.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &transaction{tx: tx}, nil
}

type transaction struct {
	tx          *gormdb.DB
	afterCommit []func()
}

func (t *transaction) Context(ctx context.Context) context.Context {
	return ormhelper.WithTransaction(WithTransaction(ctx, t.tx), t)
}

func (t *transaction) Ping(ctx context.Context) error {
	return t.tx.WithContext(ctx).Exec("SELECT 1").Error
}

func (t *transaction) AfterCommit(callback func()) {
	t.afterCommit = append(t.afterCommit, callback)
}
func (t *transaction) Commit() error {
	if err := t.tx.Commit().Error; err != nil {
		return err
	}
	callbacks := t.afterCommit
	t.afterCommit = nil
	for _, callback := range callbacks {
		callback()
	}
	return nil
}
func (t *transaction) Rollback() error {
	t.afterCommit = nil
	return t.tx.Rollback().Error
}

var _ ormhelper.TransactionManager = (*TransactionManager)(nil)
