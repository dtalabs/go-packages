package gorm

// Model converts a persistence model to its repository entity.
type Model[TEntity any] interface {
	ToEntity() TEntity
}
