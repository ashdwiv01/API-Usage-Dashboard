package service

import (
	"context"

	"api-rate-limiter/internal/model"
)

type RateLimiter interface {
	Allow(ctx context.Context, apiKey string, capacity int, refillRate float64) (bool, error)
}

type ConfigService interface {
	GetConfig(ctx context.Context, apiKey string) (*model.RateLimitConfig, error)
	ListConfigs(ctx context.Context) ([]model.RateLimitConfig, error)                            // ListConfigs retrieves all rate limit configurations.
	UpsertConfig(ctx context.Context, cfg model.RateLimitConfig) (*model.RateLimitConfig, error) // UpsertConfig creates or updates a rate limit configuration for an API key.
}

type MetricsRepository interface {
	Save(ctx context.Context, metric *model.AggregatedMetric) error
	Get(ctx context.Context, apiKey string, from, to int64) ([]model.AggregatedMetric, error)
}
