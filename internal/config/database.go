package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	defaultDBMaxConns int32 = 10
	defaultDBMinConns int32 = 1

	defaultDBHost     = "localhost"
	defaultDBPort     = "5432"
	defaultDBUsername = "postgres"
	defaultDBName     = "postgres"
	defaultDBSSLMode  = "disable"
)

type DatabaseConfig struct {
	URL      string
	MaxConns int32
	MinConns int32
}

func GetDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		URL:      buildDatabaseURL(),
		MaxConns: getInt32Env("DB_MAX_CONNS", defaultDBMaxConns),
		MinConns: getInt32Env("DB_MIN_CONNS", defaultDBMinConns),
	}
}

func buildDatabaseURL() string {
	host := getStringEnv("PG_HOST", defaultDBHost)
	port := getStringEnv("PG_PORT", defaultDBPort)
	username := getStringEnv("PG_USER", defaultDBUsername)
	password := strings.TrimSpace(os.Getenv("PG_PASSWORD"))
	database := getStringEnv("PG_DATABASE", defaultDBName)
	sslmode := getStringEnv("PG_SSLMODE", defaultDBSSLMode)

	connURL := &url.URL{
		Scheme: "postgres",
		Host:   fmt.Sprintf("%s:%s", host, port),
		Path:   database,
	}

	connURL.User = url.UserPassword(username, password)

	query := url.Values{}
	query.Set("sslmode", sslmode)
	connURL.RawQuery = query.Encode()

	return connURL.String()
}

func getInt32Env(key string, fallback int32) int32 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return int32(parsed)
}

func getStringEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}
