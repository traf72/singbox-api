package logs

import (
	"io"
	"os"

	"github.com/traf72/singbox-api/internal/apperr"
)

var errEmptyPath = apperr.NewFatalErr("Log_EmptyPath", "path to the log file is not specified")
var ErrFileNotFound = apperr.NewFatalErr("Log_NotFound", "log file not found")

func Open() (io.ReadCloser, apperr.Err) {
	path, appErr := getPath()
	if appErr != nil {
		return nil, appErr
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, ErrFileNotFound
	} else if err != nil {
		return nil, apperr.NewFatalErr("Log_StatReadError", err.Error())
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, apperr.NewFatalErr("Log_OpenError", err.Error())
	}

	return file, nil
}

func Truncate() apperr.Err {
	path, appErr := getPath()
	if appErr != nil {
		return appErr
	}

	if err := os.Truncate(path, 0); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return apperr.NewFatalErr("Log_TruncateError", err.Error())
	}

	return nil
}

func getPath() (string, apperr.Err) {
	path := os.Getenv("LOG_PATH")
	if path == "" {
		return "", errEmptyPath
	}

	return path, nil
}
