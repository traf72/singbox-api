package ip

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/singbox/config"
)

var (
	errEmptyIP = apperr.NewValidationErr("IPRule_EmptyIP", "IP is empty")
)

func errInvalidIP(ip string) apperr.Err {
	return apperr.NewValidationErr("IPRule_InvalidIP", fmt.Sprintf("IP '%s' is invalid", ip))
}

type Rule struct {
	mode config.RouteMode
	ip   string
}

func NewRule(m config.RouteMode, ip string) (*Rule, apperr.Err) {
	ip = strings.TrimSpace(ip)
	rule := &Rule{mode: m, ip: ip}
	if err := rule.validate(); err != nil {
		return nil, err
	}

	return rule, nil
}

var ipRegex = regexp.MustCompile(`^([01]?\d\d?|2[0-4]\d|25[0-5])(?:\.(?:[01]?\d\d?|2[0-4]\d|25[0-5])){3}(?:/[0-2]\d|/3[0-2])?$`)

func (r *Rule) validate() apperr.Err {
	if err := r.mode.Validate(); err != nil {
		return apperr.NewValidationErr("IPRule_InvalidRouteMode", err.Error())
	}

	if r.ip == "" {
		return errEmptyIP
	}

	if !ipRegex.MatchString(r.ip) {
		return errInvalidIP(r.ip)
	}

	return nil
}

func AddRule(r *Rule) apperr.Err {
	c, err := config.Load()
	if err != nil {
		return err
	}

	added := add(r, c.Conf)
	if added {
		if err := config.Save(c); err != nil {
			return err
		}
	}

	return nil
}

func add(r *Rule, c *config.Conf) (added bool) {
	rules := getRouteRules(r.mode, c)
	ruleIdx := slices.IndexFunc(*rules, func(ip string) bool {
		return strings.TrimSpace(ip) == r.ip
	})

	if ruleIdx == -1 {
		*rules = append(*rules, r.ip)
		return true
	}

	return false
}

func RemoveRule(r *Rule) apperr.Err {
	c, err := config.Load()
	if err != nil {
		return err
	}

	removed := removeRule(r, c.Conf)
	if removed {
		if err := config.Save(c); err != nil {
			return err
		}
	}

	return nil
}

func removeRule(r *Rule, c *config.Conf) (removed bool) {
	rules := getRouteRules(r.mode, c)
	ruleIdx := slices.IndexFunc(*rules, func(d string) bool {
		return strings.TrimSpace(d) == r.ip
	})

	if ruleIdx == -1 {
		return false
	}

	*rules = slices.Delete(*rules, ruleIdx, ruleIdx+1)
	return true
}

func getRouteRules(m config.RouteMode, c *config.Conf) *[]string {
	mode := string(m)
	ruleSetIdx := slices.IndexFunc(c.Route.Rules, func(rr config.RouteRule) bool {
		return rr.Outbound == mode
	})

	if ruleSetIdx == -1 {
		c.Route.Rules = append(c.Route.Rules, config.RouteRule{
			Outbound: mode,
			Rule:     config.Rule{},
		})
		ruleSetIdx = len(c.Route.Rules) - 1
	}

	ruleSet := &c.Route.Rules[ruleSetIdx]
	return &ruleSet.IP_CIDR
}
