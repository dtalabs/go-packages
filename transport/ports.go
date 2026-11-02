package transport

import (
	"context"
	"net/http"
)

type Handler[Output any] func(context.Context) (Output, error)

type Spec interface {
	IsSpec()
}

type Middleware interface {
	IsMiddleware()
}

type Validatable interface {
	Validate() error
}

type Transformer interface {
	OnTransformer()
}

type ResponseFrom[Source any] interface {
	From(Source) error
}

type Renderer interface {
	Render(http.ResponseWriter, int)
}
