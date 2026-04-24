package config

import (
	"log/slog"
	"os"
)

func GetLogLevel() slog.Level {
	if os.Getenv("LOG_LEVEL") == "DEBUG" {
		return slog.LevelDebug
	}
	return slog.LevelInfo
}

func GetLogStructure() string {
	if os.Getenv("LOG_STRUCTURE") == "text" {
		return "text"
	}
	return "json"
}
