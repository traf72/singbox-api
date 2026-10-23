package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackupHandlers(t *testing.T) {
	directory := t.TempDir()
	configPath := filepath.Join(directory, "config.json")
	t.Setenv("CONFIG_PATH", configPath)
	require.NoError(t, os.WriteFile(configPath, []byte(`{"version":"current"}`), 0o600))

	backupDirectory := filepath.Join(directory, "backups")
	require.NoError(t, os.Mkdir(backupDirectory, 0o700))
	oldPath := filepath.Join(backupDirectory, "20261020_120000")
	newPath := filepath.Join(backupDirectory, "20261022_172500")
	require.NoError(t, os.WriteFile(oldPath, []byte(`{"version":"old"}`), 0o600))
	require.NoError(t, os.WriteFile(newPath, []byte(`{"version":"new"}`), 0o600))
	oldTime := time.Date(2026, time.October, 20, 12, 0, 0, 0, time.UTC)
	newTime := time.Date(2026, time.October, 22, 17, 25, 0, 0, time.UTC)
	require.NoError(t, os.Chtimes(oldPath, oldTime, oldTime))
	require.NoError(t, os.Chtimes(newPath, newTime, newTime))

	router := http.NewServeMux()
	router.Handle("GET /backups", ListBackupsHandler())
	router.Handle("POST /backups", CreateBackupHandler())
	router.Handle("POST /backups/{name}/restore", RestoreBackupHandler())
	router.Handle("DELETE /backups", RemoveOldBackupsHandler())

	listResponse := serveJSONRequest(router, http.MethodGet, "/backups", "")
	require.Equal(t, http.StatusOK, listResponse.Code, listResponse.Body.String())
	assert.Contains(t, listResponse.Body.String(), `"name": "20261022_172500"`)
	assert.Contains(t, listResponse.Body.String(), `"name": "20261020_120000"`)

	createResponse := serveJSONRequest(router, http.MethodPost, "/backups", "")
	require.Equal(t, http.StatusCreated, createResponse.Code, createResponse.Body.String())
	var created struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	}
	require.NoError(t, json.Unmarshal(createResponse.Body.Bytes(), &created))
	assert.NotEmpty(t, created.Name)
	assert.Positive(t, created.Size)
	assert.FileExists(t, filepath.Join(backupDirectory, created.Name))

	invalidCleanupResponse := serveJSONRequest(router, http.MethodDelete, "/backups?before=not-a-date", "")
	assert.Equal(t, http.StatusBadRequest, invalidCleanupResponse.Code)

	cleanupResponse := serveJSONRequest(router, http.MethodDelete, "/backups?before=2026-10-21T00:00:00Z", "")
	require.Equal(t, http.StatusOK, cleanupResponse.Code, cleanupResponse.Body.String())
	assert.JSONEq(t, `{"removed":1}`, cleanupResponse.Body.String())
	assert.NoFileExists(t, oldPath)

	restoreResponse := serveJSONRequest(router, http.MethodPost, "/backups/20261022_172500/restore?norestart=true", "")
	require.Equal(t, http.StatusNoContent, restoreResponse.Code, restoreResponse.Body.String())
	contents, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.JSONEq(t, `{"version":"new"}`, string(contents))

	missingResponse := serveJSONRequest(router, http.MethodPost, "/backups/20261021_000000/restore?norestart=true", "")
	assert.Equal(t, http.StatusNotFound, missingResponse.Code)
}
