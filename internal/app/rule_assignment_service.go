package app

import (
	"strings"

	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/singbox/config"
	"github.com/traf72/singbox-api/internal/singbox/config/dns"
	"github.com/traf72/singbox-api/internal/singbox/config/ip"
)

var (
	errRuleLookupMissingValue  = apperr.NewValidationErr("RouteRuleLookup_MissingValue", "exactly one of domain or ip must be specified")
	errRuleLookupTooManyValues = apperr.NewValidationErr("RouteRuleLookup_TooManyValues", "domain and ip cannot be specified together")
	errRuleLookupInvalidMatch  = apperr.NewValidationErr("RouteRuleLookup_InvalidMatch", "match must be 'exact' or 'covering'")
	errRuleLookupCoveringType  = apperr.NewValidationErr("RouteRuleLookup_InvalidCoveringDomain", "covering domain lookup requires a hostname without a rule-type prefix, or with the 'full:' prefix")
)

type RuleAssignmentsResult struct {
	Kind        string                  `json:"kind"`
	Value       string                  `json:"value"`
	Assignments []config.RuleAssignment `json:"assignments"`
}

type CoveringRuleAssignmentsResult struct {
	Kind    string             `json:"kind"`
	Value   string             `json:"value"`
	Match   string             `json:"match"`
	Matches []config.RuleMatch `json:"matches"`
}

func GetRuleAssignments(domainValue, ipValue, matchMode string) (any, apperr.Err) {
	domainValue = strings.TrimSpace(domainValue)
	ipValue = strings.TrimSpace(ipValue)
	matchMode = strings.ToLower(strings.TrimSpace(matchMode))
	if matchMode == "" {
		matchMode = "exact"
	}
	if matchMode != "exact" && matchMode != "covering" {
		return nil, errRuleLookupInvalidMatch
	}
	if domainValue == "" && ipValue == "" {
		return nil, errRuleLookupMissingValue
	}
	if domainValue != "" && ipValue != "" {
		return nil, errRuleLookupTooManyValues
	}

	if domainValue != "" {
		kind, value, err := parseDNSRuleDomain(domainValue)
		if err != nil {
			return nil, err
		}
		rule, err := dns.NewRule(kind, config.RouteProxy, value)
		if err != nil {
			return nil, err
		}
		if matchMode == "covering" {
			if kind != dns.Domain {
				return nil, errRuleLookupCoveringType
			}
			conf, err := config.Load()
			if err != nil {
				return nil, err
			}
			hostname := strings.ToLower(strings.TrimSpace(value))
			return &CoveringRuleAssignmentsResult{
				Kind:    "domain",
				Value:   hostname,
				Match:   matchMode,
				Matches: config.FindCoveringDomainMatches(conf.Conf, hostname),
			}, nil
		}
		assignments, err := dns.FindAssignments(rule)
		if err != nil {
			return nil, err
		}
		return &RuleAssignmentsResult{
			Kind:        "domain",
			Value:       strings.ToLower(domainValue),
			Assignments: assignments,
		}, nil
	}

	rule, err := ip.NewRule(config.RouteProxy, ipValue)
	if err != nil {
		return nil, err
	}
	if matchMode == "covering" {
		conf, err := config.Load()
		if err != nil {
			return nil, err
		}
		return &CoveringRuleAssignmentsResult{
			Kind:    "ip",
			Value:   ipValue,
			Match:   matchMode,
			Matches: config.FindCoveringIPMatches(conf.Conf, ipValue),
		}, nil
	}
	assignments, err := ip.FindAssignments(rule)
	if err != nil {
		return nil, err
	}
	return &RuleAssignmentsResult{Kind: "ip", Value: ipValue, Assignments: assignments}, nil
}
