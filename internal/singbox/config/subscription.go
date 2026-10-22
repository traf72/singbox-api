package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/traf72/singbox-api/internal/apperr"
)

const subscriptionVersion = 1

type SubscriptionSettings struct {
	Version          int      `json:"version"`
	FixedOutbounds   []string `json:"fixed_outbounds"`
	FallbackOutbound string   `json:"fallback_outbound"`
}

type SubscriptionUpdate struct {
	Version   int               `json:"version"`
	Outbounds []json.RawMessage `json:"outbounds"`
}

type SkippedSubscriptionOutbound struct {
	Tag    string `json:"tag"`
	Reason string `json:"reason"`
}

type SubscriptionUpdateResult struct {
	Added     int                           `json:"added"`
	Removed   int                           `json:"removed"`
	Preserved int                           `json:"preserved"`
	Skipped   []SkippedSubscriptionOutbound `json:"skipped"`
	Default   string                        `json:"default"`
}

type BackupInfo struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	Size      int64     `json:"size"`
}

var subscriptionMutex sync.Mutex

var subscriptionUUIDRegex = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
var subscriptionHostnameRegex = regexp.MustCompile(`(?i)^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)*[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)
var subscriptionBackupNameRegex = regexp.MustCompile(`^\d{8}_\d{4}(?:\d{2})?(?:_\d+)?$`)

func errSubscriptionValidation(message string) apperr.Err {
	return apperr.NewValidationErr("Subscription_Invalid", message)
}

func errSubscriptionSettingsNotFound() apperr.Err {
	return apperr.NewNotFoundErr("Subscription_SettingsNotFound", "subscription settings were not found")
}

func errSubscriptionRead(err error) apperr.Err {
	return apperr.NewFatalErr("Subscription_ReadError", err.Error())
}

func errSubscriptionWrite(err error) apperr.Err {
	return apperr.NewFatalErr("Subscription_WriteError", err.Error())
}

func GetSubscriptionSettings() (*SubscriptionSettings, apperr.Err) {
	subscriptionMutex.Lock()
	defer subscriptionMutex.Unlock()

	return loadSubscriptionSettings()
}

func SetSubscriptionSettings(settings *SubscriptionSettings) apperr.Err {
	subscriptionMutex.Lock()
	defer subscriptionMutex.Unlock()

	if settings.Version == 0 {
		settings.Version = subscriptionVersion
	}

	c, err := Load()
	if err != nil {
		return err
	}
	if err := validateSubscriptionSettings(settings, c.Conf); err != nil {
		return err
	}

	return saveSubscriptionSettings(settings)
}

func ListBackups() ([]BackupInfo, apperr.Err) {
	subscriptionMutex.Lock()
	defer subscriptionMutex.Unlock()

	directory, err := subscriptionBackupDirectory()
	if err != nil {
		return nil, err
	}
	entries, readErr := os.ReadDir(directory)
	if errors.Is(readErr, os.ErrNotExist) {
		return []BackupInfo{}, nil
	}
	if readErr != nil {
		return nil, errSubscriptionRead(readErr)
	}

	backups := make([]BackupInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !subscriptionBackupNameRegex.MatchString(entry.Name()) {
			continue
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			return nil, errSubscriptionRead(infoErr)
		}
		backups = append(backups, BackupInfo{
			Name:      entry.Name(),
			CreatedAt: info.ModTime(),
			Size:      info.Size(),
		})
	}
	sort.Slice(backups, func(i, j int) bool {
		if backups[i].CreatedAt.Equal(backups[j].CreatedAt) {
			return backups[i].Name > backups[j].Name
		}
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	return backups, nil
}

func RestoreBackup(name string) apperr.Err {
	subscriptionMutex.Lock()
	defer subscriptionMutex.Unlock()

	backupPath, err := resolveBackupPath(name)
	if err != nil {
		return err
	}
	contents, readErr := os.ReadFile(backupPath)
	if errors.Is(readErr, os.ErrNotExist) {
		return apperr.NewNotFoundErr("Backup_NotFound", fmt.Sprintf("backup '%s' was not found", name))
	}
	if readErr != nil {
		return errSubscriptionRead(readErr)
	}

	var conf *Conf
	if unmarshalErr := json.Unmarshal(contents, &conf); unmarshalErr != nil || conf == nil {
		return errSubscriptionValidation(fmt.Sprintf("backup '%s' does not contain a valid configuration", name))
	}
	saveMutex.Lock()
	defer saveMutex.Unlock()
	if err := backupConfig(); err != nil {
		return err
	}

	configPath, err := getConfPath()
	if err != nil {
		return err
	}
	if writeErr := writeAtomicFile(configPath, contents); writeErr != nil {
		return errSubscriptionWrite(writeErr)
	}

	return nil
}

func RemoveBackupsOlderThan(cutoff time.Time) (int, apperr.Err) {
	subscriptionMutex.Lock()
	defer subscriptionMutex.Unlock()

	directory, err := subscriptionBackupDirectory()
	if err != nil {
		return 0, err
	}
	entries, readErr := os.ReadDir(directory)
	if errors.Is(readErr, os.ErrNotExist) {
		return 0, nil
	}
	if readErr != nil {
		return 0, errSubscriptionRead(readErr)
	}

	removed := 0
	for _, entry := range entries {
		if entry.IsDir() || !subscriptionBackupNameRegex.MatchString(entry.Name()) {
			continue
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			return removed, errSubscriptionRead(infoErr)
		}
		if !info.ModTime().Before(cutoff) {
			continue
		}
		if removeErr := os.Remove(filepath.Join(directory, entry.Name())); removeErr != nil {
			return removed, errSubscriptionWrite(removeErr)
		}
		removed++
	}

	return removed, nil
}

func UpdateSubscription(request *SubscriptionUpdate) (*SubscriptionUpdateResult, apperr.Err) {
	subscriptionMutex.Lock()
	defer subscriptionMutex.Unlock()

	if request.Version != subscriptionVersion {
		return nil, errSubscriptionValidation(fmt.Sprintf("unsupported subscription version %d", request.Version))
	}

	c, err := Load()
	if err != nil {
		return nil, err
	}
	settings, err := loadSubscriptionSettings()
	if err != nil {
		return nil, err
	}
	if err := validateSubscriptionSettings(settings, c.Conf); err != nil {
		return nil, err
	}

	result, err := applySubscriptionUpdate(request, settings, c.Conf)
	if err != nil {
		return nil, err
	}
	if err := backupConfig(); err != nil {
		return nil, err
	}
	if err := Save(c); err != nil {
		return nil, err
	}

	return result, nil
}

func applySubscriptionUpdate(request *SubscriptionUpdate, settings *SubscriptionSettings, conf *Conf) (*SubscriptionUpdateResult, apperr.Err) {
	selector := findProxySelector(conf)
	if selector == nil {
		return nil, errProxySelectorNotFound()
	}

	fixed := make(map[string]struct{}, len(settings.FixedOutbounds))
	for _, tag := range settings.FixedOutbounds {
		fixed[tag] = struct{}{}
	}

	structural := make(map[string]struct{})
	for _, item := range conf.Outbounds {
		if isStructuralOutbound(item) {
			structural[item.Tag] = struct{}{}
		}
	}

	candidates := make([]*outbound, 0, len(request.Outbounds))
	skipped := make([]SkippedSubscriptionOutbound, 0)
	seen := make(map[string]struct{})
	for index, raw := range request.Outbounds {
		var item outbound
		if err := json.Unmarshal(raw, &item); err != nil {
			return nil, errSubscriptionValidation(fmt.Sprintf("outbound at index %d is invalid: %s", index, err))
		}

		item.Tag = strings.TrimSpace(item.Tag)
		item.Server = strings.TrimSpace(item.Server)
		if item.Server == "" {
			skipped = append(skipped, SkippedSubscriptionOutbound{Tag: item.Tag, Reason: "missing server"})
			continue
		}
		if item.Type != "vless" {
			skipped = append(skipped, SkippedSubscriptionOutbound{Tag: item.Tag, Reason: fmt.Sprintf("unsupported outbound type '%s'", item.Type)})
			continue
		}
		if err := validateSubscriptionVLESS(&item, index); err != nil {
			return nil, err
		}
		if _, ok := seen[item.Tag]; ok {
			return nil, errSubscriptionValidation(fmt.Sprintf("duplicate outbound tag '%s'", item.Tag))
		}
		seen[item.Tag] = struct{}{}

		if _, ok := fixed[item.Tag]; ok {
			skipped = append(skipped, SkippedSubscriptionOutbound{Tag: item.Tag, Reason: "tag is fixed"})
			continue
		}
		if _, ok := structural[item.Tag]; ok {
			skipped = append(skipped, SkippedSubscriptionOutbound{Tag: item.Tag, Reason: "tag is reserved"})
			continue
		}

		candidates = append(candidates, &item)
	}

	if len(candidates) == 0 {
		return nil, errSubscriptionValidation("subscription contains no usable outbounds")
	}

	retained := make([]*outbound, 0, len(conf.Outbounds)+len(candidates))
	removed := 0
	for _, item := range conf.Outbounds {
		if item == nil {
			removed++
			continue
		}
		_, isFixed := fixed[item.Tag]
		if isFixed || isStructuralOutbound(item) {
			retained = append(retained, item)
			continue
		}

		removed++
	}
	retained = append(retained, candidates...)
	conf.Outbounds = retained

	selector.Outbounds = append([]string(nil), settings.FixedOutbounds...)
	for _, item := range candidates {
		selector.Outbounds = append(selector.Outbounds, item.Tag)
	}
	if !slices.Contains(selector.Outbounds, selector.Default) {
		selector.Default = settings.FallbackOutbound
	}

	return &SubscriptionUpdateResult{
		Added:     len(candidates),
		Removed:   removed,
		Preserved: len(retained) - len(candidates),
		Skipped:   skipped,
		Default:   selector.Default,
	}, nil
}

func validateSubscriptionVLESS(item *outbound, index int) apperr.Err {
	prefix := fmt.Sprintf("outbound at index %d", index)
	if item.Tag == "" {
		return errSubscriptionValidation(prefix + " has an empty tag")
	}
	if net.ParseIP(item.Server) == nil && !subscriptionHostnameRegex.MatchString(item.Server) {
		return errSubscriptionValidation(fmt.Sprintf("%s has invalid server '%s'", prefix, item.Server))
	}
	if item.ServerPort < 1 || item.ServerPort > 65535 {
		return errSubscriptionValidation(prefix + " has an invalid server port")
	}
	if !subscriptionUUIDRegex.MatchString(item.UUID) {
		return errSubscriptionValidation(prefix + " has a missing or invalid UUID")
	}
	if item.TLS == nil || !item.TLS.Enabled {
		return errSubscriptionValidation(prefix + " must have TLS enabled")
	}
	if item.TLS.ServerName == "" {
		return errSubscriptionValidation(prefix + " has an empty TLS server name")
	}
	if item.TLS.Reality == nil || !item.TLS.Reality.Enabled || item.TLS.Reality.PublicKey == "" || item.TLS.Reality.ShortID == "" {
		return errSubscriptionValidation(prefix + " has incomplete Reality settings")
	}

	return nil
}

func validateSubscriptionSettings(settings *SubscriptionSettings, conf *Conf) apperr.Err {
	if settings.Version != subscriptionVersion {
		return errSubscriptionValidation(fmt.Sprintf("unsupported settings version %d", settings.Version))
	}
	if len(settings.FixedOutbounds) == 0 {
		return errSubscriptionValidation("fixed_outbounds must not be empty")
	}

	selector := findProxySelector(conf)
	if selector == nil {
		return errProxySelectorNotFound()
	}

	seen := make(map[string]struct{}, len(settings.FixedOutbounds))
	for index, rawTag := range settings.FixedOutbounds {
		tag := strings.TrimSpace(rawTag)
		if tag == "" {
			return errSubscriptionValidation(fmt.Sprintf("fixed outbound at index %d is empty", index))
		}
		if _, ok := seen[tag]; ok {
			return errSubscriptionValidation(fmt.Sprintf("fixed outbound tag '%s' is duplicated", tag))
		}
		if !slices.Contains(selector.Outbounds, tag) || findOutbound(conf, tag) == nil {
			return errSubscriptionValidation(fmt.Sprintf("fixed outbound tag '%s' does not exist in selector '%s'", tag, proxySelectorTag))
		}

		settings.FixedOutbounds[index] = tag
		seen[tag] = struct{}{}
	}

	settings.FallbackOutbound = strings.TrimSpace(settings.FallbackOutbound)
	if _, ok := seen[settings.FallbackOutbound]; !ok {
		return errSubscriptionValidation("fallback_outbound must be included in fixed_outbounds")
	}

	return nil
}

func isStructuralOutbound(item *outbound) bool {
	return item != nil && (item.Type == "selector" || item.Type == "direct" || item.Type == "block")
}

func findOutbound(conf *Conf, tag string) *outbound {
	index := slices.IndexFunc(conf.Outbounds, func(item *outbound) bool {
		return item != nil && item.Tag == tag
	})
	if index == -1 {
		return nil
	}

	return conf.Outbounds[index]
}

func loadSubscriptionSettings() (*SubscriptionSettings, apperr.Err) {
	path, err := subscriptionSettingsPath()
	if err != nil {
		return nil, err
	}

	contents, readErr := os.ReadFile(path)
	if errors.Is(readErr, os.ErrNotExist) {
		return nil, errSubscriptionSettingsNotFound()
	}
	if readErr != nil {
		return nil, errSubscriptionRead(readErr)
	}

	settings := new(SubscriptionSettings)
	if err := json.Unmarshal(contents, settings); err != nil {
		return nil, errSubscriptionRead(err)
	}
	return settings, nil
}

func saveSubscriptionSettings(settings *SubscriptionSettings) apperr.Err {
	path, err := subscriptionSettingsPath()
	if err != nil {
		return err
	}
	contents, marshalErr := json.MarshalIndent(settings, "", "    ")
	if marshalErr != nil {
		return errSubscriptionWrite(marshalErr)
	}
	contents = append(contents, '\n')

	if writeErr := writeAtomicFile(path, contents); writeErr != nil {
		return errSubscriptionWrite(writeErr)
	}

	return nil
}

func backupConfig() apperr.Err {
	path, err := getConfPath()
	if err != nil {
		return err
	}
	contents, readErr := os.ReadFile(path)
	if readErr != nil {
		return errSubscriptionRead(readErr)
	}
	return storeSubscriptionBackup(contents, time.Now())
}

func storeSubscriptionBackup(contents []byte, timestamp time.Time) apperr.Err {
	backupDirectory, err := subscriptionBackupDirectory()
	if err != nil {
		return err
	}
	if mkdirErr := os.MkdirAll(backupDirectory, 0o700); mkdirErr != nil {
		return errSubscriptionWrite(mkdirErr)
	}
	backupName := timestamp.Format("20060102_150405")
	backupPath := filepath.Join(backupDirectory, backupName)
	for suffix := 2; ; suffix++ {
		_, statErr := os.Stat(backupPath)
		if errors.Is(statErr, os.ErrNotExist) {
			break
		}
		if statErr != nil {
			return errSubscriptionRead(statErr)
		}
		backupName = fmt.Sprintf("%s_%d", timestamp.Format("20060102_150405"), suffix)
		backupPath = filepath.Join(backupDirectory, backupName)
	}
	if writeErr := writeAtomicFile(backupPath, contents); writeErr != nil {
		return errSubscriptionWrite(writeErr)
	}

	return nil
}

func subscriptionBackupDirectory() (string, apperr.Err) {
	path, err := getConfPath()
	if err != nil {
		return "", err
	}

	return filepath.Join(filepath.Dir(path), "backups"), nil
}

func resolveBackupPath(name string) (string, apperr.Err) {
	name = strings.TrimSpace(name)
	if !subscriptionBackupNameRegex.MatchString(name) || filepath.Base(name) != name {
		return "", apperr.NewValidationErr("Backup_InvalidName", "backup name is invalid")
	}
	directory, err := subscriptionBackupDirectory()
	if err != nil {
		return "", err
	}

	return filepath.Join(directory, name), nil
}

func subscriptionSettingsPath() (string, apperr.Err) {
	path, err := getConfPath()
	if err != nil {
		return "", err
	}

	name := filepath.Base(path)
	extension := filepath.Ext(name)
	name = strings.TrimSuffix(name, extension)
	return filepath.Join(filepath.Dir(path), name+".subscription.state"), nil
}

func writeAtomicFile(path string, contents []byte) error {
	tempPath := path + ".tmp"
	file, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	if _, err = file.Write(contents); err == nil {
		err = file.Sync()
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(tempPath)
		return err
	}
	if err = os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath)
		return err
	}

	return nil
}

func protectManualOutbound(tag string, conf *Conf) apperr.Err {
	settings, err := loadSubscriptionSettings()
	if err != nil {
		if err.Kind() != apperr.NotFound {
			return err
		}

		selector := findProxySelector(conf)
		if selector == nil {
			return errProxySelectorNotFound()
		}
		settings = &SubscriptionSettings{Version: subscriptionVersion}
		for _, member := range selector.Outbounds {
			if findOutbound(conf, member) != nil && !slices.Contains(settings.FixedOutbounds, member) {
				settings.FixedOutbounds = append(settings.FixedOutbounds, member)
			}
		}
		if slices.Contains(settings.FixedOutbounds, selector.Default) {
			settings.FallbackOutbound = selector.Default
		} else if len(settings.FixedOutbounds) > 0 {
			settings.FallbackOutbound = settings.FixedOutbounds[0]
		}
	}

	if !slices.Contains(settings.FixedOutbounds, tag) {
		settings.FixedOutbounds = append(settings.FixedOutbounds, tag)
	}
	if settings.FallbackOutbound == "" {
		settings.FallbackOutbound = tag
	}

	return saveSubscriptionSettings(settings)
}

func removeManualOutboundProtection(tag string) apperr.Err {
	settings, err := loadSubscriptionSettings()
	if err != nil {
		if err.Kind() == apperr.NotFound {
			return nil
		}
		return err
	}
	if !slices.Contains(settings.FixedOutbounds, tag) {
		return nil
	}
	if settings.FallbackOutbound == tag {
		return apperr.NewConflictErr("Subscription_FallbackDelete", "cannot delete the subscription fallback outbound; select another fallback first")
	}

	settings.FixedOutbounds = slices.DeleteFunc(settings.FixedOutbounds, func(item string) bool {
		return item == tag
	})
	return saveSubscriptionSettings(settings)
}

func validateManualOutboundRemoval(tag string) apperr.Err {
	settings, err := loadSubscriptionSettings()
	if err != nil {
		if err.Kind() == apperr.NotFound {
			return nil
		}
		return err
	}
	if settings.FallbackOutbound == tag {
		return apperr.NewConflictErr("Subscription_FallbackDelete", "cannot delete the subscription fallback outbound; select another fallback first")
	}

	return nil
}
