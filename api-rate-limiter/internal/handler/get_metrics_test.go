package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"api-rate-limiter/internal/model"
)

type fakeMetricsReader struct {
	metrics []model.AggregatedMetric
	err     error
}

func (r *fakeMetricsReader) Get(ctx context.Context, apiKey string, from, to int64) ([]model.AggregatedMetric, error) {
	return r.metrics, r.err
}

type fakeRateLimitService struct{}

func (s *fakeRateLimitService) Allow(ctx context.Context, apiKey string) (bool, error) {
	return true, nil
}

type fakeConfigManager struct {
	configs []model.RateLimitConfig
	config  *model.RateLimitConfig
	err     error
}

func (m *fakeConfigManager) GetConfig(ctx context.Context, apiKey string) (*model.RateLimitConfig, error) {
	return m.config, m.err
}

func (m *fakeConfigManager) ListConfigs(ctx context.Context) ([]model.RateLimitConfig, error) {
	return m.configs, m.err
}

func (m *fakeConfigManager) UpsertConfig(ctx context.Context, cfg model.RateLimitConfig) (*model.RateLimitConfig, error) {
	return &cfg, m.err
}

func TestGetMetrics(t *testing.T) {
	h := NewHandler(&fakeRateLimitService{}, &fakeMetricsReader{
		metrics: []model.AggregatedMetric{
			{
				APIKey:    "test123",
				Timestamp: 1710000000,
				Total:     120,
				Rejected:  35,
			},
		},
	}, &fakeConfigManager{})

	req := httptest.NewRequest(http.MethodGet, "/metrics?apiKey=test123&from=1710000000&to=1710000010", nil)
	rr := httptest.NewRecorder()

	h.GetMetrics(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
}

func TestGetConfigs(t *testing.T) {
	h := NewHandler(&fakeRateLimitService{}, &fakeMetricsReader{}, &fakeConfigManager{
		configs: []model.RateLimitConfig{
			{APIKey: "test123", Capacity: 10, RefillRate: 2},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/configs", nil)
	rr := httptest.NewRecorder()

	h.GetConfigs(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
}

func TestUpsertConfig(t *testing.T) {
	h := NewHandler(&fakeRateLimitService{}, &fakeMetricsReader{}, &fakeConfigManager{})

	req := httptest.NewRequest(
		http.MethodPut,
		"/config",
		strings.NewReader(`{"apiKey":"test123","capacity":25,"refillRate":4}`),
	)
	rr := httptest.NewRecorder()

	h.UpsertConfig(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
}
