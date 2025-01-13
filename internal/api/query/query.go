package query

import (
	"fmt"
	"net/url"
	"strconv"
)

func GetBool(q url.Values, key string, fallback bool) (bool, error) {
	if _, ok := q[key]; !ok {
		return fallback, nil
	}

	val := q.Get(key)
	if val == "" {
		return true, nil
	}

	result, err := strconv.ParseBool(val)
	if err != nil {
		return fallback, fmt.Errorf("invalid value '%s' for query param '%s', expected a boolean (true, false)", val, key)
	}

	return result, nil
}

func GetString(q url.Values, key string, fallback string) string {
	if _, ok := q[key]; !ok {
		return fallback
	}

	return q.Get(key)
}
