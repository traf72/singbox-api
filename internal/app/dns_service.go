package app

import (
	"fmt"
	"strings"

	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/singbox"
	"github.com/traf72/singbox-api/internal/singbox/config"
	"github.com/traf72/singbox-api/internal/singbox/config/dns"
)

var (
	errDNSEmptyRule = apperr.NewValidationErr("DNSRule_Empty", "DNS rule is empty")
	errDNSEmptyType = apperr.NewValidationErr("DNSRule_EmptyType", "DNS rule type is empty")
)

func errDNSUnknownType(t string) apperr.Err {
	return apperr.NewValidationErr("DNSRule_UnknownType", fmt.Sprintf("DNS rule type '%s' is unknown", t))
}

func errDNSTooManyParts(t string) apperr.Err {
	return apperr.NewValidationErr("DNSRule_TooManyParts", fmt.Sprintf("DNS rule '%s' has too many parts", t))
}

type DNSRule struct {
	RouteMode string `json:"routeMode"`
	Outbound  string `json:"outbound,omitempty"`
	Domain    string `json:"domain"`
}

func (r *DNSRule) toConfigRule() (*dns.Rule, apperr.Err) {
	routeMode, err := config.RouteModeFromString(r.RouteMode)
	if err != nil {
		return nil, apperr.NewValidationErr("DNSRule_InvalidRouteMode", err.Error())
	}
	ruleType, domain, appErr := parseDNSRuleDomain(r.Domain)
	if appErr != nil {
		return nil, appErr
	}

	rule, appErr := dns.NewRuleForOutbound(ruleType, routeMode, r.Outbound, domain)
	if appErr != nil {
		return nil, appErr
	}

	return rule, nil
}

func parseDNSRuleDomain(input string) (dns.RuleType, string, apperr.Err) {
	if strings.TrimSpace(input) == "" {
		return -1, "", errDNSEmptyRule
	}

	parts := strings.Split(input, ":")
	if len(parts) > 2 {
		return -1, "", errDNSTooManyParts(input)
	}

	var ruleType dns.RuleType
	var domain string
	var appErr apperr.Err

	if len(parts) == 1 {
		ruleType = dns.Domain
		domain = parts[0]
	} else {
		ruleType, appErr = parseDNSRuleType(parts[0])
		if appErr != nil {
			return -1, "", appErr
		}

		domain = parts[1]
	}

	return ruleType, domain, nil
}

func parseDNSRuleType(input string) (dns.RuleType, apperr.Err) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return -1, errDNSEmptyType
	}

	switch strings.ToLower(trimmed) {
	case "full":
		return dns.Domain, nil
	case "keyword":
		return dns.Keyword, nil
	case "domain":
		return dns.Suffix, nil
	case "regexp":
		return dns.Regex, nil
	default:
		return -1, errDNSUnknownType(input)
	}
}

func AddDNSRule(r *DNSRule, restart bool) apperr.Err {
	rule, err := r.toConfigRule()
	if err != nil {
		return err
	}

	if err = dns.AddRule(rule); err != nil {
		return err
	}

	if restart {
		if err = singbox.Restart(); err != nil {
			return err
		}
	}

	return nil
}

func RemoveDNSRule(r *DNSRule, restart bool) apperr.Err {
	rule, err := r.toConfigRule()
	if err != nil {
		return err
	}

	if err = dns.RemoveRule(rule); err != nil {
		return err
	}

	if restart {
		if err = singbox.Restart(); err != nil {
			return err
		}
	}

	return nil
}
