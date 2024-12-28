package config

import (
	"errors"
	"fmt"
	"strings"
)

type RouteMode string

const (
	RouteProxy  RouteMode = "proxy"
	RouteDirect RouteMode = "direct"
	RouteBlock  RouteMode = "block"
)

func (m RouteMode) Validate() error {
	switch m {
	case RouteProxy, RouteDirect, RouteBlock:
		return nil
	default:
		return fmt.Errorf("invalid route mode '%s'", m)
	}
}

func RouteModeFromString(m string) (RouteMode, error) {
	trimmed := strings.TrimSpace(m)
	if trimmed == "" {
		return "", errors.New("route mode is empty")
	}

	switch strings.ToLower(trimmed) {
	case "proxy":
		return RouteProxy, nil
	case "block":
		return RouteBlock, nil
	case "direct":
		return RouteDirect, nil
	default:
		return "", fmt.Errorf("route mode '%s' is unknown", m)
	}
}
