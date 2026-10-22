package handlers

import (
	"net/http"

	"github.com/traf72/singbox-api/internal/api"
	"github.com/traf72/singbox-api/internal/api/middleware"
	"github.com/traf72/singbox-api/internal/api/query"
	"github.com/traf72/singbox-api/internal/app"
	"github.com/traf72/singbox-api/internal/utils"
)

func setRouteMode(w http.ResponseWriter, r *http.Request) {
	request := new(app.RouteModeRequest)
	if err := utils.FromJSON(r.Body, request); err != nil {
		api.SendBadRequest(w, err.Error())
		return
	}

	noRestart, err := query.GetBool(r.URL.Query(), "norestart", false)
	if err != nil {
		api.SendBadRequest(w, err.Error())
		return
	}

	if err := app.SetRouteMode(request, !noRestart); err != nil {
		api.SendError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func SetRouteModeHandler() http.Handler {
	return middleware.NewHandlerFunc(setRouteMode).WithJsonRequest().Build()
}
