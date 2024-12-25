package domains

import (
	"regexp"
	"strings"
)

type kind int

const (
	Suffix kind = iota
	Keyword
	Domain
)

type template struct {
	kind kind
	text string
}

func parse(input string) (template, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return template{}, errEmptyTemplate
	}

	parts := strings.Split(trimmed, ":")
	if len(parts) > 2 {
		return template{}, tooManyParts(input)
	}

	if len(parts) == 1 {
		d, err := parseDomain(parts[0])
		if err != nil {
			return template{}, err
		}

		return template{kind: Domain, text: d}, nil
	}

	t := strings.TrimSpace(parts[1])
	if t == "" {
		return template{}, errEmptyTemplate
	}

	kind, err := parseType(parts[0])
	if err != nil {
		return template{}, err
	}

	return template{kind: kind, text: strings.ToLower(t)}, nil
}

var domainRegex = regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

func parseDomain(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", errEmptyTemplate
	}

	if !domainRegex.MatchString(trimmed) {
		return "", invalidDomain(input)
	}

	return strings.ToLower(trimmed), nil
}

func parseType(input string) (kind, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return -1, errEmptyTemplateType
	}

	switch strings.ToLower(trimmed) {
	case "keyword":
		return Keyword, nil
	case "domain":
		return Suffix, nil
	default:
		return -1, invalidTemplateType(input)
	}
}
