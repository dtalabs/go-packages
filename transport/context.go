package transport

import "context"

type requestContextKey[T any] struct{}

func WithRequest[T any](ctx context.Context, request T) context.Context {
	return context.WithValue(ctx, requestContextKey[T]{}, request)
}

func GetRequest[T any](ctx context.Context) T {
	request, _ := ctx.Value(requestContextKey[T]{}).(T)
	return request
}
