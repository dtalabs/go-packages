package http

import (
	"encoding/json"
	"net/http"
	"reflect"
	"github.com/dtalabs/go-packages/transport"
)

func JSONResponse(w ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func ExceptionResponse(w ResponseWriter, exception transport.Exception) {
	JSONResponse(w, exception.Status, exception)
}

type Driver struct {
	transport.RouteStore[Route]
}

func New() *Driver { return &Driver{} }

func (*Driver) SpecType() reflect.Type { return reflect.TypeFor[HTTPSpec]() }

func (d *Driver) GetRoutes(includeDocs bool) []Route {
	routes := append([]Route(nil), d.RouteStore.Routes...)
	if includeDocs {
		docs := NewDocs("API", "1.0.0")
		for _, route := range routes {
			if err := docs.Add(route); err != nil {
				panic(err)
			}
		}
		routes = append(routes,
			Route{Method: http.MethodGet, Path: "/docs", Handler: docs.Scalar("/docs/openapi.yaml")},
			Route{Method: http.MethodGet, Path: "/docs/openapi.yaml", Handler: docs.OpenAPISpec()},
		)
		Documentation = docs
	}
	return routes
}

var _ transport.Driver = (*Driver)(nil)
