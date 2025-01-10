package handlers

import (
	"net/http"

	"github.com/traf72/singbox-api/internal/api"
	"github.com/traf72/singbox-api/internal/api/middleware"
	"github.com/traf72/singbox-api/internal/config/singbox"
)

func start(w http.ResponseWriter, r *http.Request) {
	if err := singbox.Start(); err != nil {
		api.SendError(w, err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func stop(w http.ResponseWriter, r *http.Request) {
	if err := singbox.Stop(); err != nil {
		api.SendError(w, err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func restart(w http.ResponseWriter, r *http.Request) {
	if err := singbox.Restart(); err != nil {
		api.SendError(w, err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func SingboxStartHandler() http.Handler {
	return middleware.NewHandlerFunc(start).Build()
}

func SingboxStopHandler() http.Handler {
	return middleware.NewHandlerFunc(stop).Build()
}

func SingboxRestartHandler() http.Handler {
	return middleware.NewHandlerFunc(restart).Build()
}
