package logs

import (
	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/config"
	"github.com/traf72/singbox-api/internal/config/logs"
	"github.com/traf72/singbox-api/internal/config/singbox"
)

func SetEnabled(enable bool, restart bool, truncate bool) apperr.Err {
	c, err := config.Load()
	if err != nil {
		return err
	}

	if truncate {
		if err := logs.Truncate(); err != nil {
			return err
		}
	}

	c.Conf.Log.Disabled = !enable
	if err := config.Save(c); err != nil {
		return err
	}

	if restart {
		if err := singbox.Restart(); err != nil {
			return err
		}
	}

	return nil
}
