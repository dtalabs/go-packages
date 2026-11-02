# ORM helper

Shared query types and the `BaseRepository[TEntity]` contract, imported as
`github.com/dtalabs/go-packages/ormhelper`. The root package source has no ORM dependency.
The concrete `Repository[TEntity]` is in `github.com/dtalabs/go-packages/ormhelper/adapters/gorm`.

```go
import (
    "github.com/dtalabs/go-packages/ormhelper"
    "github.com/dtalabs/go-packages/ormhelper/adapters/gorm"
)

repo := gorm.NewRepository[UserEntity, UserModel](connection)

created, err := repo.CreateOne(ctx, map[string]any{"Name": "User"})
createdMany, err := repo.CreateMany(ctx, []map[string]any{{"Name": "First"}, {"Name": "Second"}})
statement := ormhelper.Statement{}.WhereEquals("ID", created.ID)
user, err := repo.FindOne(ctx, statement)
users, err := repo.FindMany(ctx, ormhelper.Statement{Limit: 20}.OrderBy("Name", false))
updated, err := repo.UpdateOne(ctx, statement, map[string]any{"Name": "Updated User"})
updatedMany, err := repo.UpdateMany(ctx, ormhelper.Statement{}.Where("ID", ormhelper.In, ids), map[string]any{"Name": "Updated Users"})
deleted, err := repo.DeleteOne(ctx, statement)
```

Creation and updates accept plain objects using model field names or database
column names. Unknown fields return errors. Creation fills a fresh persistence
model so GORM hooks, database defaults, generated IDs, and timestamps still run;
no input-to-model mapper is required. Persistence models implement `ToEntity() TEntity`, checked by the adapter's
`Model[TEntity]` constraint. No mapper argument is needed, and domain
contracts do not reference persistence models.

`CreateOne`, `CreateMany`, `FindOne`, and `FindMany` return mapped entities.
`CreateMany` returns entities in input order, with generated IDs and timestamps;
an empty input returns an empty slice without writing to the database. `FindOne` returns
`ormhelper.ErrNotFound` when nothing matches; `FindMany` returns an empty slice.
`UpdateOne`, `UpdateMany`, and `DeleteOne` return affected row counts. `UpdateMany`
updates all matching rows in one SQL statement. Update maps preserve
zero values, and deletion follows the model's soft-delete behavior.

Use `WhereEquals("ID", id)` for equality and `Where("Rank", ormhelper.GreaterThan, 0)`
for an explicit operator. Conditions combine with AND and support comparisons, `In`, and
`NotIn`. Sort, limit, and offset apply to queries. Invalid fields, operators,
pagination, and membership values return errors.

For specialized operations, embed the adapter and use `Query(ctx, statement)`
or `DB(ctx)`. Use `RunInTransaction` to pass a transaction context to existing
repositories. It commits on success, rolls back on error or panic, and reuses
an existing context transaction for nested calls:

```go
manager := gorm.NewTransactionManager(connection)
updated, err := ormhelper.RunInTransaction(ctx, manager, func(txCtx context.Context) (int64, error) {
    return repo.UpdateOne(txCtx, statement, map[string]any{"Name": "Updated User"})
})
```

Passing a transaction's `*gorm.DB` to a repository constructor is also supported.
The caller owns the underlying database connection.

Exported Go types and functions are declared directly in their implementation
files. The repository contract, statement builders, and errors live in `base.go`,
`statement.go`, and `errors.go` at the package root.

For example, a model implements its own conversion:

```go
func (m UserModel) ToEntity() UserEntity {
    return UserEntity{ID: m.ID, Name: m.Name}
}
```
