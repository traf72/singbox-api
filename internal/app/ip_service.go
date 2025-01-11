package app

import (
	"strings"

	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/singbox"
	"github.com/traf72/singbox-api/internal/singbox/config"
	"github.com/traf72/singbox-api/internal/singbox/config/ip"
)

var (
	errIPEmptyRule = apperr.NewValidationErr("IPRule_Empty", "IP is empty")
)

type IPRule struct {
	RouteMode string `json:"routeMode"`
	IP        string `json:"ip"`
}

func (r *IPRule) toConfigRule() (*ip.Rule, apperr.Err) {
	if strings.TrimSpace(r.IP) == "" {
		return nil, errIPEmptyRule
	}

	routeMode, err := config.RouteModeFromString(r.RouteMode)
	if err != nil {
		return nil, apperr.NewValidationErr("IPRule_InvalidRouteMode", err.Error())
	}

	rule, appErr := ip.NewRule(routeMode, r.IP)
	if appErr != nil {
		return nil, appErr
	}

	return rule, nil
}

func AddIPRule(r *IPRule, restart bool) apperr.Err {
	rule, err := r.toConfigRule()
	if err != nil {
		return err
	}

	if err = ip.AddRule(rule); err != nil {
		return err
	}

	if restart {
		if err = singbox.Restart(); err != nil {
			return err
		}
	}

	return nil
}

func RemoveIPRule(r *IPRule, restart bool) apperr.Err {
	rule, err := r.toConfigRule()
	if err != nil {
		return err
	}

	if err = ip.RemoveRule(rule); err != nil {
		return err
	}

	if restart {
		if err = singbox.Restart(); err != nil {
			return err
		}
	}

	return nil
}
