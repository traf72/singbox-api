package app

import (
	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/singbox/config"
)

func GetConfig() (*config.Conf, apperr.Err) {
	c, err := config.Load()
	if err != nil {
		return nil, err
	}

	return c.Conf, nil
}
