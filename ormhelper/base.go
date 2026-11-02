package ormhelper

import "context"

// BaseRepository defines shared creation and query methods for an entity type.
type BaseRepository[TEntity any] interface {
	CreateOne(context.Context, map[string]any) (TEntity, error)
	CreateMany(context.Context, []map[string]any) ([]TEntity, error)
	FindOne(context.Context, Statement) (TEntity, error)
	FindMany(context.Context, Statement) ([]TEntity, error)
	UpdateOne(context.Context, Statement, map[string]any) (int64, error)
	UpdateMany(context.Context, Statement, map[string]any) (int64, error)
	DeleteOne(context.Context, Statement) (int64, error)
}
