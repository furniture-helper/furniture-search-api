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

func cloneHeaders(headers http.Header) map[string][]string {
	cloned := make(map[string][]string, len(headers))

	for key, values := range headers {
		copiedValues := make([]string, len(values))
		copy(copiedValues, values)
		cloned[key] = copiedValues
	}

	return cloned
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
		logResponse(r, duration, lrw.statusCode, lrw.Header())
	})
}

func logRequest(r *http.Request) {
	logAttr := map[string]interface{}{
		"method":          r.Method,
		"uri":             r.RequestURI,
		"request_headers": cloneHeaders(r.Header),
	}

	helpers.LogDebug("HTTP Request received", r.Context(), logAttr)
}

func logResponse(r *http.Request, duration time.Duration, statusCode int, responseHeaders http.Header) {
	durationMs := float64(duration) / float64(time.Millisecond)

	logAttr := map[string]interface{}{
		"status":           statusCode,
		"duration_ms":      durationMs,
		"response_headers": cloneHeaders(responseHeaders),
	}

	if statusCode >= 400 {
		helpers.LogError("HTTP request failed", r.Context(), nil, logAttr)
	} else {
		helpers.LogDebug("HTTP request successful", r.Context(), logAttr)
	}

}
