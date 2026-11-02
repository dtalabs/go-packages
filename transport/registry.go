package transport

import (
	"fmt"
	"reflect"
	"sync"
)

type Registry struct {
	mu      sync.RWMutex
	drivers map[reflect.Type]Driver
}

var DefaultRegistry = NewRegistry()

func NewRegistry(drivers ...Driver) *Registry {
	registry := &Registry{}
	for _, driver := range drivers {
		registry.RegisterDriver(driver)
	}
	return registry
}

func (r *Registry) RegisterDriver(driver Driver) {
	if driver == nil || (reflect.ValueOf(driver).Kind() == reflect.Pointer && reflect.ValueOf(driver).IsNil()) {
		panic("transport: cannot register a nil driver")
	}
	specType := driver.SpecType()
	if specType == nil || specType.Kind() == reflect.Interface || !specType.Implements(reflect.TypeFor[Spec]()) {
		panic("transport: driver must declare a concrete type implementing Spec")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.drivers[specType]; exists {
		panic(fmt.Sprintf("transport: driver already registered for %v", specType))
	}
	if r.drivers == nil {
		r.drivers = make(map[reflect.Type]Driver)
	}
	r.drivers[specType] = driver
}

func (r *Registry) registerRoute(route *Route) {
	r.mu.RLock()
	selected := make([]Driver, 0, len(route.Specs))
	for specType := range route.Specs {
		driver, ok := r.drivers[specType]
		if !ok {
			r.mu.RUnlock()
			panic(fmt.Sprintf("transport: no driver registered for %v", specType))
		}
		selected = append(selected, driver)
	}
	r.mu.RUnlock()
	for _, driver := range selected {
		if err := driver.RegisterRoute(route); err != nil {
			panic(fmt.Errorf("transport: register route for %v: %w", driver.SpecType(), err))
		}
	}
}
