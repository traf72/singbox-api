package handlers

import (
	"io"
	"net/http"

	"github.com/traf72/singbox-api/internal/api"
	"github.com/traf72/singbox-api/internal/api/middleware"
	"github.com/traf72/singbox-api/internal/api/query"
	"github.com/traf72/singbox-api/internal/app/logs"
	"github.com/traf72/singbox-api/internal/apperr"
	dlogs "github.com/traf72/singbox-api/internal/config/logs"
)

func download(w http.ResponseWriter, r *http.Request) {
	file, appErr := dlogs.Open()
	if appErr != nil {
		if appErr == dlogs.ErrFileNotFound {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		api.SendError(w, appErr)
		return
	}

	defer file.Close()

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment; filename=log.txt")

	if _, err := io.Copy(w, file); err != nil {
		api.SendInternalServerError(w, apperr.NewFatalErr("Log_WriteError", err.Error()))
	}
}

func enable(w http.ResponseWriter, r *http.Request) {
	setEnabled(w, r, true)
}

func disable(w http.ResponseWriter, r *http.Request) {
	setEnabled(w, r, false)
}

func setEnabled(w http.ResponseWriter, r *http.Request, enable bool) {
	q := r.URL.Query()
	restart, err := query.GetBool(q, "restart", true)
	if err != nil {
		api.SendBadRequest(w, err.Error())
		return
	}

	truncate, err := query.GetBool(q, "truncate", false)
	if err != nil {
		api.SendBadRequest(w, err.Error())
		return
	}

	if err := logs.SetEnabled(enable, restart, truncate); err != nil {
		api.SendError(w, err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func LogsEnableHandler() http.Handler {
	return middleware.NewHandlerFunc(enable).Build()
}

func LogsDisableHandler() http.Handler {
	return middleware.NewHandlerFunc(disable).Build()
}

func LogsDownloadHandler() http.Handler {
	return middleware.NewHandlerFunc(download).Build()
}
