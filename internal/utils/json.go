package utils

import (
	"encoding/json"
	"fmt"
	"io"
)

func ParseJson[T any](r io.Reader, target *T) error {
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("failed to parse json: %w", err)
	}

	return nil
}
