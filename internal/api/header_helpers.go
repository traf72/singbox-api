package api

import (
	"fmt"
	"net/http"
)

const (
	HeaderContentType        = "Content-Type"
	HeaderContentDisposition = "Content-Disposition"
)

const (
	ContentTypeJson      = "application/json"
	ContentTypeTextPlain = "text/plain"
)

func SetContentType(w http.ResponseWriter, value string) {
	w.Header().Set(HeaderContentType, value)
}

func SetContentDisposition(w http.ResponseWriter, value string) {
	w.Header().Set(HeaderContentDisposition, value)
}

func SetAttachment(w http.ResponseWriter, fileName string) {
	SetContentDisposition(w, fmt.Sprintf("attachment; filename=%s", fileName))
}
