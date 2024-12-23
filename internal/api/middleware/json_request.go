package middleware

import (
	"fmt"
	"net/http"

	"github.com/traf72/singbox-api/internal/api"
)

func JsonRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(api.HeaderContentType) != api.ContentTypeJson {
			http.Error(w, fmt.Sprintf(`The "%s" must be "%s"`, api.HeaderContentType, api.ContentTypeJson), http.StatusUnsupportedMediaType)
			return
		}

		next.ServeHTTP(w, r)
	})
}
