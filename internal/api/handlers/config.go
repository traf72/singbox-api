package handlers

import (
	"net/http"

	"github.com/traf72/singbox-api/internal/api"
	"github.com/traf72/singbox-api/internal/api/middleware"
	"github.com/traf72/singbox-api/internal/api/query"
	"github.com/traf72/singbox-api/internal/app/config"
)

func getConfig(w http.ResponseWriter, r *http.Request) {
	c, appErr := config.GetConfig()
	if appErr != nil {
		api.SendError(w, appErr)
		return
	}

	download, err := query.GetBool(r.URL.Query(), "download", false)
	if err != nil {
		api.SendBadRequest(w, err.Error())
		return
	}

	if download {
		api.SetAttachment(w, "config.json")
	}

	api.SendJson(w, c)
}

func GetConfigHandler() http.Handler {
	return middleware.NewHandlerFunc(getConfig).Build()
}
