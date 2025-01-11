package config

import (
	"github.com/traf72/singbox-api/internal/apperr"
)

func EnableLog() apperr.Err {
	return setLogEnabled(true)
}

func DisableLog() apperr.Err {
	return setLogEnabled(false)
}

func setLogEnabled(enable bool) apperr.Err {
	c, err := Load()
	if err != nil {
		return err
	}

	c.Conf.Log.Disabled = !enable
	if err := Save(c); err != nil {
		return err
	}

	return nil
}
