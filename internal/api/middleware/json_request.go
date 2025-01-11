package middleware

import (
	"fmt"
	"net/http"

	"github.com/traf72/singbox-api/internal/api/header"
)

func JsonRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(header.ContentType) != header.ContentTypeJson {
			http.Error(w, fmt.Sprintf(`The "%s" must be "%s"`, header.ContentType, header.ContentTypeJson), http.StatusUnsupportedMediaType)
			return
		}

		next.ServeHTTP(w, r)
	})
}
