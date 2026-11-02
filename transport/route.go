package transport

import (
	"context"
	"maps"
	"net/http"
	"reflect"
)

type RouteBuilder[Output any] struct {
	specs        map[reflect.Type]Spec
	summary      string
	tags         []string
	middlewares  []Middleware
	status       int
	bindRequest  func(context.Context, func(any) error) (context.Context, error)
	convert      func(Output) (any, error)
	requestType  reflect.Type
	responseType reflect.Type
}

type Route struct {
	Specs        map[reflect.Type]Spec
	Summary      string
	Tags         []string
	Middlewares  []Middleware
	Status       int
	RequestType  reflect.Type
	ResponseType reflect.Type
	BindRequest  func(context.Context, func(any) error) (context.Context, error)
	Invoke       func(context.Context) (any, error)
}

func NewRoute[Output any]() *RouteBuilder[Output] {
	return &RouteBuilder[Output]{
		specs:        make(map[reflect.Type]Spec),
		status:       http.StatusOK,
		responseType: reflect.TypeFor[Output](),
	}
}

func (b *RouteBuilder[Output]) Spec(specs ...Spec) *RouteBuilder[Output] {
	if b.specs == nil {
		b.specs = make(map[reflect.Type]Spec, len(specs))
	}
	for _, spec := range specs {
		if spec == nil {
			continue
		}
		b.specs[reflect.TypeOf(spec)] = spec
	}
	return b
}

func (b *RouteBuilder[Output]) ApiSummary(summary string) *RouteBuilder[Output] {
	b.summary = summary
	return b
}

func (b *RouteBuilder[Output]) ApiTags(tags ...string) *RouteBuilder[Output] {
	b.tags = append([]string(nil), tags...)
	return b
}

func (b *RouteBuilder[Output]) Use(middlewares ...Middleware) *RouteBuilder[Output] {
	b.middlewares = append(b.middlewares, middlewares...)
	return b
}

func (b *RouteBuilder[Output]) Request[T any]() *RouteBuilder[Output] {
	route := *b
	route.specs = maps.Clone(b.specs)
	route.requestType = reflect.TypeFor[T]()
	route.middlewares = append([]Middleware(nil), b.middlewares...)
	route.bindRequest = func(ctx context.Context, decode func(any) error) (context.Context, error) {
		var request T
		if err := decode(&request); err != nil {
			return ctx, err
		}
		if validatable, ok := any(&request).(Validatable); ok {
			if err := validatable.Validate(); err != nil {
				return ctx, err
			}
		}
		return WithRequest(ctx, request), nil
	}
	return &route
}

func (b *RouteBuilder[Output]) Response[Destination any, Pointer interface {
	*Destination
	ResponseFrom[Output]
}]() *RouteBuilder[Output] {
	route := *b
	route.specs = maps.Clone(b.specs)
	route.middlewares = append([]Middleware(nil), b.middlewares...)
	route.responseType = reflect.TypeFor[Destination]()
	route.convert = func(data Output) (any, error) {
		var dto Destination
		if err := Pointer(&dto).From(data); err != nil {
			return nil, err
		}
		if transformer, ok := any(&dto).(Transformer); ok {
			transformer.OnTransformer()
		}
		return &dto, nil
	}
	return &route
}

func (b *RouteBuilder[Output]) StatusCode(status int) *RouteBuilder[Output] {
	b.status = status
	return b
}

func (b *RouteBuilder[Output]) Handle(handle Handler[Output]) *Route {
	status, convert := b.status, b.convert
	return &Route{
		Specs:        maps.Clone(b.specs),
		Summary:      b.summary,
		Tags:         append([]string(nil), b.tags...),
		Middlewares:  append([]Middleware(nil), b.middlewares...),
		Status:       status,
		RequestType:  b.requestType,
		ResponseType: b.responseType,
		BindRequest:  b.bindRequest,
		Invoke: func(ctx context.Context) (any, error) {
			data, err := handle(ctx)
			if err != nil {
				return nil, err
			}
			if status == http.StatusNoContent {
				return nil, nil
			}
			if convert != nil {
				return convert(data)
			}
			return data, nil
		},
	}
}

func (r *Route) Register(registries ...*Registry) *Route {
	if len(registries) == 0 {
		registries = []*Registry{nil}
	}
	for _, registry := range registries {
		if registry == nil {
			registry = DefaultRegistry
		}
		registry.registerRoute(r)
	}
	return r
}
