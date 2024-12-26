package handlers

import (
	"net/http"

	"github.com/traf72/singbox-api/internal/api"
	"github.com/traf72/singbox-api/internal/utils"
)

func addDnsRule(w http.ResponseWriter, r *http.Request) {
	var dnsRule string

	if err := utils.ParseJson(r.Body, &dnsRule); err != nil {
		api.SendBadRequest(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func AddDnsRuleHandler() http.Handler {
	return NewHandlerFunc(addDnsRule).WithJsonRequest().handler
}
