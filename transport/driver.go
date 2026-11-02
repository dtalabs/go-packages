package transport

import "reflect"

type Driver interface {
	SpecType() reflect.Type
	RegisterRoute(*Route) error
}

func RegisterDriver(driver Driver) {
	DefaultRegistry.RegisterDriver(driver)
}
