package middleware

import (
	"net/http"
)

type HttpHandler struct {
	handler http.Handler
}

func NewHandler(h http.Handler) *HttpHandler {
	return &HttpHandler{handler: h}
}

func NewHandlerFunc(h http.HandlerFunc) *HttpHandler {
	return &HttpHandler{handler: http.HandlerFunc(h)}
}

func (h *HttpHandler) WithJsonRequest() *HttpHandler {
	h.handler = JsonRequest(h.handler)
	return h
}

func (h *HttpHandler) WithRequestLogging() *HttpHandler {
	h.handler = LogRequest(h.handler)
	return h
}

func (h *HttpHandler) Build() http.Handler {
	return h.handler
}
