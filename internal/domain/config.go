package domain

import (
	"os"

	"github.com/traf72/singbox-api/internal/err"
)

func AddTemplate(t *Template) *err.AppErr {
	// where to get the file path
	return nil
}

var errEmptyPath = err.NewFatalErr("EmptyConfigPath", "Path to the configuration file is not specified")

func open() (*os.File, error) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		return nil, errEmptyPath
	}

	// file, err := os.ReadFile(path)
	// if err != nil {
	// 	return nil, err
	// }

	return nil, nil
}
