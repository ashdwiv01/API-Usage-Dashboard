package handler

import (
	"context"

	"api-rate-limiter/internal/model"
)

type RateLimitService interface {
	Allow(ctx context.Context, apiKey string) (bool, error)
}

type MetricsReader interface {
	Get(ctx context.Context, apiKey string, from, to int64) ([]model.AggregatedMetric, error)
}

type ConfigManager interface {
	GetConfig(ctx context.Context, apiKey string) (*model.RateLimitConfig, error)
	ListConfigs(ctx context.Context) ([]model.RateLimitConfig, error)
	UpsertConfig(ctx context.Context, cfg model.RateLimitConfig) (*model.RateLimitConfig, error)
}

type Handler struct {
	service       RateLimitService
	metricsReader MetricsReader
	configManager ConfigManager
}

func NewHandler(service RateLimitService, metricsReader MetricsReader, configManager ConfigManager) *Handler {
	return &Handler{
		service:       service,
		metricsReader: metricsReader,
		configManager: configManager,
	}
}
