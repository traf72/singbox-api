package core

type RouteMode string

const (
	RouteProxy  RouteMode = "proxy"
	RouteDirect RouteMode = "direct"
	RouteBlock  RouteMode = "block"
)

func (m RouteMode) isValid() bool {
	switch m {
	case RouteProxy, RouteDirect, RouteBlock:
		return true
	default:
		return false
	}
}
