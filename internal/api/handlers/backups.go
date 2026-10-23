package handlers

import (
	"net/http"
	"time"

	"github.com/traf72/singbox-api/internal/api"
	"github.com/traf72/singbox-api/internal/api/middleware"
	"github.com/traf72/singbox-api/internal/api/query"
	"github.com/traf72/singbox-api/internal/app"
)

func listBackups(w http.ResponseWriter, _ *http.Request) {
	backups, err := app.ListBackups()
	if err != nil {
		api.SendError(w, err)
		return
	}

	api.SendJson(w, backups)
}

func createBackup(w http.ResponseWriter, _ *http.Request) {
	backup, err := app.CreateBackup()
	if err != nil {
		api.SendError(w, err)
		return
	}

	api.SendJsonStatus(w, http.StatusCreated, backup)
}

func restoreBackup(w http.ResponseWriter, r *http.Request) {
	noRestart, err := query.GetBool(r.URL.Query(), "norestart", false)
	if err != nil {
		api.SendBadRequest(w, err.Error())
		return
	}
	if appErr := app.RestoreBackup(r.PathValue("name"), !noRestart); appErr != nil {
		api.SendError(w, appErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func removeOldBackups(w http.ResponseWriter, r *http.Request) {
	before := r.URL.Query().Get("before")
	cutoff, err := time.Parse(time.RFC3339, before)
	if err != nil {
		api.SendBadRequest(w, "before must be an RFC3339 date and time")
		return
	}

	result, appErr := app.RemoveBackupsOlderThan(cutoff)
	if appErr != nil {
		api.SendError(w, appErr)
		return
	}

	api.SendJson(w, result)
}

func ListBackupsHandler() http.Handler {
	return middleware.NewHandlerFunc(listBackups).Build()
}

func CreateBackupHandler() http.Handler {
	return middleware.NewHandlerFunc(createBackup).Build()
}

func RestoreBackupHandler() http.Handler {
	return middleware.NewHandlerFunc(restoreBackup).Build()
}

func RemoveOldBackupsHandler() http.Handler {
	return middleware.NewHandlerFunc(removeOldBackups).Build()
}
