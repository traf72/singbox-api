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
	mode   config.RouteMode
	target *config.RuleTarget
	ip     string
}

func NewRule(m config.RouteMode, ip string) (*Rule, apperr.Err) {
	return NewRuleForOutbound(m, "", ip)
}

func NewRuleForOutbound(m config.RouteMode, outbound, ip string) (*Rule, apperr.Err) {
	if err := m.Validate(); err != nil {
		return nil, apperr.NewValidationErr("IPRule_InvalidRouteMode", err.Error())
	}
	target, err := config.NewRuleTarget(m, outbound)
	if err != nil {
		return nil, apperr.NewValidationErr("IPRule_InvalidOutbound", err.Error())
	}
	ip = strings.TrimSpace(ip)
	rule := &Rule{mode: m, target: target, ip: ip}
	if err := rule.validate(); err != nil {
		return nil, err
	}

	return rule, nil
}

var ipRegex = regexp.MustCompile(`^([01]?\d\d?|2[0-4]\d|25[0-5])(?:\.(?:[01]?\d\d?|2[0-4]\d|25[0-5])){3}(?:/(?:[0-9]|[12]\d|3[0-2]))?$`)

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
	if err := r.target.ValidateForAdd(c.Conf); err != nil {
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
	ruleSet := config.FindOrCreateRouteRule(c, r.target)
	rules := &ruleSet.IP_CIDR
	ruleIdx := slices.IndexFunc(*rules, func(ip string) bool {
		return strings.TrimSpace(ip) == r.ip
	})

	if ruleIdx == -1 {
		*rules = append(*rules, r.ip)
		config.NormalizeRouteRuleOrder(c)
		return true
	}

	return false
}

func RemoveRule(r *Rule) apperr.Err {
	c, err := config.Load()
	if err != nil {
		return err
	}
	if err := r.target.ValidateForRemoval(c.Conf); err != nil {
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
	ruleSet, ok := config.FindRouteRule(c, r.target)
	if !ok {
		return false
	}
	rules := &ruleSet.IP_CIDR
	ruleIdx := slices.IndexFunc(*rules, func(d string) bool {
		return strings.TrimSpace(d) == r.ip
	})

	if ruleIdx == -1 {
		return false
	}

	*rules = slices.Delete(*rules, ruleIdx, ruleIdx+1)
	config.RemoveRouteRuleIfEmpty(c, r.target)
	config.NormalizeRouteRuleOrder(c)
	return true
}

func FindAssignments(r *Rule) ([]config.RuleAssignment, apperr.Err) {
	c, err := config.Load()
	if err != nil {
		return nil, err
	}
	return config.FindRuleAssignments(c.Conf, config.RuleIPCIDR, r.ip), nil
}
