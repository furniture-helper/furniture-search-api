package middleware

import (
	"furniture-search-api/internal/helpers"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func RequestIdMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestId := strings.TrimSpace(r.Header.Get(helpers.RequestIDHeader))
		if requestId == "" {
			requestId = uuid.NewString()
		}

		r.Header.Set(helpers.RequestIDHeader, requestId)
		r = r.WithContext(helpers.WithRequestId(r.Context(), requestId))
		w.Header().Set(helpers.RequestIDHeader, requestId)

		next.ServeHTTP(w, r)
	})
}
