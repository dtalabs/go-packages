package middleware

import (
	"context"
	"time"
	"github.com/dtalabs/go-packages/transport"
	httpdriver "github.com/dtalabs/go-packages/transport/drivers/http"
	"github.com/dtalabs/go-packages/transport/pkg/limiter"
)

type rateLimitMiddleware struct {
	limiter  *limiter.Limiter
	clientIP func(context.Context) string
}

func LimitAttempts(limit int, window time.Duration, clientIP func(context.Context) string) transport.Middleware {
	return &rateLimitMiddleware{
		limiter:  limiter.New(limit, window),
		clientIP: clientIP,
	}
}

func (*rateLimitMiddleware) IsMiddleware() {}

func (m *rateLimitMiddleware) HTTP(next httpdriver.Handler) httpdriver.Handler {
	return httpdriver.HandlerFunc(func(w httpdriver.ResponseWriter, r *httpdriver.Request) {
		if !m.limiter.Allow(m.clientIP(r.Context())) {
			httpdriver.ExceptionResponse(w, transport.TooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

var _ httpdriver.HTTPMiddleware = (*rateLimitMiddleware)(nil)
