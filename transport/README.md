# Transport

Declarative route registration with an HTTP driver. The core defines typed
handlers, request context, response mapping, middleware, exceptions, and driver
registries. This package includes HTTP support only.

```go
import (
    "context"
    "net/http"
    "github.com/dtalabs/go-packages/transport"
    httpdriver "github.com/dtalabs/go-packages/transport/drivers/http"
)

type CreateRequest struct {
    Name string `json:"name"`
}

type Item struct {
    Name string `json:"name"`
}

func Routes() http.Handler {
    driver := httpdriver.New()
    registry := transport.NewRegistry(driver)

    transport.NewRoute[Item]().
        Spec(httpdriver.HTTPSpec{Method: http.MethodPost, Path: "/items"}).
        ApiTags("Items").
        ApiSummary("Create an item").
        Request[CreateRequest]().
        StatusCode(http.StatusCreated).
        Handle(func(ctx context.Context) (Item, error) {
            request := transport.GetRequest[CreateRequest](ctx)
            return Item{Name: request.Name}, nil
        }).
        Register(registry)

    mux := http.NewServeMux()
    for _, route := range driver.GetRoutes(true) {
        mux.Handle(route.Method+" "+route.Path, route.Handler)
    }
    return mux
}
```

Register drivers before routes and mount the converted handlers before serving
requests. Each server can own a registry; package-level `RegisterDriver` and
`Route.Register()` use `DefaultRegistry`. Missing drivers, duplicate driver
registrations, and conversion failures panic during startup.

`GetRoutes(true)` includes `/docs` and `/docs/openapi.yaml`, generated from the
registered HTTP routes. `GetRoutes(false)` returns only application routes.
`ApiTags("Items")` groups an operation in Scalar/OpenAPI. Pass multiple tags
to include an operation in multiple groups. Tags are copied when the route is
built and converted by the HTTP driver.

Handlers are built once; repeated calls return a copy of the route slice.

Handlers accept `context.Context` and return the declared output type plus an
error. `Request[T]()` binds JSON, explicitly tagged `query` fields, then `param`
fields from `Request.PathValue`. Use `json:"-"` on URL-only fields. A pointer to
the request may implement `Validate() error`; binding and validation failures
return HTTP 422. Bound values are available through `transport.GetRequest[T]`.

`Response[DTO]()` requires `*DTO` to implement `From(Output) error`. After
conversion, `OnTransformer()` runs when implemented. Without response mapping,
output is serialized directly. `StatusCode` defaults to 200; 204 emits no body.
A response implementing `transport.Renderer` can render a custom format.

Use `transport.Register` to map domain errors to `transport.Exception`. Unmapped
errors return HTTP 500. `httpdriver.ExceptionResponse` and `JSONResponse` are
available for raw HTTP handlers.

Route middleware implements `transport.Middleware` (`IsMiddleware()`) and
`httpdriver.HTTPMiddleware` (`HTTP(Handler) Handler`). Adapt existing standard
HTTP middleware with `httpdriver.MiddlewareFunc(authenticate)` and pass it to
`.Use(...)`. Middleware runs in declaration order, before request binding.

`github.com/dtalabs/go-packages/transport/middleware` provides CORS, security headers, and `LimitAttempts`.
The rate limiter accepts a caller-owned `func(context.Context) string` that
returns a client IP. The application must populate any context values it needs
before invoking the route.

Validate from this directory with `go vet ./...` and `go build ./...`.
