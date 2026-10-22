package app

import (
	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/singbox"
	"github.com/traf72/singbox-api/internal/singbox/config"
)

type RouteModeRequest struct {
	Mode string `json:"mode"`
}

func SetRouteMode(request *RouteModeRequest, restart bool) apperr.Err {
	mode, err := config.RouteModeFromString(request.Mode)
	if err != nil {
		return apperr.NewValidationErr("Route_InvalidMode", err.Error())
	}
	if mode != config.RouteProxy && mode != config.RouteDirect {
		return apperr.NewValidationErr("Route_UnsupportedMode", "route mode must be 'proxy' or 'direct'")
	}

	if err := config.SetFinalRouteMode(mode); err != nil {
		return err
	}

	if restart {
		return singbox.Restart()
	}

	return nil
}
