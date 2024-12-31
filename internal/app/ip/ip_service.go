package ip

import (
	"strings"

	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/config"
	"github.com/traf72/singbox-api/internal/config/ip"
	"github.com/traf72/singbox-api/internal/config/singbox"
)

var (
	errEmptyIP = apperr.NewValidationErr("IPRule_Empty", "IP is empty")
)

type Rule struct {
	RouteMode string `json:"routeMode"`
	IP        string `json:"ip"`
}

func (r *Rule) toConfigRule() (*ip.Rule, apperr.Err) {
	if strings.TrimSpace(r.IP) == "" {
		return nil, errEmptyIP
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

func AddRule(r *Rule) apperr.Err {
	rule, err := r.toConfigRule()
	if err != nil {
		return err
	}

	if err = ip.AddRule(rule); err != nil {
		return err
	}

	if err = singbox.Restart(); err != nil {
		return err
	}

	return nil
}

func RemoveRule(r *Rule) apperr.Err {
	rule, err := r.toConfigRule()
	if err != nil {
		return err
	}

	if err = ip.RemoveRule(rule); err != nil {
		return err
	}

	if err = singbox.Restart(); err != nil {
		return err
	}

	return nil
}
