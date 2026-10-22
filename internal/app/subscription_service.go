package app

import (
	"time"

	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/singbox"
	"github.com/traf72/singbox-api/internal/singbox/config"
)

type SubscriptionSettings = config.SubscriptionSettings
type SubscriptionUpdate = config.SubscriptionUpdate
type SubscriptionUpdateResult = config.SubscriptionUpdateResult
type BackupInfo = config.BackupInfo

type BackupCleanupResult struct {
	Removed int `json:"removed"`
}

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

func ListBackups() ([]BackupInfo, apperr.Err) {
	return config.ListBackups()
}

func RestoreBackup(name string, restart bool) apperr.Err {
	if err := config.RestoreBackup(name); err != nil {
		return err
	}
	if restart {
		return singbox.Restart()
	}

	return nil
}

func RemoveBackupsOlderThan(cutoff time.Time) (*BackupCleanupResult, apperr.Err) {
	removed, err := config.RemoveBackupsOlderThan(cutoff)
	if err != nil {
		return nil, err
	}

	return &BackupCleanupResult{Removed: removed}, nil
}
