package handlers

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSingboxStatusHandlerWhenInteractionDisabled(t *testing.T) {
	t.Setenv("DISABLE_SINGBOX_INTERACTION", "true")
	router := http.NewServeMux()
	router.Handle("GET /singbox/status", SingboxStatusHandler())

	response := serveJSONRequest(router, http.MethodGet, "/singbox/status", "")

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.JSONEq(t, `{
		"main_pid": 0,
		"exec_main_status": 0,
		"active": false,
		"interaction_disabled": true
	}`, response.Body.String())
}
