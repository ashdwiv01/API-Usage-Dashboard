package service

import (
	"context"
	"time"

	"api-rate-limiter/internal/model"
)

type RateLimiterService struct {
	limiter   RateLimiter
	configSvc ConfigService
	metricsCh chan<- model.MetricEvent
}

func NewRateLimiterService(
	limiter RateLimiter,
	configSvc ConfigService,
	metricsCh chan<- model.MetricEvent,
) *RateLimiterService {
	return &RateLimiterService{
		limiter:   limiter,
		configSvc: configSvc,
		metricsCh: metricsCh,
	}
}

func (s *RateLimiterService) Allow(ctx context.Context, apiKey string) (bool, error) {
	cfg, err := s.configSvc.GetConfig(ctx, apiKey)
	if err != nil {
		return false, err
	}

	allowed, err := s.limiter.Allow(ctx, apiKey, cfg.Capacity, cfg.RefillRate)

	// If the request is allowed and the metrics channel is configured,
	// send a MetricEvent to the channel for aggregation and storage.
	if err == nil && s.metricsCh != nil {
		select {
		// Non-blocking send to the metrics channel to avoid delaying the response if the channel is full.
		case s.metricsCh <- model.MetricEvent{
			APIKey:    apiKey,
			Allowed:   allowed,
			Timestamp: time.Now(),
		}:
		// If the channel is full, we skip sending the metric to avoid blocking the request processing.
		default:
		}
	}

	return allowed, err
}
