package transport

import "reflect"

// Driver identifies the specification type handled by a transport driver.
// Routes signatures and converted route storage are defined by each driver.
type Driver interface {
	SpecType() reflect.Type
	RegisterRoute(*Route) error
}

// RegisterDriver associates a concrete specification type with a driver.
// Duplicate registrations and nil drivers are configuration errors and panic.
func RegisterDriver(driver Driver) {
	DefaultRegistry.RegisterDriver(driver)
}
