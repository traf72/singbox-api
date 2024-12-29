package utils

import (
	"os"
	"strconv"
)

func GetEnv(name string, defaultVal string) string {
	val := os.Getenv(name)
	if val == "" {
		val = defaultVal
	}

	return val
}

func GetEnvBool(name string, defaultVal bool) (bool, error) {
	strVal := GetEnv(name, strconv.FormatBool(defaultVal))
	val, err := strconv.ParseBool(strVal)
	if err != nil {
		return false, err
	}

	return val, nil
}
