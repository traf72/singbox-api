package handlers

import (
	"net/http"

	"github.com/traf72/singbox-api/internal/api"
	"github.com/traf72/singbox-api/internal/api/middleware"
	"github.com/traf72/singbox-api/internal/app"
)

func getRuleAssignments(w http.ResponseWriter, r *http.Request) {
	result, err := app.GetRuleAssignments(r.URL.Query().Get("domain"), r.URL.Query().Get("ip"))
	if err != nil {
		api.SendError(w, err)
		return
	}
	api.SendJson(w, result)
}

func GetRuleAssignmentsHandler() http.Handler {
	return middleware.NewHandlerFunc(getRuleAssignments).Build()
}
