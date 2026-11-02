package http

import (
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"slices"
	"github.com/dtalabs/go-packages/transport"
)

type Route struct {
	Method       string
	Path         string
	Handler      Handler
	Summary      string
	Tags         []string
	Status       int
	RequestType  reflect.Type
	ResponseType reflect.Type
}

// From builds an HTTP route from its transport definition and HTTP specification.
func (Route) From(source *transport.Route) (Route, error) {
	if source == nil {
		return Route{}, fmt.Errorf("missing HTTP route definition")
	}
	spec, ok := source.Specs[reflect.TypeFor[HTTPSpec]()].(HTTPSpec)
	if !ok || spec.Method == "" || spec.Path == "" {
		return Route{}, fmt.Errorf("invalid HTTP specification: %+v", spec)
	}
	return Route{
		Method: spec.Method, Path: spec.Path, Handler: buildHandler(source),
		Summary: source.Summary, Status: source.Status,
		Tags:        append([]string(nil), source.Tags...),
		RequestType: source.RequestType, ResponseType: source.ResponseType,
	}, nil
}

func buildHandler(route *transport.Route) Handler {
	status, bind, invoke := route.Status, route.BindRequest, route.Invoke
	var handler Handler = HandlerFunc(func(w ResponseWriter, r *Request) {
		ctx := r.Context()
		if bind != nil {
			var err error
			ctx, err = bind(ctx, func(dst any) error {
				return bindRequest(r, dst)
			})
			if err != nil {
				ExceptionResponse(w, transport.UnprocessableEntity)
				return
			}
		}
		response, err := invoke(ctx)
		if err != nil {
			exception := transport.ToException(err)
			if exception.Status == http.StatusInternalServerError {
				slog.ErrorContext(ctx, "HTTP response failed", "error", err)
			}
			ExceptionResponse(w, exception)
			return
		}
		if status == http.StatusNoContent {
			w.WriteHeader(status)
			return
		}
		if renderer, ok := response.(transport.Renderer); ok {
			renderer.Render(w, status)
			return
		}
		JSONResponse(w, status, response)
	})
	for _, v := range slices.Backward(route.Middlewares) {
		if middleware, ok := v.(HTTPMiddleware); ok {
			handler = middleware.HTTP(handler)
		}
	}
	return handler
}
