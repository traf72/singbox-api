package config

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/traf72/singbox-api/internal/err"
)

var (
	errEmptyDomain     = err.NewValidationErr("EmptyDomain", "domain is empty")
	errInvalidRuleType = err.NewValidationErr("InvalidRuleType", "rule type is invalid")
)

func errDomainHasSpaces(t string) *err.AppErr {
	return err.NewValidationErr("DomainHasSpaces", fmt.Sprintf("domain '%s' has spaces", t))
}

func errInvalidDomain(d string) *err.AppErr {
	return err.NewValidationErr("InvalidDomain", fmt.Sprintf("domain '%s' is invalid", d))
}

type DnsRuleType int

const (
	Suffix DnsRuleType = iota
	Keyword
	Domain
)

func (k DnsRuleType) String() string {
	switch k {
	case Suffix:
		return "Suffix"
	case Keyword:
		return "Keyword"
	case Domain:
		return "Domain"
	default:
		return "Unknown"
	}
}

func (k DnsRuleType) isValid() bool {
	switch k {
	case Suffix, Keyword, Domain:
		return true
	default:
		return false
	}
}

type DnsRule struct {
	kind   DnsRuleType
	domain string
}

func NewDnsRule(kind DnsRuleType, text string) (*DnsRule, *err.AppErr) {
	text = strings.ToLower(strings.TrimSpace(text))
	t := &DnsRule{kind: kind, domain: text}
	if err := t.validate(); err != nil {
		return nil, err
	}

	return t, nil
}

var domainRegex = regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

func (t *DnsRule) validate() *err.AppErr {
	if !t.kind.isValid() {
		return errInvalidRuleType
	}

	if t.domain == "" {
		return errEmptyDomain
	}

	if strings.ContainsAny(t.domain, " \t\n\r") {
		return errDomainHasSpaces(t.domain)
	}

	if t.kind == Domain && !domainRegex.MatchString(t.domain) {
		return errInvalidDomain(t.domain)
	}

	return nil
}

func AddDnsRule(r *DnsRule) *err.AppErr {
	return nil
}

func RemoveDnsRule(r *DnsRule) *err.AppErr {
	return nil
}
