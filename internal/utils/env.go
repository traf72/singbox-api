package utils

import "os"

func GetEnv(name string, defaultVal string) string {
	val := os.Getenv(name)
	if val == "" {
		val = defaultVal
	}

	return val
}
