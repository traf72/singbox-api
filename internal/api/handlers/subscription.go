package handlers

import (
	"net/http"

	"github.com/traf72/singbox-api/internal/api"
	"github.com/traf72/singbox-api/internal/api/middleware"
	"github.com/traf72/singbox-api/internal/api/query"
	"github.com/traf72/singbox-api/internal/app"
	"github.com/traf72/singbox-api/internal/utils"
)

const maxSubscriptionRequestSize = 2 << 20

func getSubscriptionSettings(w http.ResponseWriter, _ *http.Request) {
	settings, err := app.GetSubscriptionSettings()
	if err != nil {
		api.SendError(w, err)
		return
	}

	api.SendJson(w, settings)
}

func setSubscriptionSettings(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxSubscriptionRequestSize)
	settings := new(app.SubscriptionSettings)
	if err := utils.FromJSON(r.Body, settings); err != nil {
		api.SendBadRequest(w, err.Error())
		return
	}

	if err := app.SetSubscriptionSettings(settings); err != nil {
		api.SendError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func updateSubscription(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxSubscriptionRequestSize)
	request := new(app.SubscriptionUpdate)
	if err := utils.FromJSON(r.Body, request); err != nil {
		api.SendBadRequest(w, err.Error())
		return
	}

	noRestart, err := query.GetBool(r.URL.Query(), "norestart", false)
	if err != nil {
		api.SendBadRequest(w, err.Error())
		return
	}

	result, appErr := app.UpdateSubscription(request, !noRestart)
	if appErr != nil {
		api.SendError(w, appErr)
		return
	}

	api.SendJson(w, result)
}

func GetSubscriptionSettingsHandler() http.Handler {
	return middleware.NewHandlerFunc(getSubscriptionSettings).Build()
}

func SetSubscriptionSettingsHandler() http.Handler {
	return middleware.NewHandlerFunc(setSubscriptionSettings).WithJsonRequest().Build()
}

func UpdateSubscriptionHandler() http.Handler {
	return middleware.NewHandlerFunc(updateSubscription).WithJsonRequest().Build()
}
