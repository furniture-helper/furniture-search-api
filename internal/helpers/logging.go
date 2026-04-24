package helpers

import (
	"furniture-search-api/internal/config"
	"log/slog"
	"net/http"
	"os"
)

func InitLogger() {
	opts := &slog.HandlerOptions{
		Level: config.GetLogLevel(),
	}

	jsonHandler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(jsonHandler)

	if config.GetLogStructure() == "text" {
		textHandler := slog.NewTextHandler(os.Stdout, opts)
		logger = slog.New(textHandler)
	}

	slog.SetDefault(logger)
}

func getBaseLogAttr(r *http.Request, args map[string]any) []any {
	var logAttrs []any

	if r != nil {
		logAttrs = append(logAttrs, slog.String("request_id", GetRequestId(r)))
		logAttrs = append(logAttrs, slog.String("method", r.Method))
		logAttrs = append(logAttrs, slog.String("uri", r.RequestURI))
	}

	for k, v := range args {
		logAttrs = append(logAttrs, slog.Any(k, v))
	}

	return logAttrs
}

func LogError(message string, r *http.Request, err error, args map[string]any) {
	logAttrs := getBaseLogAttr(r, args)

	if err != nil {
		logAttrs = append(logAttrs, slog.String("error", err.Error()))
	}

	slog.Error(message, logAttrs...)
}

func LogInfo(message string, r *http.Request, args map[string]any) {
	logAttrs := getBaseLogAttr(r, args)
	slog.Info(message, logAttrs...)
}

func LogDebug(message string, r *http.Request, args map[string]any) {
	logAttrs := getBaseLogAttr(r, args)
	slog.Debug(message, logAttrs...)
}
