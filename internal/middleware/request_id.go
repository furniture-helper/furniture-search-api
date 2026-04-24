package middleware

import (
	"net/http"

	"github.com/google/uuid"
)

func RequestIdMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestId := uuid.New()
		r.Header.Set("X-Request-Id", requestId.String())

		next.ServeHTTP(w, r)
	})
}
