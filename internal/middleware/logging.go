package middleware

import (
	"furniture-search-api/internal/helpers"
	"net/http"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		logRequest(r)

		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(lrw, r)

		duration := time.Since(start)
		logResponse(r, duration, lrw.statusCode)
	})
}

func logRequest(r *http.Request) {
	logAttr := map[string]interface{}{
		"method": r.Method,
		"uri":    r.RequestURI,
	}

	helpers.LogDebug("HTTP Request received", r, logAttr)
}

func logResponse(r *http.Request, duration time.Duration, statusCode int) {
	durationMs := float64(duration) / float64(time.Millisecond)

	logAttr := map[string]interface{}{
		"request_id":  helpers.GetRequestId(r),
		"method":      r.Method,
		"uri":         r.RequestURI,
		"status":      statusCode,
		"duration_ms": durationMs,
	}

	if statusCode >= 400 {
		helpers.LogError("HTTP request failed", r, nil, logAttr)
	} else {
		helpers.LogDebug("HTTP request successful", r, logAttr)
	}

}
