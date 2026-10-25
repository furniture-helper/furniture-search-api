package repositories

import (
	"context"
	"errors"
	"fmt"
	"furniture-search-api/internal/config"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	if cfg.URL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.HealthCheckPeriod = 30 * time.Second

	if cfg.UseRDSIAMAuth {
		if poolConfig.ConnConfig.TLSConfig == nil {
			return nil, errors.New("RDS IAM auth requires TLS; set PG_SSLMODE to require, verify-ca, or verify-full")
		}

		refreshLead, err := time.ParseDuration(cfg.IAMRefreshLead)
		if err != nil {
			return nil, fmt.Errorf("parse DB_IAM_REFRESH_LEAD_TIME: %w", err)
		}

		provider, err := newRDSIAMTokenProvider(ctx, poolConfig.ConnConfig, cfg.AWSRegion, refreshLead)
		if err != nil {
			return nil, err
		}

		poolConfig.BeforeConnect = func(ctx context.Context, connConfig *pgx.ConnConfig) error {
			return provider.ApplyToken(ctx, connConfig)
		}

		// RDS IAM tokens expire after 15m. Recycle pooled connections earlier so new
		// connections always authenticate with a fresh token.
		if poolConfig.MaxConnLifetime == 0 || poolConfig.MaxConnLifetime >= 15*time.Minute {
			poolConfig.MaxConnLifetime = 14 * time.Minute
		}
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := Ping(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

func Ping(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return errors.New("postgres pool is nil")
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}

	return nil
}
