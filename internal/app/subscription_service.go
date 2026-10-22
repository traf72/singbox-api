package app

import (
	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/singbox"
	"github.com/traf72/singbox-api/internal/singbox/config"
)

type SubscriptionSettings = config.SubscriptionSettings
type SubscriptionUpdate = config.SubscriptionUpdate
type SubscriptionUpdateResult = config.SubscriptionUpdateResult

func GetSubscriptionSettings() (*SubscriptionSettings, apperr.Err) {
	return config.GetSubscriptionSettings()
}

func SetSubscriptionSettings(settings *SubscriptionSettings) apperr.Err {
	return config.SetSubscriptionSettings(settings)
}

func UpdateSubscription(request *SubscriptionUpdate, restart bool) (*SubscriptionUpdateResult, apperr.Err) {
	result, err := config.UpdateSubscription(request)
	if err != nil {
		return nil, err
	}

	if restart {
		if err := singbox.Restart(); err != nil {
			return nil, err
		}
	}

	return result, nil
}
