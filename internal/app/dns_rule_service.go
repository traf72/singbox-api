package app

import (
	"fmt"
	"strings"

	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/core"
)

var (
	errEmptyRule      = apperr.NewValidationErr("DNSRule_Empty", "DNS rule is empty")
	errEmptyType      = apperr.NewValidationErr("DNSRule_EmptyType", "DNS rule type is empty")
	errEmptyRouteMode = apperr.NewValidationErr("DNSRule_EmptyRouteMode", "Route mode is empty")
)

func errUnknownType(t string) *apperr.Err {
	return apperr.NewValidationErr("DNSRule_UnknownType", fmt.Sprintf("unknown DNS rule type '%s'", t))
}

func errUnknownRouteMode(m string) *apperr.Err {
	return apperr.NewValidationErr("DNSRule_UnknownRouteMode", fmt.Sprintf("unknown route mode '%s'", m))
}

func errTooManyParts(t string) *apperr.Err {
	return apperr.NewValidationErr("DNSRule_TooManyParts", fmt.Sprintf("DNS rule '%s' has too many parts", t))
}

type DNSRuleDTO struct {
	RouteMode string `json:"routeMode"`
	Domain    string `json:"domain"`
}

func (r *DNSRuleDTO) toDNSRule() (*core.DNSRule, *apperr.Err) {
	if strings.TrimSpace(r.Domain) == "" {
		return nil, errEmptyRule
	}

	routeMode, err := parseRouteMode(r.RouteMode)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(r.Domain, ":")
	if len(parts) > 2 {
		return nil, errTooManyParts(r.Domain)
	}

	var dnsRuleType core.DNSRuleType
	var domain string

	if len(parts) == 1 {
		dnsRuleType = core.DNSRuleDomain
		domain = parts[0]
	} else {
		dnsRuleType, err = parseDNSType(parts[0])
		if err != nil {
			return nil, err
		}

		domain = parts[1]
	}

	dnsRule, err := core.NewDNSRule(dnsRuleType, routeMode, domain)
	if err != nil {
		return nil, err
	}

	return dnsRule, nil
}

func parseDNSType(input string) (core.DNSRuleType, *apperr.Err) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return -1, errEmptyType
	}

	switch strings.ToLower(trimmed) {
	case "keyword":
		return core.DNSRuleKeyword, nil
	case "domain":
		return core.DNSRuleSuffix, nil
	default:
		return -1, errUnknownType(input)
	}
}

func parseRouteMode(input string) (core.RouteMode, *apperr.Err) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", errEmptyRouteMode
	}

	switch strings.ToLower(trimmed) {
	case "proxy":
		return core.RouteProxy, nil
	case "block":
		return core.RouteBlock, nil
	case "direct":
		return core.RouteDirect, nil
	default:
		return "", errUnknownRouteMode(input)
	}
}

func AddDNSRule(r *DNSRuleDTO) *apperr.Err {
	dnsRule, err := r.toDNSRule()
	if err != nil {
		return err
	}

	if err = core.AddDNSRule(dnsRule); err != nil {
		return err
	}

	return nil
}
