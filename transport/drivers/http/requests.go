package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"github.com/dtalabs/go-packages/transport"

	"github.com/go-playground/form/v4"
)

var queryDecoder = newTaggedDecoder("query")
var paramDecoder = newTaggedDecoder("param")
var paramNames sync.Map // map[reflect.Type][]string

func newTaggedDecoder(tag string) *form.Decoder {
	decoder := form.NewDecoder()
	decoder.SetTagName(tag)
	decoder.SetMode(form.ModeExplicit)
	decoder.RegisterTagNameFunc(func(field reflect.StructField) string {
		if name, exists := field.Tag.Lookup(tag); exists {
			return name
		}
		typeOf := field.Type
		if typeOf.Kind() == reflect.Pointer {
			typeOf = typeOf.Elem()
		}
		if field.Anonymous && typeOf.Kind() == reflect.Struct {
			return field.Name
		}
		return ""
	})
	return decoder
}

func ParseRequest[T any](w ResponseWriter, r *Request) (T, error) {
	var request T
	if err := bindRequest(r, &request); err != nil {
		ExceptionResponse(w, transport.UnprocessableEntity)
		return request, err
	}
	if validatable, ok := any(&request).(transport.Validatable); ok {
		if err := validatable.Validate(); err != nil {
			ExceptionResponse(w, transport.UnprocessableEntity)
			return request, err
		}
	}
	return request, nil
}

// bindRequest binds JSON, explicitly tagged query fields, and then path params.
// URL-only fields should use json:"-" to exclude them from the request body.
func bindRequest(r *Request, request any) error {
	if r.Body != nil && r.Body != http.NoBody {
		decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(request); err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
			if err == nil {
				return errors.New("request body contains multiple JSON values")
			}
			return err
		}
	}
	query, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return err
	}
	if len(query) > 0 {
		if err := queryDecoder.Decode(request, query); err != nil {
			return err
		}
	}
	params := make(url.Values)
	for _, name := range requestParamNames(reflect.TypeOf(request)) {
		if value := r.PathValue(name); value != "" {
			params.Set(name, value)
		}
	}
	if len(params) > 0 {
		if err := paramDecoder.Decode(request, params); err != nil {
			return err
		}
	}
	return nil
}

// PathValue requires a name, so cache the names accepted by the param decoder.
func requestParamNames(typ reflect.Type) []string {
	if names, ok := paramNames.Load(typ); ok {
		return names.([]string)
	}
	var names []string
	visiting := make(map[reflect.Type]bool)
	var collect func(reflect.Type, string)
	collect = func(typ reflect.Type, prefix string) {
		for typ.Kind() == reflect.Pointer {
			typ = typ.Elem()
		}
		if typ.Kind() != reflect.Struct || visiting[typ] {
			return
		}
		visiting[typ] = true
		defer delete(visiting, typ)
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			if !field.IsExported() && !field.Anonymous {
				continue
			}
			name, tagged := field.Tag.Lookup("param")
			if !tagged && field.Anonymous {
				name = field.Name
			}
			name, _, _ = strings.Cut(name, ",")
			if name == "" || name == "-" {
				continue
			}
			if field.Anonymous {
				collect(field.Type, prefix)
			}
			names = append(names, prefix+name)
			collect(field.Type, prefix+name+".")
		}
	}
	collect(typ, "")
	actual, _ := paramNames.LoadOrStore(typ, names)
	return actual.([]string)
}
