package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"api-rate-limiter/internal/model"
)

type fakeRow struct {
	scan func(dest ...any) error
}

func (r fakeRow) Scan(dest ...any) error {
	return r.scan(dest...)
}

func TestPostgresConfigServiceGetConfig(t *testing.T) {
	svc := &PostgresConfigService{
		queryRow: func(ctx context.Context, query string, args ...any) rowScanner {
			return fakeRow{
				scan: func(dest ...any) error {
					*(dest[0].(*string)) = "client-a"
					*(dest[1].(*int)) = 7
					*(dest[2].(*float64)) = 1.25
					return nil
				},
			}
		},
	}

	cfg, err := svc.GetConfig(context.Background(), "client-a")
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	if cfg.APIKey != "client-a" {
		t.Fatalf("expected APIKey client-a, got %q", cfg.APIKey)
	}

	if cfg.Capacity != 7 {
		t.Fatalf("expected capacity 7, got %d", cfg.Capacity)
	}

	if cfg.RefillRate != 1.25 {
		t.Fatalf("expected refill rate 1.25, got %v", cfg.RefillRate)
	}
}

func TestPostgresConfigServiceGetConfigMissingKey(t *testing.T) {
	svc := &PostgresConfigService{
		queryRow: func(ctx context.Context, query string, args ...any) rowScanner {
			return fakeRow{
				scan: func(dest ...any) error {
					return sql.ErrNoRows
				},
			}
		},
	}

	_, err := svc.GetConfig(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected missing key error")
	}
}

func TestPostgresConfigServiceGetConfigWrapsDatabaseError(t *testing.T) {
	svc := &PostgresConfigService{
		queryRow: func(ctx context.Context, query string, args ...any) rowScanner {
			return fakeRow{
				scan: func(dest ...any) error {
					return errors.New("db unavailable")
				},
			}
		},
	}

	_, err := svc.GetConfig(context.Background(), "client-a")
	if err == nil {
		t.Fatal("expected database error")
	}
}

type fakeRows struct {
	configs []model.RateLimitConfig
	index   int
	err     error
}

func (r *fakeRows) Next() bool {
	return r.index < len(r.configs)
}

func (r *fakeRows) Scan(dest ...any) error {
	cfg := r.configs[r.index]
	r.index++
	*(dest[0].(*string)) = cfg.APIKey
	*(dest[1].(*int)) = cfg.Capacity
	*(dest[2].(*float64)) = cfg.RefillRate
	return nil
}

func (r *fakeRows) Close() error {
	return nil
}

func (r *fakeRows) Err() error {
	return r.err
}

// Additional tests for ListConfigs can be added here, following a similar pattern to the above tests.
func TestPostgresConfigServiceListConfigs(t *testing.T) {
	svc := &PostgresConfigService{
		query: func(ctx context.Context, query string, args ...any) (rowsScanner, error) {
			return &fakeRows{
				configs: []model.RateLimitConfig{
					{APIKey: "client-a", Capacity: 5, RefillRate: 1.5},
					{APIKey: "client-b", Capacity: 8, RefillRate: 2.0},
				},
			}, nil
		},
	}

	configs, err := svc.ListConfigs(context.Background())
	if err != nil {
		t.Fatalf("ListConfigs() error = %v", err)
	}

	if len(configs) != 2 {
		t.Fatalf("expected 2 configs, got %d", len(configs))
	}
}

// Additional tests for UpsertConfig can be added here, following a similar pattern to the above tests.
func TestPostgresConfigServiceUpsertConfig(t *testing.T) {
	svc := &PostgresConfigService{
		queryRow: func(ctx context.Context, query string, args ...any) rowScanner {
			return fakeRow{
				scan: func(dest ...any) error {
					*(dest[0].(*string)) = args[0].(string)
					*(dest[1].(*int)) = args[1].(int)
					*(dest[2].(*float64)) = args[2].(float64)
					return nil
				},
			}
		},
	}

	cfg, err := svc.UpsertConfig(context.Background(), model.RateLimitConfig{
		APIKey:     "client-a",
		Capacity:   12,
		RefillRate: 3.5,
	})
	if err != nil {
		t.Fatalf("UpsertConfig() error = %v", err)
	}

	if cfg.Capacity != 12 || cfg.RefillRate != 3.5 {
		t.Fatalf("unexpected saved config: %+v", cfg)
	}
}
