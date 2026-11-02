package transport

// RouteStore converts definitions and stores driver-specific routes.
// Embed it in a driver to share conversion and registration behavior.
// Register routes during startup, before serving requests.
type RouteStore[T interface {
	From(*Route) (T, error)
}] struct {
	Routes []T
}

// RegisterRoute converts a definition once and stores it only on success.
func (s *RouteStore[T]) RegisterRoute(source *Route) error {
	var converter T
	route, err := converter.From(source)
	if err != nil {
		return err
	}
	s.Routes = append(s.Routes, route)
	return nil
}
