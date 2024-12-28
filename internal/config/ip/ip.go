package ip

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/config"
)

var (
	errEmptyIP = apperr.NewValidationErr("IPRule_EmptyIP", "IP is empty")
)

func errInvalidIP(ip string) *apperr.Err {
	return apperr.NewValidationErr("IPRule_InvalidIP", fmt.Sprintf("IP '%s' is invalid", ip))
}

type Rule struct {
	mode config.RouteMode
	ip   string
}

func NewRule(m config.RouteMode, ip string) (*Rule, *apperr.Err) {
	ip = strings.TrimSpace(ip)
	t := &Rule{mode: m, ip: ip}
	if err := t.validate(); err != nil {
		return nil, err
	}

	return t, nil
}

var ipRegex = regexp.MustCompile(`^([01]?\d\d?|2[0-4]\d|25[0-5])(?:\.(?:[01]?\d\d?|2[0-4]\d|25[0-5])){3}(?:/[0-2]\d|/3[0-2])?$`)

func (t *Rule) validate() *apperr.Err {
	if err := t.mode.Validate(); err != nil {
		return apperr.NewValidationErr("IPRule_InvalidRouteMode", err.Error())
	}

	if t.ip == "" {
		return errEmptyIP
	}

	if !ipRegex.MatchString(t.ip) {
		return errInvalidIP(t.ip)
	}

	return nil
}

func AddRule(r *Rule) *apperr.Err {
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
	addedToRoute := addToRoute(r, c)
	return addedToRoute
}

func addToRoute(r *Rule, c *config.Conf) bool {
	rules := getRouteRules(r, c)
	ruleIdx := slices.IndexFunc(*rules, func(ip string) bool {
		return strings.TrimSpace(ip) == r.ip
	})

	if ruleIdx == -1 {
		*rules = append(*rules, r.ip)
		return true
	}

	return false
}

func RemoveRule(r *Rule) *apperr.Err {
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
	removedFromRoute := removeFromRoute(r, c)
	return removedFromRoute
}

func removeFromRoute(r *Rule, c *config.Conf) bool {
	return true
}

func getRouteRules(r *Rule, c *config.Conf) *[]string {
	return nil
}
