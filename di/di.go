package di

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// Container caches singleton instances. Its zero value is ready to use.
// Resource ownership remains with the caller.
type Container struct {
	mu        sync.RWMutex
	resolveMu sync.Mutex
	instances map[reflect.Type]any
}

func New() *Container { return &Container{} }

var shared = New()

func GetContainer() *Container { return shared }

func validateInstance(key reflect.Type, instance any) error {
	value := reflect.ValueOf(instance)
	if !value.IsValid() {
		return fmt.Errorf("%w: %v", ErrNilInstance, key)
	}
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		if value.IsNil() {
			return fmt.Errorf("%w: %v", ErrNilInstance, key)
		}
	}
	return nil
}

// Register runs registration callbacks in order, stopping at the first error.
// Registrations completed before an error remain registered.
func (c *Container) Register(registrations ...func() error) error {
	for _, register := range registrations {
		if register == nil {
			return fmt.Errorf("%w: registration callback", ErrNilInstance)
		}
		if err := register(); err != nil {
			return err
		}
	}
	return nil
}

// RegisterInstance registers a value under its explicit type.
func (c *Container) RegisterInstance[T any](instance T) error {
	key := reflect.TypeFor[T]()
	if err := validateInstance(key, instance); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.instances[key]; ok {
		return fmt.Errorf("%w: %v", ErrAlreadyRegistered, key)
	}
	if c.instances == nil {
		c.instances = make(map[reflect.Type]any)
	}
	c.instances[key] = instance
	return nil
}

// GetInstance resolves registered types or constructs a pointer to a struct.
// Fields tagged di:"false" are skipped and retain their zero values.
// Other exported fields are required dependencies; private state requires an explicitly registered instance.
func (c *Container) GetInstance[T any]() T {
	return c.Resolve(reflect.TypeFor[T]()).(T)
}

func (c *Container) Resolve(key reflect.Type) any {
	c.resolveMu.Lock()
	defer c.resolveMu.Unlock()
	instance, err := (&resolver{container: c}).resolve(key)
	if err != nil {
		panic(err)
	}
	return instance
}

var (
	ErrNotRegistered     = errors.New("dependency is not registered")
	ErrAlreadyRegistered = errors.New("dependency is already registered")
	ErrNilInstance       = errors.New("dependency instance is nil")
	ErrDependencyCycle   = errors.New("dependency cycle")
)
