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
)

type RuleAssignmentsResult struct {
	Kind        string                  `json:"kind"`
	Value       string                  `json:"value"`
	Assignments []config.RuleAssignment `json:"assignments"`
}

func GetRuleAssignments(domainValue, ipValue string) (*RuleAssignmentsResult, apperr.Err) {
	domainValue = strings.TrimSpace(domainValue)
	ipValue = strings.TrimSpace(ipValue)
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
	assignments, err := ip.FindAssignments(rule)
	if err != nil {
		return nil, err
	}
	return &RuleAssignmentsResult{Kind: "ip", Value: ipValue, Assignments: assignments}, nil
}
