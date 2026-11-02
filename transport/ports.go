package transport

import (
	"context"
	"net/http"
)

type Handler[Output any] func(context.Context) (Output, error)

// Spec marks a driver-specific route specification.
type Spec interface {
	IsSpec()
}

// Middleware marks middleware implementations selected by each driver.
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

// Renderer overrides JSON rendering for response formats such as plain text.
type Renderer interface {
	Render(http.ResponseWriter, int)
}
