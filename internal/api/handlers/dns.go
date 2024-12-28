package handlers

import (
	"net/http"

	"github.com/traf72/singbox-api/internal/api"
	"github.com/traf72/singbox-api/internal/api/middleware"
	"github.com/traf72/singbox-api/internal/app/dns"
	"github.com/traf72/singbox-api/internal/utils"
)

func addDNSRule(w http.ResponseWriter, r *http.Request) {
	dnsReq := new(dns.Rule)

	if err := utils.ParseJson(r.Body, dnsReq); err != nil {
		api.SendBadRequest(w, err.Error())
		return
	}

	if err := dns.AddRule(dnsReq); err != nil {
		api.SendError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func removeDNSRule(w http.ResponseWriter, r *http.Request) {
	dnsReq := new(dns.Rule)

	if err := utils.ParseJson(r.Body, dnsReq); err != nil {
		api.SendBadRequest(w, err.Error())
		return
	}

	if err := dns.RemoveRule(dnsReq); err != nil {
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
