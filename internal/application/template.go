package application

import (
	"fmt"
	"strings"

	"github.com/traf72/singbox-api/internal/domain"
	"github.com/traf72/singbox-api/internal/err"
)

func AddDomain(input string) *err.AppErr {
	template, err := parse(input)
	if err != nil {
		return err
	}

	if err = domain.AddTemplate(template); err != nil {
		return err
	}

	return nil
}

var errEmptyTemplate = err.NewValidationErr("EmptyTemplate", "template is empty")

func errTooManyParts(t string) *err.AppErr {
	return err.NewValidationErr("TemplateHasTooManyParts", fmt.Sprintf("template '%s' has too many parts", t))
}

func parse(input string) (*domain.Template, *err.AppErr) {
	if strings.TrimSpace(input) == "" {
		return nil, errEmptyTemplate
	}

	parts := strings.Split(input, ":")
	if len(parts) > 2 {
		return nil, errTooManyParts(input)
	}

	var kind domain.TemplateKind
	var text string

	if len(parts) == 1 {
		kind = domain.Domain
		text = parts[0]
	} else {
		kind = parseKind(parts[0])
		text = parts[1]
	}

	template, err := domain.NewTemplate(kind, text)
	if err != nil {
		return nil, err
	}

	return template, nil
}

func parseKind(input string) domain.TemplateKind {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return -1
	}

	switch strings.ToLower(trimmed) {
	case "keyword":
		return domain.Keyword
	case "domain":
		return domain.Suffix
	default:
		return -1
	}
}
