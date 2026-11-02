package http

import "github.com/dtalabs/go-packages/transport"

type HTTPSpec struct {
	Method string
	Path   string
}

func (HTTPSpec) IsSpec() {}

var _ transport.Spec = HTTPSpec{}
