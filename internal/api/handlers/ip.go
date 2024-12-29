package handlers

import (
	"net/http"

	"github.com/traf72/singbox-api/internal/api"
	"github.com/traf72/singbox-api/internal/api/middleware"
	"github.com/traf72/singbox-api/internal/app/ip"
	"github.com/traf72/singbox-api/internal/utils"
)

func addIPRule(w http.ResponseWriter, r *http.Request) {
	dnsReq := new(ip.Rule)

	if err := utils.ParseJson(r.Body, dnsReq); err != nil {
		api.SendBadRequest(w, err.Error())
		return
	}

	if err := ip.AddRule(dnsReq); err != nil {
		api.SendError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func removeIPRule(w http.ResponseWriter, r *http.Request) {
	dnsReq := new(ip.Rule)

	if err := utils.ParseJson(r.Body, dnsReq); err != nil {
		api.SendBadRequest(w, err.Error())
		return
	}

	if err := ip.RemoveRule(dnsReq); err != nil {
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
