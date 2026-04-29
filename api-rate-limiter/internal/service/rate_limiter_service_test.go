package service

import (
	"context"
	"testing"
	"time"

	"api-rate-limiter/internal/model"
)

type fakeRedisLimiter struct {
	allowed bool
	err     error
}

func (l *fakeRedisLimiter) Allow(
	ctx context.Context,
	apiKey string,
	capacity int,
	refillRate float64,
) (bool, error) {
	return l.allowed, l.err
}

type fakeConfigService struct {
	config *model.RateLimitConfig
	err    error
}

func (s *fakeConfigService) GetConfig(ctx context.Context, apiKey string) (*model.RateLimitConfig, error) {
	return s.config, s.err
}

func (s *fakeConfigService) ListConfigs(ctx context.Context) ([]model.RateLimitConfig, error) {
	if s.config == nil {
		return nil, s.err
	}
	return []model.RateLimitConfig{*s.config}, s.err
}

func (s *fakeConfigService) UpsertConfig(ctx context.Context, cfg model.RateLimitConfig) (*model.RateLimitConfig, error) {
	s.config = &cfg
	return &cfg, s.err
}

func TestRateLimiterServiceEmitsMetricEvent(t *testing.T) {
	metricsCh := make(chan model.MetricEvent, 1)
	service := &RateLimiterService{
		limiter: &fakeRedisLimiter{
			allowed: true,
		},
		configSvc: &fakeConfigService{
			config: &model.RateLimitConfig{
				APIKey:     "test123",
				Capacity:   10,
				RefillRate: 2,
			},
		},
		metricsCh: metricsCh,
	}

	allowed, err := service.Allow(context.Background(), "test123")
	if err != nil {
		t.Fatalf("Allow() error = %v", err)
	}

	if !allowed {
		t.Fatal("expected request to be allowed")
	}

	select {
	case event := <-metricsCh:
		if event.APIKey != "test123" {
			t.Fatalf("expected API key test123, got %q", event.APIKey)
		}
		if !event.Allowed {
			t.Fatal("expected emitted event to be allowed")
		}
		if time.Since(event.Timestamp) > time.Second {
			t.Fatal("expected recent timestamp on emitted event")
		}
	default:
		t.Fatal("expected metric event to be emitted")
	}
}
