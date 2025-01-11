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

func EnableLog(restart bool, truncate bool) apperr.Err {
	return setLogEnabled(true, restart, truncate)
}

func DisableLog(restart bool, truncate bool) apperr.Err {
	return setLogEnabled(false, restart, truncate)
}

func setLogEnabled(enable bool, restart bool, truncate bool) apperr.Err {
	if truncate {
		if err := TruncateLog(); err != nil {
			return err
		}
	}

	var f func() apperr.Err
	if enable {
		f = config.EnableLog
	} else {
		f = config.DisableLog
	}

	if err := f(); err != nil {
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
