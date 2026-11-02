package http

import "github.com/dtalabs/go-packages/transport"

// HTTPMiddleware wraps handlers for the HTTP driver.
type HTTPMiddleware interface {
	transport.Middleware
	HTTP(Handler) Handler
}

// MiddlewareFunc adapts standard HTTP middleware for use in transport routes.
type MiddlewareFunc func(Handler) Handler

func (MiddlewareFunc) IsMiddleware() {}

func (m MiddlewareFunc) HTTP(next Handler) Handler { return m(next) }

var _ HTTPMiddleware = MiddlewareFunc(nil)
