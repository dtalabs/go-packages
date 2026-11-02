package di

import (
	"fmt"
	"reflect"
	"strings"
)

// resolver tracks recursive dependencies during automatic construction.
type resolver struct {
	container *Container
	path      []reflect.Type
}

func (r *resolver) GetInstance[T any]() T {
	instance, err := r.resolve(reflect.TypeFor[T]())
	if err != nil {
		panic(err)
	}
	return instance.(T)
}

func (r *resolver) resolve(key reflect.Type) (any, error) {
	for _, ancestor := range r.path {
		if ancestor == key {
			return nil, r.describe(fmt.Errorf("%w: %v", ErrDependencyCycle, key), key)
		}
	}
	c := r.container
	c.mu.RLock()
	instance, exists := c.instances[key]
	c.mu.RUnlock()
	if exists {
		return instance, nil
	}
	child := &resolver{container: c, path: append(append([]reflect.Type{}, r.path...), key)}
	if key.Kind() != reflect.Pointer || key.Elem().Kind() != reflect.Struct {
		return nil, r.describe(fmt.Errorf("%w: %v", ErrNotRegistered, key), key)
	}
	value := reflect.New(key.Elem())
	for i := 0; i < key.Elem().NumField(); i++ {
		field := key.Elem().Field(i)
		if field.Tag.Get("di") == "false" {
			continue
		}
		if !field.IsExported() {
			return nil, child.describe(fmt.Errorf("field %s is private; register an initialized instance", field.Name), field.Type)
		}
		dependency, err := child.resolve(field.Type)
		if err != nil {
			return nil, fmt.Errorf("field %s.%s: %w", key, field.Name, err)
		}
		value.Elem().Field(i).Set(reflect.ValueOf(dependency))
	}
	instance = value.Interface()
	c.mu.Lock()
	defer c.mu.Unlock()
	if previous, ok := c.instances[key]; ok {
		return previous, nil
	}
	if c.instances == nil {
		c.instances = make(map[reflect.Type]any)
	}
	c.instances[key] = instance
	return instance, nil
}

func (r *resolver) describe(err error, key reflect.Type) error {
	names := make([]string, 0, len(r.path)+1)
	for _, step := range r.path {
		names = append(names, step.String())
	}
	names = append(names, key.String())
	return fmt.Errorf("resolve %s: %w", strings.Join(names, " -> "), err)
}
