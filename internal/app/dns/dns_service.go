package dns

import (
	"fmt"
	"strings"

	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/config"
	"github.com/traf72/singbox-api/internal/config/dns"
)

var (
	errEmptyRule = apperr.NewValidationErr("DNSRule_Empty", "DNS rule is empty")
	errEmptyType = apperr.NewValidationErr("DNSRule_EmptyType", "DNS rule type is empty")
)

func errUnknownType(t string) *apperr.Err {
	return apperr.NewValidationErr("DNSRule_UnknownType", fmt.Sprintf("DNS rule type '%s' is unknown", t))
}

func errTooManyParts(t string) *apperr.Err {
	return apperr.NewValidationErr("DNSRule_TooManyParts", fmt.Sprintf("DNS rule '%s' has too many parts", t))
}

type Rule struct {
	RouteMode string `json:"routeMode"`
	Domain    string `json:"domain"`
}

func (r *Rule) toConfigRule() (*dns.Rule, *apperr.Err) {
	if strings.TrimSpace(r.Domain) == "" {
		return nil, errEmptyRule
	}

	routeMode, err := config.RouteModeFromString(r.RouteMode)
	if err != nil {
		return nil, apperr.NewValidationErr("DNSRule_InvalidRouteMode", err.Error())
	}

	parts := strings.Split(r.Domain, ":")
	if len(parts) > 2 {
		return nil, errTooManyParts(r.Domain)
	}

	var ruleType dns.RuleType
	var domain string
	var appErr *apperr.Err

	if len(parts) == 1 {
		ruleType = dns.Domain
		domain = parts[0]
	} else {
		ruleType, appErr = parseType(parts[0])
		if appErr != nil {
			return nil, appErr
		}

		domain = parts[1]
	}

	rule, appErr := dns.NewRule(ruleType, routeMode, domain)
	if appErr != nil {
		return nil, appErr
	}

	return rule, nil
}

func parseType(input string) (dns.RuleType, *apperr.Err) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return -1, errEmptyType
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
		return -1, errUnknownType(input)
	}
}

func AddRule(r *Rule) *apperr.Err {
	rule, err := r.toConfigRule()
	if err != nil {
		return err
	}

	if err = dns.AddRule(rule); err != nil {
		return err
	}

	return nil
}

func RemoveRule(r *Rule) *apperr.Err {
	rule, err := r.toConfigRule()
	if err != nil {
		return err
	}

	if err = dns.RemoveRule(rule); err != nil {
		return err
	}

	return nil
}
