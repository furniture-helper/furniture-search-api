package main

import (
	"context"
	"furniture-search-api/internal/config"
	"furniture-search-api/internal/helpers"
	"furniture-search-api/internal/repositories"
	"os"
	"time"
)

func main() {
	helpers.InitLogger()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	cfg := config.GetDatabaseConfig(ctx)
	defer cancel()

	pool, err := repositories.NewPostgresPool(ctx, cfg)
	if err != nil {
		helpers.LogError("Failed to connect to postgres", nil, err, nil)
		os.Exit(1)
	}
	defer pool.Close()

	helpers.LogInfo("Postgres connection successful", nil, map[string]any{
		"max_conns": cfg.MaxConns,
		"min_conns": cfg.MinConns,
	})
}
