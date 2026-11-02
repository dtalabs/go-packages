package gorm

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	gormdb "gorm.io/gorm"
	"gorm.io/gorm/clause"
	"github.com/dtalabs/go-packages/ormhelper"
)

type Repository[TEntity any] struct {
	db         *gormdb.DB
	newModel   func() interface{}
	newModels  func() interface{}
	toEntity   func(interface{}) TEntity
	toEntities func(interface{}) []TEntity
}

func NewRepository[TEntity any, TModel Model[TEntity]](connection *gormdb.DB) Repository[TEntity] {
	return Repository[TEntity]{
		db:        connection,
		newModel:  func() interface{} { return new(TModel) },
		newModels: func() interface{} { return new([]TModel) },
		toEntity:  func(model interface{}) TEntity { return (*model.(*TModel)).ToEntity() },
		toEntities: func(models interface{}) []TEntity {
			rows := *models.(*[]TModel)
			entities := make([]TEntity, 0, len(rows))
			for _, row := range rows {
				entities = append(entities, row.ToEntity())
			}
			return entities
		},
	}
}

func (r *Repository[TEntity]) CreateOne(ctx context.Context, input map[string]any) (TEntity, error) {
	model := r.newModel()
	connection := r.DB(ctx)
	schema := &gormdb.Statement{DB: connection}
	if err := schema.Parse(model); err != nil {
		var zero TEntity
		return zero, err
	}
	if err := setCreateFields(ctx, schema, reflect.ValueOf(model), input); err != nil {
		var zero TEntity
		return zero, err
	}
	if err := connection.Omit(clause.Associations).Create(model).Error; err != nil {
		var zero TEntity
		return zero, err
	}
	return r.toEntity(model), nil
}

func (r *Repository[TEntity]) CreateMany(ctx context.Context, input []map[string]any) ([]TEntity, error) {
	if len(input) == 0 {
		return []TEntity{}, nil
	}
	models := r.newModels()
	connection := r.DB(ctx)
	schema := &gormdb.Statement{DB: connection}
	if err := schema.Parse(models); err != nil {
		return nil, err
	}
	rows := reflect.ValueOf(models).Elem()
	rows.Set(reflect.MakeSlice(rows.Type(), len(input), len(input)))
	for i, values := range input {
		if err := setCreateFields(ctx, schema, rows.Index(i).Addr(), values); err != nil {
			return nil, fmt.Errorf("create row %d: %w", i, err)
		}
	}
	if err := connection.Omit(clause.Associations).Create(models).Error; err != nil {
		return nil, err
	}
	return r.toEntities(models), nil
}

func setCreateFields(ctx context.Context, schema *gormdb.Statement, model reflect.Value, input map[string]any) error {
	for name, value := range input {
		field := schema.Schema.LookUpField(name)
		if field == nil || field.DBName == "" {
			return fmt.Errorf("unknown create field %q", name)
		}
		if err := field.Set(ctx, model, value); err != nil {
			return fmt.Errorf("set create field %q: %w", name, err)
		}
	}
	return nil
}

func (r *Repository[TEntity]) FindOne(ctx context.Context, statement ormhelper.Statement) (TEntity, error) {
	model := r.newModel()
	err := r.Query(ctx, statement).Take(model).Error
	if err != nil {
		var zero TEntity
		if errors.Is(err, gormdb.ErrRecordNotFound) {
			err = ormhelper.ErrNotFound
		}
		return zero, err
	}
	return r.toEntity(model), nil
}

func (r *Repository[TEntity]) FindMany(ctx context.Context, statement ormhelper.Statement) ([]TEntity, error) {
	models := r.newModels()
	if err := r.Query(ctx, statement).Find(models).Error; err != nil {
		return nil, err
	}
	return r.toEntities(models), nil
}

func (r *Repository[TEntity]) UpdateOne(ctx context.Context, statement ormhelper.Statement, input map[string]any) (int64, error) {
	return r.UpdateMany(ctx, statement, input)
}

func (r *Repository[TEntity]) UpdateMany(ctx context.Context, statement ormhelper.Statement, input map[string]any) (int64, error) {
	query := r.Query(ctx, statement)
	if query.Error != nil {
		return 0, query.Error
	}
	values := make(map[string]any, len(input))
	for name, value := range input {
		field := query.Statement.Schema.LookUpField(name)
		if field == nil || field.DBName == "" {
			return 0, fmt.Errorf("unknown update field %q", name)
		}
		values[field.Name] = value
	}
	result := query.Updates(values)
	return result.RowsAffected, result.Error
}

func (r *Repository[TEntity]) DeleteOne(ctx context.Context, statement ormhelper.Statement) (int64, error) {
	result := r.Query(ctx, statement).Delete(r.newModel())
	return result.RowsAffected, result.Error
}

func (r *Repository[TEntity]) DB(ctx context.Context) *gormdb.DB {
	if tx := transactionFromContext(ctx); tx != nil {
		return tx.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}

func (r *Repository[TEntity]) Query(ctx context.Context, statement ormhelper.Statement) *gormdb.DB {
	return applyStatement(r.DB(ctx).Model(r.newModel()), statement)
}

var _ ormhelper.BaseRepository[struct{}] = (*Repository[struct{}])(nil)

func applyStatement(connection *gormdb.DB, statement ormhelper.Statement) *gormdb.DB {
	if err := connection.Statement.Parse(connection.Statement.Model); err != nil {
		connection.AddError(err)
		return connection
	}
	column := func(name string) (clause.Column, bool) {
		field := connection.Statement.Schema.LookUpField(name)
		if field == nil || field.DBName == "" {
			connection.AddError(fmt.Errorf("unknown query field %q", name))
			return clause.Column{}, false
		}
		return clause.Column{Name: field.DBName}, true
	}
	for _, condition := range statement.Conditions {
		field, ok := column(condition.Field)
		if !ok {
			return connection
		}
		var expression clause.Expression
		switch condition.Operator {
		case ormhelper.Equal:
			expression = clause.Eq{Column: field, Value: condition.Value}
		case ormhelper.NotEqual:
			expression = clause.Neq{Column: field, Value: condition.Value}
		case ormhelper.LessThan:
			expression = clause.Lt{Column: field, Value: condition.Value}
		case ormhelper.LessThanOrEqual:
			expression = clause.Lte{Column: field, Value: condition.Value}
		case ormhelper.GreaterThan:
			expression = clause.Gt{Column: field, Value: condition.Value}
		case ormhelper.GreaterThanOrEqual:
			expression = clause.Gte{Column: field, Value: condition.Value}
		case ormhelper.In, ormhelper.NotIn:
			values := reflect.ValueOf(condition.Value)
			if !values.IsValid() || (values.Kind() != reflect.Slice && values.Kind() != reflect.Array) {
				connection.AddError(fmt.Errorf("query operator %q requires a slice or array", condition.Operator))
				return connection
			}
			items := make([]interface{}, values.Len())
			for i := range items {
				items[i] = values.Index(i).Interface()
			}
			expression = clause.IN{Column: field, Values: items}
			if condition.Operator == ormhelper.NotIn {
				expression = clause.Not(expression)
			}
		default:
			connection.AddError(fmt.Errorf("unsupported query operator %q", condition.Operator))
			return connection
		}
		connection = connection.Where(expression)
	}
	for _, sort := range statement.Sort {
		field, ok := column(sort.Field)
		if !ok {
			return connection
		}
		connection = connection.Order(clause.OrderByColumn{Column: field, Desc: sort.Descending})
	}
	if statement.Limit < 0 || statement.Offset < 0 {
		connection.AddError(fmt.Errorf("query limit and offset must be non-negative"))
		return connection
	}
	if statement.Limit > 0 {
		connection = connection.Limit(statement.Limit)
	}
	if statement.Offset > 0 {
		connection = connection.Offset(statement.Offset)
	}
	return connection
}
