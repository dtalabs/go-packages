# Go packages

Shared Go modules used by Bifrost and other services. Each module directory at the repository root is an independent module; module paths follow `github.com/dtalabs/go-packages/<name>`.

| Package | Module | Purpose |
| --- | --- | --- |
| [DI](di/README.md) | `github.com/dtalabs/go-packages/di` | Dependency injection container |
| [ORM helper](ormhelper/README.md) | `github.com/dtalabs/go-packages/ormhelper` | Repository contracts and GORM adapter |
| [Storage](storage/README.md) | `github.com/dtalabs/go-packages/storage` | Object storage contract and S3 adapter |
| [Transport](transport/README.md) | `github.com/dtalabs/go-packages/transport` | HTTP routes, binding, and responses |

Clone this repository alongside `bifrost-backend`. Its application uses local `replace` directives pointing here.

Validate all modules from this directory:

```sh
for pkg in di ormhelper storage transport; do
  (cd "$pkg" && go vet ./... && go build ./...) || exit 1
done
```
