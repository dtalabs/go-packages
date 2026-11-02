package http

import (
	"fmt"
	"net/http"
	"reflect"
	"github.com/dtalabs/go-packages/transport"

	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi3"
)

type Docs struct {
	reflector openapi3.Reflector
}

var Documentation = NewDocs("API", "1.0.0")

func NewDocs(title, version string) *Docs {
	docs := &Docs{}
	docs.reflector.SpecEns().Info.WithTitle(title).WithVersion(version)
	docs.reflector.DefaultOptions = append(docs.reflector.DefaultOptions,
		jsonschema.InterceptDefName(func(t reflect.Type, defaultDefName string) string {
			for t.Kind() == reflect.Pointer {
				t = t.Elem()
			}
			if name := t.Name(); name != "" {
				return name
			}
			return defaultDefName
		}),
	)
	return docs
}

func (d *Docs) Add(route Route) error {
	operation, err := d.reflector.NewOperationContext(route.Method, route.Path)
	if err != nil {
		return err
	}
	if route.Summary != "" {
		operation.SetSummary(route.Summary)
	}
	if len(route.Tags) > 0 {
		operation.SetTags(route.Tags...)
	}
	if route.RequestType != nil {
		operation.AddReqStructure(reflect.New(route.RequestType).Interface())
	}
	_, _, params, err := openapi.SanitizeMethodPath(route.Method, route.Path)
	if err != nil {
		return err
	}
	if len(params) > 0 {
		fields := make([]reflect.StructField, len(params))
		for i, name := range params {
			fields[i] = reflect.StructField{
				Name: fmt.Sprintf("Param%d", i), Type: reflect.TypeFor[string](),
				Tag: reflect.StructTag(fmt.Sprintf("json:\"-\" path:%q", name)),
			}
		}
		operation.AddReqStructure(reflect.New(reflect.StructOf(fields)).Interface())
	}
	var response any
	if route.Status != http.StatusNoContent && route.ResponseType != nil {
		value := reflect.New(route.ResponseType).Interface()
		if _, custom := value.(transport.Renderer); !custom {
			response = value
		}
	}
	operation.AddRespStructure(response, openapi.WithHTTPStatus(route.Status))
	return d.reflector.AddOperation(operation)
}

func (d *Docs) OpenAPISpec() HandlerFunc {
	document, err := d.reflector.SpecEns().MarshalYAML()
	if err != nil {
		panic(err)
	}
	return func(w ResponseWriter, r *Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if r.Method != http.MethodHead {
			_, _ = w.Write(document)
		}
	}
}

func (d *Docs) Scalar(specURL string) HandlerFunc {
	title := d.reflector.SpecEns().Info.Title
	if title == "" {
		title = "API Reference"
	}
	html := fmt.Sprintf(`
		<!doctype html>
		<html>
		  <head>
			<title>%s</title>
			<meta charset="utf-8" />
			<meta name="viewport" content="width=device-width, initial-scale=1" />
			<style>
			  body {
				margin: 0;
			  }
			</style>
		  </head>
		  <body>
			<script id="api-reference" data-url="%s"></script>
			<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference@1.74.0"></script>
		  </body>
		</html>`,
		title,
		specURL,
	)

	return func(w ResponseWriter, r *Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if r.Method != http.MethodHead {
			_, _ = w.Write([]byte(html))
		}
	}
}
