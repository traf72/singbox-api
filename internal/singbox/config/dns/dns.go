package dns

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/singbox/config"
)

var (
	errEmptyDomain     = apperr.NewValidationErr("DNSRule_EmptyDomain", "domain is empty")
	errInvalidRuleType = apperr.NewValidationErr("DNSRule_InvalidType", "rule type is invalid")
)

func errDomainHasSpaces(t string) apperr.Err {
	return apperr.NewValidationErr("DNSRule_DomainHasSpaces", fmt.Sprintf("domain '%s' has spaces", t))
}

func errInvalidDomain(d string) apperr.Err {
	return apperr.NewValidationErr("DNSRule_InvalidDomain", fmt.Sprintf("domain '%s' is invalid", d))
}

func errInvalidRegexp(r string) apperr.Err {
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
	target *config.RuleTarget
	domain string
}

func NewRule(kind RuleType, mode config.RouteMode, domain string) (*Rule, apperr.Err) {
	return NewRuleForOutbound(kind, mode, "", domain)
}

func NewRuleForOutbound(kind RuleType, mode config.RouteMode, outbound, domain string) (*Rule, apperr.Err) {
	if err := mode.Validate(); err != nil {
		return nil, apperr.NewValidationErr("DNSRule_InvalidRouteMode", err.Error())
	}
	domain = strings.ToLower(strings.TrimSpace(domain))
	target, err := config.NewRuleTarget(mode, outbound)
	if err != nil {
		return nil, apperr.NewValidationErr("DNSRule_InvalidOutbound", err.Error())
	}
	rule := &Rule{kind: kind, mode: mode, target: target, domain: domain}
	if err := rule.validate(); err != nil {
		return nil, err
	}

	return rule, nil
}

var domainRegex = regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

func (r *Rule) validate() apperr.Err {
	if !r.kind.isValid() {
		return errInvalidRuleType
	}

	if err := r.mode.Validate(); err != nil {
		return apperr.NewValidationErr("DNSRule_InvalidRouteMode", err.Error())
	}

	if r.domain == "" {
		return errEmptyDomain
	}

	if strings.ContainsAny(r.domain, " \t\n\r") {
		return errDomainHasSpaces(r.domain)
	}

	if r.kind == Domain && !domainRegex.MatchString(r.domain) {
		return errInvalidDomain(r.domain)
	}

	if r.kind == Regex {
		_, err := regexp.Compile(r.domain)
		if err != nil {
			return errInvalidRegexp(r.domain)
		}
	}

	return nil
}

var dnsRoute = map[config.RouteMode]string{
	config.RouteDirect: "dns-direct",
	config.RouteProxy:  "dns-remote",
	config.RouteBlock:  "dns-block",
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
	addedToRoute := addToRoute(r, c)
	config.NormalizeRouteRuleOrder(c)
	addedToDNS := reconcileDNS(r, c)
	return addedToRoute || addedToDNS
}

func addToRoute(r *Rule, c *config.Conf) bool {
	ruleSet := config.FindOrCreateRouteRule(c, r.target)
	rules := getRulesForType(r.kind, &ruleSet.Rule)
	ruleIdx := slices.IndexFunc(*rules, func(d string) bool {
		return strings.EqualFold(strings.TrimSpace(d), r.domain)
	})

	if ruleIdx == -1 {
		*rules = append(*rules, r.domain)
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
	if removedFromRoute {
		config.RemoveRouteRuleIfEmpty(c, r.target)
	}
	config.NormalizeRouteRuleOrder(c)
	removedFromDNS := reconcileDNS(r, c)
	return removedFromRoute || removedFromDNS
}

func removeFromRoute(r *Rule, c *config.Conf) bool {
	ruleSet, ok := config.FindRouteRule(c, r.target)
	if !ok {
		return false
	}
	rules := getRulesForType(r.kind, &ruleSet.Rule)
	ruleIdx := slices.IndexFunc(*rules, func(d string) bool {
		return strings.EqualFold(strings.TrimSpace(d), r.domain)
	})

	if ruleIdx == -1 {
		return false
	}

	*rules = slices.Delete(*rules, ruleIdx, ruleIdx+1)
	return true
}

func reconcileDNS(r *Rule, c *config.Conf) bool {
	wantServer, wantReject := effectiveDNSDestination(r, c)
	changed := false
	for index := range c.DNS.Rules {
		dnsRule := &c.DNS.Rules[index]
		if !isManagedDNSRule(*dnsRule) {
			continue
		}
		values := getRulesForType(r.kind, &dnsRule.Rule)
		before := len(*values)
		*values = slices.DeleteFunc(*values, func(value string) bool {
			return strings.EqualFold(strings.TrimSpace(value), r.domain)
		})
		changed = changed || len(*values) != before
	}

	if wantServer == "" && !wantReject {
		return changed
	}

	dnsRule := findOrCreateDNSRule(c, wantServer, wantReject)
	values := getRulesForType(r.kind, &dnsRule.Rule)
	if !slices.ContainsFunc(*values, func(value string) bool {
		return strings.EqualFold(strings.TrimSpace(value), r.domain)
	}) {
		*values = append(*values, r.domain)
		changed = true
	}
	return changed
}

func effectiveDNSDestination(r *Rule, c *config.Conf) (server string, reject bool) {
	valueType := configRuleValueType(r.kind)
	assignments := config.FindRuleAssignments(c, valueType, r.domain)
	if len(assignments) == 0 {
		return "", false
	}
	assignment := assignments[0]
	if assignment.RouteMode == config.RouteBlock {
		return "", true
	}
	if assignment.RouteMode == config.RouteDirect {
		return dnsRoute[config.RouteDirect], false
	}
	return dnsRoute[config.RouteProxy], false
}

func isManagedDNSRule(rule config.DNSRule) bool {
	return rule.Action == "reject" || rule.Server == "dns-block" ||
		((rule.Action == "" || rule.Action == "route") &&
			(rule.Server == dnsRoute[config.RouteDirect] || rule.Server == dnsRoute[config.RouteProxy]))
}

func findOrCreateDNSRule(c *config.Conf, server string, reject bool) *config.DNSRule {
	index := slices.IndexFunc(c.DNS.Rules, func(rule config.DNSRule) bool {
		if reject {
			return rule.Action == "reject" || rule.Server == "dns-block"
		}
		return (rule.Action == "" || rule.Action == "route") && rule.Server == server
	})
	if index == -1 {
		rule := config.DNSRule{Rule: config.Rule{Action: "route", Server: server}}
		if reject {
			rule.Action = "reject"
			rule.Server = ""
		}
		c.DNS.Rules = append(c.DNS.Rules, rule)
		index = len(c.DNS.Rules) - 1
	}
	return &c.DNS.Rules[index]
}

func configRuleValueType(kind RuleType) config.RuleValueType {
	switch kind {
	case Suffix:
		return config.RuleDomainSuffix
	case Keyword:
		return config.RuleDomainKeyword
	case Domain:
		return config.RuleDomain
	case Regex:
		return config.RuleDomainRegex
	default:
		return ""
	}
}

func FindAssignments(r *Rule) ([]config.RuleAssignment, apperr.Err) {
	c, err := config.Load()
	if err != nil {
		return nil, err
	}
	return config.FindRuleAssignments(c.Conf, configRuleValueType(r.kind), r.domain), nil
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
