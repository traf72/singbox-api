package application

import (
	"fmt"
	"strings"

	"github.com/traf72/singbox-api/internal/config"
	"github.com/traf72/singbox-api/internal/err"
)

func AddDnsRule(input string) *err.AppErr {
	dnsRule, err := parse(input)
	if err != nil {
		return err
	}

	if err = config.AddDnsRule(dnsRule); err != nil {
		return err
	}

	return nil
}

var errEmptyRule = err.NewValidationErr("EmptyDnsRule", "dns rule is empty")

func errTooManyParts(t string) *err.AppErr {
	return err.NewValidationErr("DnsRuleHasTooManyParts", fmt.Sprintf("dns rule '%s' has too many parts", t))
}

func parse(input string) (*config.DnsRule, *err.AppErr) {
	if strings.TrimSpace(input) == "" {
		return nil, errEmptyRule
	}

	parts := strings.Split(input, ":")
	if len(parts) > 2 {
		return nil, errTooManyParts(input)
	}

	var kind config.DnsRuleType
	var text string

	if len(parts) == 1 {
		kind = config.Domain
		text = parts[0]
	} else {
		kind = parseKind(parts[0])
		text = parts[1]
	}

	template, err := config.NewDnsRule(kind, text)
	if err != nil {
		return nil, err
	}

	return template, nil
}

func parseKind(input string) config.DnsRuleType {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return -1
	}

	switch strings.ToLower(trimmed) {
	case "keyword":
		return config.Keyword
	case "domain":
		return config.Suffix
	default:
		return -1
	}
}
