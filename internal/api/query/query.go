package query

import (
	"fmt"
	"net/url"
	"strconv"
)

func GetBool(q url.Values, key string, fallback bool) (bool, error) {
	val := q.Get(key)
	if val == "" {
		return fallback, nil
	}

	result, err := strconv.ParseBool(val)
	if err != nil {
		return false, fmt.Errorf("invalid value '%s' for query param '%s', expected a boolean (true, false)", val, key)
	}

	return result, nil
}
