package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"api-rate-limiter/internal/model"
)

const getRateLimitConfigQuery = `
SELECT api_key, capacity, refill_rate
FROM rate_limit_configs
WHERE api_key = $1
`

const listRateLimitConfigsQuery = `
SELECT api_key, capacity, refill_rate
FROM rate_limit_configs
ORDER BY api_key
`

const upsertRateLimitConfigQuery = `
INSERT INTO rate_limit_configs (api_key, capacity, refill_rate)
VALUES ($1, $2, $3)
ON CONFLICT (api_key) DO UPDATE
SET capacity = EXCLUDED.capacity,
    refill_rate = EXCLUDED.refill_rate
RETURNING api_key, capacity, refill_rate
`

type rowScanner interface {
	Scan(dest ...any) error
}

type rowsScanner interface {
	Next() bool
	Scan(dest ...any) error
	Close() error
	Err() error
}

type PostgresConfigService struct {
	queryRow func(ctx context.Context, query string, args ...any) rowScanner
	query    func(ctx context.Context, query string, args ...any) (rowsScanner, error)
}

func NewPostgresConfigService(db *sql.DB) *PostgresConfigService {
	return &PostgresConfigService{
		queryRow: func(ctx context.Context, query string, args ...any) rowScanner {
			return db.QueryRowContext(ctx, query, args...)
		},
		query: func(ctx context.Context, query string, args ...any) (rowsScanner, error) {
			return db.QueryContext(ctx, query, args...)
		},
	}
}

func (s *PostgresConfigService) GetConfig(ctx context.Context, apiKey string) (*model.RateLimitConfig, error) {
	cfg := model.RateLimitConfig{}
	err := s.queryRow(ctx, getRateLimitConfigQuery, apiKey).Scan(&cfg.APIKey, &cfg.Capacity, &cfg.RefillRate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("api key not configured: %s", apiKey)
		}
		return nil, fmt.Errorf("query rate limit config: %w", err)
	}

	return &cfg, nil
}

// ListConfigs retrieves all rate limit configurations from the database and returns them as a 
// slice of RateLimitConfig structs.
func (s *PostgresConfigService) ListConfigs(ctx context.Context) ([]model.RateLimitConfig, error) {
	rows, err := s.query(ctx, listRateLimitConfigsQuery)
	if err != nil {
		return nil, fmt.Errorf("list rate limit configs: %w", err)
	}
	defer rows.Close()

	var configs []model.RateLimitConfig
	for rows.Next() {
		var cfg model.RateLimitConfig
		if err := rows.Scan(&cfg.APIKey, &cfg.Capacity, &cfg.RefillRate); err != nil {
			return nil, fmt.Errorf("scan rate limit config: %w", err)
		}
		configs = append(configs, cfg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rate limit configs: %w", err)
	}

	return configs, nil
}

// UpsertConfig creates or updates a rate limit configuration for an API key in the database and 
// returns the saved configuration.
func (s *PostgresConfigService) UpsertConfig(
	ctx context.Context,
	cfg model.RateLimitConfig,
) (*model.RateLimitConfig, error) {
	saved := model.RateLimitConfig{}
	err := s.queryRow(
		ctx,
		upsertRateLimitConfigQuery,
		cfg.APIKey,
		cfg.Capacity,
		cfg.RefillRate,
	).Scan(&saved.APIKey, &saved.Capacity, &saved.RefillRate)
	if err != nil {
		return nil, fmt.Errorf("upsert rate limit config: %w", err)
	}

	return &saved, nil
}
