package app

import (
	"fmt"
	"strings"

	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/core"
)

func AddDnsRule(input string) *apperr.Err {
	dnsRule, err := parse(input)
	if err != nil {
		return err
	}

	if err = core.AddDnsRule(dnsRule); err != nil {
		return err
	}

	return nil
}

var (
	errEmptyRule        = apperr.NewValidationErr("EmptyDnsRule", "DNS rule is empty")
	errEmptyDnsRuleType = apperr.NewValidationErr("EmptyDnsRuleType", "DNS rule type is empty")
)

func errUnknownDnsRuleType(t string) *apperr.Err {
	return apperr.NewValidationErr("UnknownDnsRuleType", fmt.Sprintf("unknown DNS rule type '%s'", t))
}

func errTooManyParts(t string) *apperr.Err {
	return apperr.NewValidationErr("DnsRuleHasTooManyParts", fmt.Sprintf("DNS rule '%s' has too many parts", t))
}

func parse(input string) (*core.DnsRule, *apperr.Err) {
	if strings.TrimSpace(input) == "" {
		return nil, errEmptyRule
	}

	parts := strings.Split(input, ":")
	if len(parts) > 2 {
		return nil, errTooManyParts(input)
	}

	var dnsRuleType core.DnsRuleType
	var domain string
	var err *apperr.Err

	if len(parts) == 1 {
		dnsRuleType = core.Domain
		domain = parts[0]
	} else {
		dnsRuleType, err = parseDnsType(parts[0])
		if err != nil {
			return nil, err
		}

		domain = parts[1]
	}

	dnsRule, err := core.NewDnsRule(dnsRuleType, domain)
	if err != nil {
		return nil, err
	}

	return dnsRule, nil
}

func parseDnsType(input string) (core.DnsRuleType, *apperr.Err) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return -1, errEmptyDnsRuleType
	}

	switch strings.ToLower(trimmed) {
	case "keyword":
		return core.Keyword, nil
	case "domain":
		return core.Suffix, nil
	default:
		return -1, errUnknownDnsRuleType(input)
	}
}
