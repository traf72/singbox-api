package handlers

import (
	"net/http"

	"github.com/traf72/singbox-api/internal/api"
	"github.com/traf72/singbox-api/internal/app"
	"github.com/traf72/singbox-api/internal/utils"
)

func addDNSRule(w http.ResponseWriter, r *http.Request) {
	dnsReq := new(app.DNSRuleDTO)

	if err := utils.ParseJson(r.Body, dnsReq); err != nil {
		api.SendBadRequest(w, err.Error())
		return
	}

	if err := app.AddDNSRule(dnsReq); err != nil {
		api.SendError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func AddDNSRuleHandler() http.Handler {
	return NewHandlerFunc(addDNSRule).WithJsonRequest().handler
}
