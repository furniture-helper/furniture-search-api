package helpers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

const RequestIDHeader = "X-Request-Id"

type requestIDContextKey string

const requestIDKey requestIDContextKey = "request_id"

func WriteJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func WithRequestId(ctx context.Context, requestId string) context.Context {
	return context.WithValue(ctx, requestIDKey, strings.TrimSpace(requestId))
}

func GetRequestIdFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	requestId, ok := ctx.Value(requestIDKey).(string)
	if !ok {
		return ""
	}

	return strings.TrimSpace(requestId)
}
