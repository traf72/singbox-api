package middleware

import (
	"net/http"
)

type Middleware func(next http.Handler) http.Handler

func Chain(ms ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(ms) - 1; i >= 0; i-- {
			h := ms[i]
			next = h(next)
		}

		return next
	}
}
