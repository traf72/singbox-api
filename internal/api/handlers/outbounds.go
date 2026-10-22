package handlers

import (
	"net/http"

	"github.com/traf72/singbox-api/internal/api"
	"github.com/traf72/singbox-api/internal/api/middleware"
	"github.com/traf72/singbox-api/internal/api/query"
	"github.com/traf72/singbox-api/internal/app"
	"github.com/traf72/singbox-api/internal/utils"
)

func addOutbound(w http.ResponseWriter, r *http.Request) {
	request := new(app.OutboundRequest)
	if err := utils.FromJSON(r.Body, request); err != nil {
		api.SendBadRequest(w, err.Error())
		return
	}

	noRestart, err := query.GetBool(r.URL.Query(), "norestart", false)
	if err != nil {
		api.SendBadRequest(w, err.Error())
		return
	}

	if err := app.AddOutbound(request, !noRestart); err != nil {
		api.SendError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func removeOutbound(w http.ResponseWriter, r *http.Request) {
	noRestart, err := query.GetBool(r.URL.Query(), "norestart", false)
	if err != nil {
		api.SendBadRequest(w, err.Error())
		return
	}

	if err := app.RemoveOutbound(r.PathValue("tag"), !noRestart); err != nil {
		api.SendError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func AddOutboundHandler() http.Handler {
	return middleware.NewHandlerFunc(addOutbound).WithJsonRequest().Build()
}

func RemoveOutboundHandler() http.Handler {
	return middleware.NewHandlerFunc(removeOutbound).Build()
}
