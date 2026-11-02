package transport

type RouteStore[T interface {
	From(*Route) (T, error)
}] struct {
	Routes []T
}

func (s *RouteStore[T]) RegisterRoute(source *Route) error {
	var converter T
	route, err := converter.From(source)
	if err != nil {
		return err
	}
	s.Routes = append(s.Routes, route)
	return nil
}
