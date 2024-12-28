package core

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/traf72/singbox-api/internal/apperr"
)

var (
	errEmptyDomain     = apperr.NewValidationErr("DNSRule_EmptyDomain", "domain is empty")
	errInvalidRuleType = apperr.NewValidationErr("DNSRule_InvalidType", "rule type is invalid")
)

func errDomainHasSpaces(t string) *apperr.Err {
	return apperr.NewValidationErr("DNSRule_DomainHasSpaces", fmt.Sprintf("domain '%s' has spaces", t))
}

func errInvalidDomain(d string) *apperr.Err {
	return apperr.NewValidationErr("DNSRule_InvalidDomain", fmt.Sprintf("invalid domain '%s'", d))
}

func errInvalidRegexp(r string) *apperr.Err {
	return apperr.NewValidationErr("DNSRule_InvalidRegexp", fmt.Sprintf("invalid regexp '%s'", r))
}

func errUnknownRouteMode(m RouteMode) *apperr.Err {
	return apperr.NewValidationErr("DNSRule_UnknownRouteMode", fmt.Sprintf("unknown route mode '%s'", m))
}

type DNSRuleType int

const (
	DNSRuleSuffix DNSRuleType = iota
	DNSRuleKeyword
	DNSRuleDomain
	DNSRuleRegex
)

func (k DNSRuleType) isValid() bool {
	switch k {
	case DNSRuleSuffix, DNSRuleKeyword, DNSRuleDomain, DNSRuleRegex:
		return true
	default:
		return false
	}
}

type DNSRule struct {
	kind   DNSRuleType
	mode   RouteMode
	domain string
}

func NewDNSRule(kind DNSRuleType, mode RouteMode, domain string) (*DNSRule, *apperr.Err) {
	domain = strings.ToLower(strings.TrimSpace(domain))
	t := &DNSRule{kind: kind, mode: mode, domain: domain}
	if err := t.validate(); err != nil {
		return nil, err
	}

	return t, nil
}

var domainRegex = regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

func (t *DNSRule) validate() *apperr.Err {
	if !t.kind.isValid() {
		return errInvalidRuleType
	}

	if !t.mode.isValid() {
		return errUnknownRouteMode(t.mode)
	}

	if t.domain == "" {
		return errEmptyDomain
	}

	if strings.ContainsAny(t.domain, " \t\n\r") {
		return errDomainHasSpaces(t.domain)
	}

	if t.kind == DNSRuleDomain && !domainRegex.MatchString(t.domain) {
		return errInvalidDomain(t.domain)
	}

	if t.kind == DNSRuleRegex {
		_, err := regexp.Compile(t.domain)
		if err != nil {
			return errInvalidRegexp(t.domain)
		}
	}

	return nil
}

var dnsRoute = map[RouteMode]string{
	RouteDirect: "dns-direct",
	RouteProxy:  "dns-remote",
	RouteBlock:  "dns-block",
}

func AddDNSRule(r *DNSRule) *apperr.Err {
	c, err := load()
	if err != nil {
		return err
	}

	added := addRule(r, c.config)
	if added {
		if err := save(c); err != nil {
			return err
		}
	}

	return nil
}

func addRule(r *DNSRule, c *config) (added bool) {
	addedToRoute := addRuleToRoute(r, c)
	addedToDNS := addRuleToDNS(r, c)
	return addedToRoute || addedToDNS
}

func addRuleToRoute(r *DNSRule, c *config) bool {
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

func addRuleToDNS(r *DNSRule, c *config) bool {
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

func RemoveDNSRule(r *DNSRule) *apperr.Err {
	c, err := load()
	if err != nil {
		return err
	}

	removed := removeRule(r, c.config)
	if removed {
		if err := save(c); err != nil {
			return err
		}
	}

	return nil
}

func removeRule(r *DNSRule, c *config) (removed bool) {
	removedFromRoute := removeFromRoute(r, c)
	removedFromDNS := removeFromDNS(r, c)
	return removedFromRoute || removedFromDNS
}

func removeFromRoute(r *DNSRule, c *config) bool {
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

func removeFromDNS(r *DNSRule, c *config) bool {
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

func getRouteRules(r *DNSRule, c *config) *[]string {
	mode := string(r.mode)
	ruleSetIdx := slices.IndexFunc(c.Route.Rules, func(rr routeRule) bool {
		return rr.Outbound == mode
	})

	if ruleSetIdx == -1 {
		c.Route.Rules = append(c.Route.Rules, routeRule{
			Outbound: mode,
			rule:     rule{},
		})
		ruleSetIdx = len(c.Route.Rules) - 1
	}

	ruleSet := &c.Route.Rules[ruleSetIdx]
	return getRulesByType(r.kind, &ruleSet.rule)
}

func getDNSRules(r *DNSRule, c *config) *[]string {
	ruleSetIdx := slices.IndexFunc(c.DNS.Rules, func(dr dnsRule) bool {
		return dr.Server == dnsRoute[r.mode]
	})

	if ruleSetIdx == -1 {
		c.DNS.Rules = append(c.DNS.Rules, dnsRule{
			Server: dnsRoute[r.mode],
			rule:   rule{},
		})
		ruleSetIdx = len(c.Route.Rules) - 1
	}

	ruleSet := &c.DNS.Rules[ruleSetIdx]
	return getRulesByType(r.kind, &ruleSet.rule)
}

func getRulesByType(t DNSRuleType, r *rule) *[]string {
	switch t {
	case DNSRuleSuffix:
		return &r.DomainSuffix
	case DNSRuleKeyword:
		return &r.DomainKeyword
	case DNSRuleDomain:
		return &r.Domain
	case DNSRuleRegex:
		return &r.DomainRegex
	default:
		return nil
	}
}
