package app

import (
	"io"

	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/singbox"
	"github.com/traf72/singbox-api/internal/singbox/config"
)

func GetLog() (io.ReadCloser, apperr.Err) {
	return singbox.GetLog()
}

func EnableLog(restart bool, truncate bool, level string) apperr.Err {
	if truncate {
		if err := TruncateLog(); err != nil {
			return err
		}
	}

	if err := config.EnableLog(config.LogLevel(level)); err != nil {
		return err
	}

	if restart {
		if err := singbox.Restart(); err != nil {
			return err
		}
	}

	return nil
}

func DisableLog(restart bool) apperr.Err {
	if err := config.DisableLog(); err != nil {
		return err
	}

	if restart {
		if err := singbox.Restart(); err != nil {
			return err
		}
	}

	return nil
}

func TruncateLog() apperr.Err {
	if err := singbox.TruncateLog(); err != nil && err != singbox.ErrLogNotFound {
		return err
	}

	return nil
}

func SetLogLevel(l string, restart bool) apperr.Err {
	if err := config.SetLogLevel(config.LogLevel(l)); err != nil {
		return err
	}

	if restart {
		if err := singbox.Restart(); err != nil {
			return err
		}
	}

	return nil
}
