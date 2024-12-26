package handlers

import (
	"net/http"

	"github.com/traf72/singbox-api/internal/api"
	"github.com/traf72/singbox-api/internal/utils"
)

func addTemplate(w http.ResponseWriter, r *http.Request) {
	var template string

	if err := utils.ParseJson(r.Body, &template); err != nil {
		api.SendBadRequest(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func AddTemplateHandler() http.Handler {
	return NewHandlerFunc(addTemplate).WithJsonRequest().handler
}
