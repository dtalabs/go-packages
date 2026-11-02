package http

import "github.com/dtalabs/go-packages/transport"

type HTTPMiddleware interface {
	transport.Middleware
	HTTP(Handler) Handler
}

type MiddlewareFunc func(Handler) Handler

func (MiddlewareFunc) IsMiddleware() {}

func (m MiddlewareFunc) HTTP(next Handler) Handler { return m(next) }

var _ HTTPMiddleware = MiddlewareFunc(nil)
