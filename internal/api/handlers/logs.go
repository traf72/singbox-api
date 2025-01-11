package handlers

import (
	"io"
	"net/http"

	"github.com/traf72/singbox-api/internal/api"
	"github.com/traf72/singbox-api/internal/api/header"
	"github.com/traf72/singbox-api/internal/api/middleware"
	"github.com/traf72/singbox-api/internal/api/query"
	"github.com/traf72/singbox-api/internal/app"
	"github.com/traf72/singbox-api/internal/apperr"
	"github.com/traf72/singbox-api/internal/singbox"
)

func downloadLog(w http.ResponseWriter, r *http.Request) {
	file, appErr := singbox.GetLog()
	if appErr != nil {
		if appErr == singbox.ErrLogNotFound {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		api.SendError(w, appErr)
		return
	}

	defer file.Close()

	header.SetContentType(w, header.ContentTypeTextPlain)
	header.SetAttachment(w, "log.txt")

	if _, err := io.Copy(w, file); err != nil {
		api.SendInternalServerError(w, apperr.NewFatalErr("Log_WriteError", err.Error()))
	}
}

func enableLog(w http.ResponseWriter, r *http.Request) {
	setLogEnabled(w, r, true)
}

func disableLog(w http.ResponseWriter, r *http.Request) {
	setLogEnabled(w, r, false)
}

func setLogEnabled(w http.ResponseWriter, r *http.Request, enable bool) {
	q := r.URL.Query()
	noRestart, err := query.GetBool(q, "norestart", false)
	if err != nil {
		api.SendBadRequest(w, err.Error())
		return
	}

	truncate, err := query.GetBool(q, "truncate", false)
	if err != nil {
		api.SendBadRequest(w, err.Error())
		return
	}

	var f func(bool, bool) apperr.Err
	if enable {
		f = app.EnableLog
	} else {
		f = app.DisableLog
	}

	if err := f(!noRestart, truncate); err != nil {
		api.SendError(w, err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func truncateLog(w http.ResponseWriter, _ *http.Request) {
	if appErr := app.TruncateLog(); appErr != nil {
		api.SendError(w, appErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func LogDownloadHandler() http.Handler {
	return middleware.NewHandlerFunc(downloadLog).Build()
}

func LogsEnableHandler() http.Handler {
	return middleware.NewHandlerFunc(enableLog).Build()
}

func LogsDisableHandler() http.Handler {
	return middleware.NewHandlerFunc(disableLog).Build()
}

func LogTruncateHandler() http.Handler {
	return middleware.NewHandlerFunc(truncateLog).Build()
}
