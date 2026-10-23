package config

import (
	"fmt"
	"slices"
	"strings"

	"github.com/traf72/singbox-api/internal/apperr"
)

type RuleTarget struct {
	mode     RouteMode
	outbound string
}

type RuleValueType string

const (
	RuleDomain        RuleValueType = "domain"
	RuleDomainSuffix  RuleValueType = "domain_suffix"
	RuleDomainKeyword RuleValueType = "domain_keyword"
	RuleDomainRegex   RuleValueType = "domain_regex"
	RuleIPCIDR        RuleValueType = "ip_cidr"
)

type RuleAssignment struct {
	RouteMode RouteMode `json:"routeMode"`
	Outbound  string    `json:"outbound,omitempty"`
	Order     int       `json:"order"`
}

func NewRuleTarget(mode RouteMode, outbound string) (*RuleTarget, apperr.Err) {
	if err := mode.Validate(); err != nil {
		return nil, apperr.NewValidationErr("RouteRule_InvalidRouteMode", err.Error())
	}

	outbound = strings.TrimSpace(outbound)
	if mode != RouteProxy && outbound != "" {
		return nil, apperr.NewValidationErr(
			"RouteRule_UnexpectedOutbound",
			"outbound can only be specified when routeMode is 'proxy'",
		)
	}
	if mode == RouteProxy && outbound == "" {
		outbound = proxySelectorTag
	}

	return &RuleTarget{mode: mode, outbound: outbound}, nil
}

func (t *RuleTarget) Mode() RouteMode {
	return t.mode
}

func (t *RuleTarget) Outbound() string {
	if t.mode == RouteBlock {
		return ""
	}
	if t.mode == RouteDirect {
		return string(RouteDirect)
	}
	return t.outbound
}

func (t *RuleTarget) ValidateForAdd(conf *Conf) apperr.Err {
	if t.mode == RouteBlock {
		return nil
	}

	tag := t.Outbound()
	if findOutbound(conf, tag) == nil {
		return invalidRuleOutbound(tag, "does not exist")
	}
	if t.mode != RouteProxy || tag == proxySelectorTag {
		return nil
	}
	if IsDirectOutbound(conf, tag) {
		return invalidRuleOutbound(tag, "is direct; use routeMode 'direct' instead")
	}

	selector := findProxySelector(conf)
	if selector == nil {
		return invalidRuleOutbound(tag, fmt.Sprintf("cannot be selected because selector '%s' does not exist", proxySelectorTag))
	}
	if !slices.Contains(selector.Outbounds, tag) {
		return invalidRuleOutbound(tag, fmt.Sprintf("is not a member of selector '%s'", proxySelectorTag))
	}

	return nil
}

func (t *RuleTarget) ValidateForRemoval(conf *Conf) apperr.Err {
	if t.matchesAnyRouteRule(conf) {
		return nil
	}
	return t.ValidateForAdd(conf)
}

func (t *RuleTarget) Matches(rule RouteRule) bool {
	if t.mode == RouteBlock {
		return rule.Action == "reject"
	}
	return rule.Outbound == t.Outbound()
}

func (t *RuleTarget) matchesAnyRouteRule(conf *Conf) bool {
	return slices.ContainsFunc(conf.Route.Rules, func(rule RouteRule) bool {
		return t.Matches(rule) && hasRouteMatchers(rule)
	})
}

func invalidRuleOutbound(tag, reason string) apperr.Err {
	return apperr.NewValidationErr(
		"RouteRule_InvalidOutbound",
		fmt.Sprintf("outbound tag '%s' %s", tag, reason),
	)
}

func FindRouteRule(conf *Conf, target *RuleTarget) (*RouteRule, bool) {
	index := slices.IndexFunc(conf.Route.Rules, func(rule RouteRule) bool {
		return target.Matches(rule) && hasRouteMatchers(rule)
	})
	if index == -1 {
		return nil, false
	}
	return &conf.Route.Rules[index], true
}

func FindOrCreateRouteRule(conf *Conf, target *RuleTarget) *RouteRule {
	if rule, ok := FindRouteRule(conf, target); ok {
		return rule
	}

	rule := RouteRule{Rule: Rule{}}
	if target.mode == RouteBlock {
		rule.Action = "reject"
	} else {
		rule.Outbound = target.Outbound()
	}
	conf.Route.Rules = append(conf.Route.Rules, rule)
	return &conf.Route.Rules[len(conf.Route.Rules)-1]
}

func NormalizeRouteRuleOrder(conf *Conf) {
	preprocessing := make([]RouteRule, 0, len(conf.Route.Rules))
	conditional := make([]RouteRule, 0, len(conf.Route.Rules))
	proxy := make([]RouteRule, 0, 1)
	terminal := make([]RouteRule, 0, len(conf.Route.Rules))

	for _, rule := range conf.Route.Rules {
		if !isFinalRouteRule(rule) {
			preprocessing = append(preprocessing, rule)
			continue
		}
		if !hasRouteMatchers(rule) {
			terminal = append(terminal, rule)
			continue
		}
		if rule.Outbound == proxySelectorTag {
			proxy = append(proxy, rule)
			continue
		}
		conditional = append(conditional, rule)
	}

	selectorRanks := make(map[string]int)
	if selector := findProxySelector(conf); selector != nil {
		for index, tag := range selector.Outbounds {
			selectorRanks[tag] = index
		}
	}
	slices.SortStableFunc(conditional, func(a, b RouteRule) int {
		return routeRuleRank(a, selectorRanks) - routeRuleRank(b, selectorRanks)
	})

	rules := make([]RouteRule, 0, len(conf.Route.Rules))
	rules = append(rules, preprocessing...)
	rules = append(rules, conditional...)
	rules = append(rules, proxy...)
	rules = append(rules, terminal...)
	conf.Route.Rules = rules
}

func routeRuleRank(rule RouteRule, selectorRanks map[string]int) int {
	if rule.Action == "reject" {
		return 0
	}
	if rule.Outbound == string(RouteDirect) {
		return 1
	}
	if rank, ok := selectorRanks[rule.Outbound]; ok {
		return 2 + rank
	}
	return 2 + len(selectorRanks)
}

func isFinalRouteRule(rule RouteRule) bool {
	return rule.Outbound != "" || rule.Action == "route" || rule.Action == "reject"
}

func hasRouteMatchers(rule RouteRule) bool {
	return len(rule.Domain) > 0 || len(rule.DomainSuffix) > 0 || len(rule.DomainKeyword) > 0 ||
		len(rule.DomainRegex) > 0 || len(rule.IP_CIDR) > 0 || len(rule.Inbound) > 0 || rule.Protocol != ""
}

func RemoveRouteRuleIfEmpty(conf *Conf, target *RuleTarget) {
	index := slices.IndexFunc(conf.Route.Rules, func(rule RouteRule) bool {
		return target.Matches(rule) && !hasRouteMatchers(rule)
	})
	if index != -1 {
		conf.Route.Rules = slices.Delete(conf.Route.Rules, index, index+1)
	}
}

func IsDirectOutbound(conf *Conf, tag string) bool {
	item := findOutbound(conf, tag)
	return item != nil && item.Type == "direct"
}

func FindRuleAssignments(conf *Conf, valueType RuleValueType, value string) []RuleAssignment {
	assignments := make([]RuleAssignment, 0)
	for index, rule := range conf.Route.Rules {
		if !ruleContainsValue(rule, valueType, value) {
			continue
		}
		if rule.Action == "reject" {
			assignments = append(assignments, RuleAssignment{RouteMode: RouteBlock, Order: index})
			continue
		}
		if rule.Outbound == "" {
			continue
		}
		mode := RouteProxy
		if IsDirectOutbound(conf, rule.Outbound) {
			mode = RouteDirect
		}
		assignments = append(assignments, RuleAssignment{
			RouteMode: mode,
			Outbound:  rule.Outbound,
			Order:     index,
		})
	}
	return assignments
}

func ruleContainsValue(rule RouteRule, valueType RuleValueType, value string) bool {
	var values []string
	switch valueType {
	case RuleDomain:
		values = rule.Domain
	case RuleDomainSuffix:
		values = rule.DomainSuffix
	case RuleDomainKeyword:
		values = rule.DomainKeyword
	case RuleDomainRegex:
		values = rule.DomainRegex
	case RuleIPCIDR:
		values = rule.IP_CIDR
	default:
		return false
	}
	return slices.ContainsFunc(values, func(candidate string) bool {
		return strings.EqualFold(strings.TrimSpace(candidate), value)
	})
}
