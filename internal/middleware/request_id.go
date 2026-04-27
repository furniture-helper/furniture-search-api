package middleware

import (
	"furniture-search-api/internal/helpers"
	"net/http"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/google/uuid"
)

func RequestIdMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requestId string

		lc, ok := lambdacontext.FromContext(r.Context())
		if ok {
			requestId = lc.AwsRequestID
		} else {
			requestId = uuid.New().String()
		}

		r.Header.Set(helpers.RequestIDHeader, requestId)
		r = r.WithContext(helpers.WithRequestId(r.Context(), requestId))
		w.Header().Set(helpers.RequestIDHeader, requestId)

		next.ServeHTTP(w, r)
	})
}
