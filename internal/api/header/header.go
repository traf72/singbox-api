package header

import (
	"fmt"
	"net/http"
)

const (
	ContentType        = "Content-Type"
	ContentDisposition = "Content-Disposition"
)

const (
	ContentTypeJson      = "application/json"
	ContentTypeTextPlain = "text/plain"
)

func SetContentType(w http.ResponseWriter, value string) {
	w.Header().Set(ContentType, value)
}

func SetContentDisposition(w http.ResponseWriter, value string) {
	w.Header().Set(ContentDisposition, value)
}

func SetAttachment(w http.ResponseWriter, fileName string) {
	SetContentDisposition(w, fmt.Sprintf("attachment; filename=%s", fileName))
}
