package handlers

import (
	"net/http"

	"github.com/traf72/singbox-api/internal/api"
	"github.com/traf72/singbox-api/internal/api/middleware"
	"github.com/traf72/singbox-api/internal/api/query"
	"github.com/traf72/singbox-api/internal/app"
	"github.com/traf72/singbox-api/internal/utils"
)

func addIPRule(w http.ResponseWriter, r *http.Request) {
	ipReq := new(app.IPRule)

	if err := utils.FromJSON(r.Body, ipReq); err != nil {
		api.SendBadRequest(w, err.Error())
		return
	}

	noRestart, err := query.GetBool(r.URL.Query(), "norestart", false)
	if err != nil {
		api.SendBadRequest(w, err.Error())
		return
	}

	if err := app.AddIPRule(ipReq, !noRestart); err != nil {
		api.SendError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func removeIPRule(w http.ResponseWriter, r *http.Request) {
	ipReq := new(app.IPRule)

	if err := utils.FromJSON(r.Body, ipReq); err != nil {
		api.SendBadRequest(w, err.Error())
		return
	}

	noRestart, err := query.GetBool(r.URL.Query(), "norestart", false)
	if err != nil {
		api.SendBadRequest(w, err.Error())
		return
	}

	if err := app.RemoveIPRule(ipReq, !noRestart); err != nil {
		api.SendError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func AddIPRuleHandler() http.Handler {
	return middleware.NewHandlerFunc(addIPRule).WithJsonRequest().Build()
}

func RemoveIPRuleHandler() http.Handler {
	return middleware.NewHandlerFunc(removeIPRule).WithJsonRequest().Build()
}
