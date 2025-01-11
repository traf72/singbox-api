package handlers

import (
	"net/http"

	"github.com/traf72/singbox-api/internal/api"
	"github.com/traf72/singbox-api/internal/api/middleware"
	"github.com/traf72/singbox-api/internal/singbox"
)

func startSingbox(w http.ResponseWriter, r *http.Request) {
	if err := singbox.Start(); err != nil {
		api.SendError(w, err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func stopSingbox(w http.ResponseWriter, r *http.Request) {
	if err := singbox.Stop(); err != nil {
		api.SendError(w, err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func restartSingbox(w http.ResponseWriter, r *http.Request) {
	if err := singbox.Restart(); err != nil {
		api.SendError(w, err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func SingboxStartHandler() http.Handler {
	return middleware.NewHandlerFunc(startSingbox).Build()
}

func SingboxStopHandler() http.Handler {
	return middleware.NewHandlerFunc(stopSingbox).Build()
}

func SingboxRestartHandler() http.Handler {
	return middleware.NewHandlerFunc(restartSingbox).Build()
}
