package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/traf72/singbox-api/internal/api"
	"github.com/traf72/singbox-api/internal/api/middleware"
	"github.com/traf72/singbox-api/internal/app/config"
)

const downloadQueryParam = "download"

func getConfig(w http.ResponseWriter, r *http.Request) {
	c, appErr := config.GetConfig()
	if appErr != nil {
		api.SendError(w, appErr)
		return
	}

	download := false

	downloadVal := r.URL.Query().Get(downloadQueryParam)
	if downloadVal != "" {
		var err error
		download, err = strconv.ParseBool(downloadVal)
		if err != nil {
			api.SendBadRequest(w, fmt.Sprintf("invalid value '%s' for query parameter '%s': expected a boolean (true, false)", downloadVal, downloadQueryParam))
			return
		}
	}

	if download {
		w.Header().Set("Content-Disposition", "attachment; filename=config.json")
	}

	api.SendJson(w, c)
}

func GetConfigHandler() http.Handler {
	return middleware.NewHandlerFunc(getConfig).Build()
}
