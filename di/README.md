# DI

Dependency container, imported as `github.com/dtalabs/go-packages/di`. Resolves dependencies
from struct fields and caches one instance per registered or constructed type.

```go
type Config struct {
    Name string
}

type Service struct {
    Config Config
}

container := di.New()
registrations := []func() error{
    func() error {
        return container.RegisterInstance[Config](Config{Name: "LittleDays"})
    },
}
if err := container.Register(registrations...); err != nil {
    panic(err)
}
service := container.GetInstance[*Service]()
```

Use `RegisterInstance[T](value)` to register values and interface implementations
under an explicit type. `Register(registrations...)` runs registration callbacks
in order and stops on the first error; earlier registrations remain registered.
Module `ConfigureDI` functions can be included in the same callback list.

Register values and interface implementations explicitly. Unregistered pointers
to structs are constructed automatically; their exported fields are required
dependencies. Use `di:"false"` to skip a field. Resolution failures panic;
registration returns errors for nil or duplicate instances. The caller owns
resource cleanup.

Validate from this directory:

```bash
go vet ./...
go build ./...
```
