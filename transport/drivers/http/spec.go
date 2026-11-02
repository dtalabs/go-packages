package http

import "github.com/dtalabs/go-packages/transport"

// HTTPSpec configures a route for the HTTP driver.
type HTTPSpec struct {
	Method string
	Path   string
}

// IsSpec marks HTTPSpec as a transport route specification.
func (HTTPSpec) IsSpec() {}

var _ transport.Spec = HTTPSpec{}
