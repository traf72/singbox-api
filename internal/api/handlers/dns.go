package handlers

import (
	"net/http"

	"github.com/traf72/singbox-api/internal/api"
	"github.com/traf72/singbox-api/internal/api/middleware"
	"github.com/traf72/singbox-api/internal/api/query"
	"github.com/traf72/singbox-api/internal/app"
	"github.com/traf72/singbox-api/internal/utils"
)

func addDNSRule(w http.ResponseWriter, r *http.Request) {
	dnsReq := new(app.DNSRule)

	if err := utils.FromJSON(r.Body, dnsReq); err != nil {
		api.SendBadRequest(w, err.Error())
		return
	}

	noRestart, err := query.GetBool(r.URL.Query(), "norestart", false)
	if err != nil {
		api.SendBadRequest(w, err.Error())
		return
	}

	if err := app.AddDNSRule(dnsReq, !noRestart); err != nil {
		api.SendError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func removeDNSRule(w http.ResponseWriter, r *http.Request) {
	dnsReq := new(app.DNSRule)

	if err := utils.FromJSON(r.Body, dnsReq); err != nil {
		api.SendBadRequest(w, err.Error())
		return
	}

	noRestart, err := query.GetBool(r.URL.Query(), "norestart", false)
	if err != nil {
		api.SendBadRequest(w, err.Error())
		return
	}

	if err := app.RemoveDNSRule(dnsReq, !noRestart); err != nil {
		api.SendError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func AddDNSRuleHandler() http.Handler {
	return middleware.NewHandlerFunc(addDNSRule).WithJsonRequest().Build()
}

func RemoveDNSRuleHandler() http.Handler {
	return middleware.NewHandlerFunc(removeDNSRule).WithJsonRequest().Build()
}
