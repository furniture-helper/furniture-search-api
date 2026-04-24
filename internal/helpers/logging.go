package helpers

import (
	"context"
	"furniture-search-api/internal/config"
	"log/slog"
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

func getBaseLogAttr(ctx context.Context, args map[string]any) []any {
	var logAttrs []any

	if ctx != nil {
		logAttrs = append(logAttrs, slog.String("request_id", GetRequestIdFromContext(ctx)))
	}

	for k, v := range args {
		logAttrs = append(logAttrs, slog.Any(k, v))
	}

	return logAttrs
}

func LogError(message string, ctx context.Context, err error, args map[string]any) {
	logAttrs := getBaseLogAttr(ctx, args)

	if err != nil {
		logAttrs = append(logAttrs, slog.String("error", err.Error()))
	}

	slog.Error(message, logAttrs...)
}

func LogInfo(message string, ctx context.Context, args map[string]any) {
	logAttrs := getBaseLogAttr(ctx, args)
	slog.Info(message, logAttrs...)
}

func LogDebug(message string, ctx context.Context, args map[string]any) {
	logAttrs := getBaseLogAttr(ctx, args)
	slog.Debug(message, logAttrs...)
}
