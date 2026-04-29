package service

import (
	"context"
	"testing"
	"time"

	"api-rate-limiter/internal/model"
)

type fakeMetricsRepository struct {
	saved []*model.AggregatedMetric
}

func (r *fakeMetricsRepository) Save(ctx context.Context, metric *model.AggregatedMetric) error {
	r.saved = append(r.saved, metric)
	return nil
}

func (r *fakeMetricsRepository) Get(ctx context.Context, apiKey string, from, to int64) ([]model.AggregatedMetric, error) {
	var metrics []model.AggregatedMetric
	for _, metric := range r.saved {
		if metric.APIKey != apiKey {
			continue
		}
		if metric.Timestamp < from || metric.Timestamp > to {
			continue
		}
		metrics = append(metrics, *metric)
	}
	return metrics, nil
}

func TestMetricsAggregatorFlushAggregatesBySecond(t *testing.T) {
	repo := &fakeMetricsRepository{}
	aggregator := NewMetricsAggregator(make(chan model.MetricEvent), repo, time.Second)

	ts := time.Unix(1710000000, 0)
	aggregator.bufferEvent(model.MetricEvent{APIKey: "test123", Allowed: true, Timestamp: ts})
	aggregator.bufferEvent(model.MetricEvent{APIKey: "test123", Allowed: false, Timestamp: ts})

	aggregator.flush(context.Background())

	if len(repo.saved) != 1 {
		t.Fatalf("expected 1 saved metric, got %d", len(repo.saved))
	}

	if repo.saved[0].Total != 2 {
		t.Fatalf("expected total 2, got %d", repo.saved[0].Total)
	}

	if repo.saved[0].Rejected != 1 {
		t.Fatalf("expected rejected 1, got %d", repo.saved[0].Rejected)
	}
}

func TestMetricsFlowEmitsAggregatesAndReads(t *testing.T) {
	metricsCh := make(chan model.MetricEvent, 1)
	repo := &fakeMetricsRepository{}
	aggregator := NewMetricsAggregator(metricsCh, repo, time.Second)

	service := &RateLimiterService{
		limiter: &fakeRedisLimiter{
			allowed: false,
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
	if allowed {
		t.Fatal("expected request to be rejected by limiter")
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		aggregator.consume(ctx)
	}()

	time.Sleep(10 * time.Millisecond)
	cancel()
	<-done

	aggregator.flush(context.Background())

	metrics, err := repo.Get(context.Background(), "test123", 0, time.Now().Unix())
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}

	if metrics[0].Total != 1 {
		t.Fatalf("expected total 1, got %d", metrics[0].Total)
	}

	if metrics[0].Rejected != 1 {
		t.Fatalf("expected rejected 1, got %d", metrics[0].Rejected)
	}
}
