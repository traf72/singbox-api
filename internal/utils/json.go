package utils

import (
	"encoding/json"
	"fmt"
	"io"
)

type JSONOptions struct {
	Indent     string
	Prefix     string
	EscapeHTML bool
}

func FromJSON[T any](r io.Reader, target *T) error {
	d := json.NewDecoder(r)
	if err := d.Decode(target); err != nil {
		return fmt.Errorf("failed to parse json: %w", err)
	}

	return nil
}

func ToJSON(w io.Writer, source any, options *JSONOptions) error {
	e := json.NewEncoder(w)
	if options != nil {
		e.SetIndent(options.Prefix, options.Indent)
		e.SetEscapeHTML(options.EscapeHTML)
	}

	if err := e.Encode(source); err != nil {
		return fmt.Errorf("failed to serialize to json: %w", err)
	}

	return nil
}
