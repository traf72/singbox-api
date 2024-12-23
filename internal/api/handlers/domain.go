package handlers

import (
	"net/http"

	"github.com/traf72/singbox-api/internal/api"
	"github.com/traf72/singbox-api/internal/utils"
)

func addDomains(w http.ResponseWriter, r *http.Request) {
	var domains []string

	if err := utils.ParseJson(r.Body, &domains); err != nil {
		api.SendBadRequest(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func AddDomainsHandler() http.Handler {
	return NewHandlerFunc(addDomains).WithJsonRequest().handler
}
