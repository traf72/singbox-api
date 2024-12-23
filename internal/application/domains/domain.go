package domains

import (
	"strings"
)

type domainType int

const (
	Suffix domainType = iota
	Keyword
	Strict
	Invalid = -1
)

type domain struct {
	kind domainType
	name string
}

func parse(input string) (domain, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return domain{}, domainIsEmpty
	}

	parts := strings.Split(trimmed, ":")
	if len(parts) > 2 {
		return domain{}, tooManyParts(input)
	}

	if len(parts) == 1 {
		d := parts[0]
		if isValid(d) {
			return domain{kind: Strict, name: strings.ToLower(d)}, nil
		}

		return domain{}, invalidDomain(input)
	}

	dt := parts[0]
	d := parts[1]

	if !isValid(d) {
		return domain{}, invalidDomain(input)
	}

	if !isValid(d) {
		return domain{}, invalidDomain(input)
	}
}

func parseType(input string) (domainType, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return Invalid, domainTypeIsEmpty
	}

	switch strings.ToLower(trimmed) {
	case "keyword":
		return Keyword, nil
	case "domain":
		return Suffix, nil
	default:
		return Invalid, invalidDomainType(input)
	}
}
