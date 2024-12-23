package handlers

import (
	"net/http"

	"github.com/traf72/singbox-api/internal/api/middleware"
)

type HttpHandler struct {
	handler http.Handler
}

func NewHandler(h http.Handler) *HttpHandler {
	return &HttpHandler{handler: h}
}

func NewHandlerFunc(hf http.HandlerFunc) *HttpHandler {
	return &HttpHandler{handler: http.HandlerFunc(hf)}
}

func (h *HttpHandler) WithJsonRequest() *HttpHandler {
	h.handler = middleware.JsonRequest(h.handler)
	return h
}

func (h *HttpHandler) WithRequestLogging() *HttpHandler {
	h.handler = middleware.LogRequest(h.handler)
	return h
}

func (h *HttpHandler) Build() http.Handler {
	return h.handler
}
