package gorm

type Model[TEntity any] interface {
	ToEntity() TEntity
}
