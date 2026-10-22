package config

import (
	"fmt"

	"github.com/traf72/singbox-api/internal/apperr"
)

func errUnsupportedFinalRouteMode(mode RouteMode) apperr.Err {
	return apperr.NewValidationErr("Route_UnsupportedFinalMode", fmt.Sprintf("route mode '%s' is unsupported; expected 'proxy' or 'direct'", mode))
}

func SetFinalRouteMode(mode RouteMode) apperr.Err {
	c, err := Load()
	if err != nil {
		return err
	}

	if err := setFinalRouteMode(mode, c.Conf); err != nil {
		return err
	}

	return Save(c)
}

func setFinalRouteMode(mode RouteMode, conf *Conf) apperr.Err {
	switch mode {
	case RouteProxy:
		conf.Route.Final = string(RouteProxy)
		conf.DNS.Final = "dns-remote"
	case RouteDirect:
		conf.Route.Final = string(RouteDirect)
		conf.DNS.Final = "dns-direct"
	default:
		return errUnsupportedFinalRouteMode(mode)
	}

	return nil
}
