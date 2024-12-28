package dns

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/config"
)

var (
	errEmptyDomain     = apperr.NewValidationErr("DNSRule_EmptyDomain", "domain is empty")
	errInvalidRuleType = apperr.NewValidationErr("DNSRule_InvalidType", "rule type is invalid")
)

func errDomainHasSpaces(t string) *apperr.Err {
	return apperr.NewValidationErr("DNSRule_DomainHasSpaces", fmt.Sprintf("domain '%s' has spaces", t))
}

func errInvalidDomain(d string) *apperr.Err {
	return apperr.NewValidationErr("DNSRule_InvalidDomain", fmt.Sprintf("domain '%s' is invalid", d))
}

func errInvalidRegexp(r string) *apperr.Err {
	return apperr.NewValidationErr("DNSRule_InvalidRegexp", fmt.Sprintf("regexp is '%s' invalid", r))
}

type RuleType int

const (
	Suffix RuleType = iota
	Keyword
	Domain
	Regex
)

func (k RuleType) isValid() bool {
	switch k {
	case Suffix, Keyword, Domain, Regex:
		return true
	default:
		return false
	}
}

type Rule struct {
	kind   RuleType
	mode   config.RouteMode
	domain string
}

func NewRule(kind RuleType, mode config.RouteMode, domain string) (*Rule, *apperr.Err) {
	domain = strings.ToLower(strings.TrimSpace(domain))
	t := &Rule{kind: kind, mode: mode, domain: domain}
	if err := t.validate(); err != nil {
		return nil, err
	}

	return t, nil
}

var domainRegex = regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

func (t *Rule) validate() *apperr.Err {
	if !t.kind.isValid() {
		return errInvalidRuleType
	}

	if err := t.mode.Validate(); err != nil {
		return apperr.NewValidationErr("DNSRule_InvalidRouteMode", err.Error())
	}

	if t.domain == "" {
		return errEmptyDomain
	}

	if strings.ContainsAny(t.domain, " \t\n\r") {
		return errDomainHasSpaces(t.domain)
	}

	if t.kind == Domain && !domainRegex.MatchString(t.domain) {
		return errInvalidDomain(t.domain)
	}

	if t.kind == Regex {
		_, err := regexp.Compile(t.domain)
		if err != nil {
			return errInvalidRegexp(t.domain)
		}
	}

	return nil
}

var dnsRoute = map[config.RouteMode]string{
	config.RouteDirect: "dns-direct",
	config.RouteProxy:  "dns-remote",
	config.RouteBlock:  "dns-block",
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
	addedToDNS := addToDNS(r, c)
	return addedToRoute || addedToDNS
}

func addToRoute(r *Rule, c *config.Conf) bool {
	rulesSlice := getRouteRules(r, c)
	ruleIdx := slices.IndexFunc(*rulesSlice, func(d string) bool {
		return strings.EqualFold(strings.TrimSpace(d), r.domain)
	})

	if ruleIdx == -1 {
		*rulesSlice = append(*rulesSlice, r.domain)
		return true
	}

	return false
}

func addToDNS(r *Rule, c *config.Conf) bool {
	rulesSlice := getDNSRules(r, c)
	ruleIdx := slices.IndexFunc(*rulesSlice, func(d string) bool {
		return strings.EqualFold(strings.TrimSpace(d), r.domain)
	})

	if ruleIdx == -1 {
		*rulesSlice = append(*rulesSlice, r.domain)
		return true
	}

	return false
}

func RemoveRule(r *Rule) *apperr.Err {
	c, err := config.Load()
	if err != nil {
		return err
	}

	removed := remove(r, c.Conf)
	if removed {
		if err := config.Save(c); err != nil {
			return err
		}
	}

	return nil
}

func remove(r *Rule, c *config.Conf) (removed bool) {
	removedFromRoute := removeFromRoute(r, c)
	removedFromDNS := removeFromDNS(r, c)
	return removedFromRoute || removedFromDNS
}

func removeFromRoute(r *Rule, c *config.Conf) bool {
	rulesSlice := getRouteRules(r, c)
	ruleIdx := slices.IndexFunc(*rulesSlice, func(d string) bool {
		return strings.EqualFold(strings.TrimSpace(d), r.domain)
	})

	if ruleIdx == -1 {
		return false
	}

	*rulesSlice = slices.Delete(*rulesSlice, ruleIdx, ruleIdx+1)
	return true
}

func removeFromDNS(r *Rule, c *config.Conf) bool {
	rulesSlice := getDNSRules(r, c)
	ruleIdx := slices.IndexFunc(*rulesSlice, func(d string) bool {
		return strings.EqualFold(strings.TrimSpace(d), r.domain)
	})

	if ruleIdx == -1 {
		return false
	}

	*rulesSlice = slices.Delete(*rulesSlice, ruleIdx, ruleIdx+1)
	return true
}

func getRouteRules(r *Rule, c *config.Conf) *[]string {
	mode := string(r.mode)
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
	return getRulesForType(r.kind, &ruleSet.Rule)
}

func getDNSRules(r *Rule, c *config.Conf) *[]string {
	ruleSetIdx := slices.IndexFunc(c.DNS.Rules, func(dr config.DNSRule) bool {
		return dr.Server == dnsRoute[r.mode]
	})

	if ruleSetIdx == -1 {
		c.DNS.Rules = append(c.DNS.Rules, config.DNSRule{
			Server: dnsRoute[r.mode],
			Rule:   config.Rule{},
		})
		ruleSetIdx = len(c.Route.Rules) - 1
	}

	ruleSet := &c.DNS.Rules[ruleSetIdx]
	return getRulesForType(r.kind, &ruleSet.Rule)
}

func getRulesForType(t RuleType, r *config.Rule) *[]string {
	switch t {
	case Suffix:
		return &r.DomainSuffix
	case Keyword:
		return &r.DomainKeyword
	case Domain:
		return &r.Domain
	case Regex:
		return &r.DomainRegex
	default:
		return nil
	}
}
