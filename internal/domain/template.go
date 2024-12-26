package domain

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/traf72/singbox-api/internal/err"
)

var errEmptyTemplate = err.NewValidationErr("EmptyTemplate", "template is empty")

func errInvalidKind(k TemplateKind) *err.AppErr {
	return err.NewValidationErr("InvalidTemplateKind", fmt.Sprintf("kind '%d' is invalid", k))
}

func errTemplateHasSpaces(t string) *err.AppErr {
	return err.NewValidationErr("TemplateHasSpaces", fmt.Sprintf("template '%s' has spaces", t))
}

func errInvalidDomain(d string) *err.AppErr {
	return err.NewValidationErr("InvalidDomain", fmt.Sprintf("domain '%s' is invalid", d))
}

type TemplateKind int

const (
	Suffix TemplateKind = iota
	Keyword
	Domain
)

func (k TemplateKind) String() string {
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

func (k TemplateKind) isValid() bool {
	switch k {
	case Suffix, Keyword, Domain:
		return true
	default:
		return false
	}
}

type Template struct {
	kind TemplateKind
	text string
}

var domainRegex = regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

func (t *Template) validate() *err.AppErr {
	if !t.kind.isValid() {
		return errInvalidKind(t.kind)
	}

	if t.text == "" {
		return errEmptyTemplate
	}

	if strings.ContainsAny(t.text, " \t\n\r") {
		return errTemplateHasSpaces(t.text)
	}

	if t.kind == Domain && !domainRegex.MatchString(t.text) {
		return errInvalidDomain(t.text)
	}

	return nil
}

func NewTemplate(kind TemplateKind, text string) (*Template, *err.AppErr) {
	text = strings.ToLower(strings.TrimSpace(text))
	t := &Template{kind: kind, text: text}
	if err := t.validate(); err != nil {
		return nil, err
	}

	return t, nil
}
