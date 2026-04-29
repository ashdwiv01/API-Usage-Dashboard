package service

import (
	"context"
	"database/sql"
	"fmt"

	"api-rate-limiter/internal/model"
)

// PostgresMetricsRepository implements MetricsRepository using a PostgreSQL database.

// SQL queries for saving and retrieving metrics - steps ->
// Save: Insert a new metric or update existing one for the same api key and bucket timestamp
// rate_limit_metrics is coming from -> api-rate-limiter/internal/migration/001_create_tables.sql

const saveMetricQuery = `
INSERT INTO rate_limit_metrics (api_key, bucket_ts, total, rejected)
VALUES ($1, $2, $3, $4)
ON CONFLICT (api_key, bucket_ts) DO UPDATE
SET total = rate_limit_metrics.total + EXCLUDED.total,
    rejected = rate_limit_metrics.rejected + EXCLUDED.rejected
`

// Get retrieves aggregated metrics for a given API key and time range from the database.
const getMetricsQuery = `
SELECT api_key, bucket_ts, total, rejected
FROM rate_limit_metrics
WHERE api_key = $1
  AND bucket_ts BETWEEN $2 AND $3
ORDER BY bucket_ts
`

// PostgresMetricsRepository implements MetricsRepository using a PostgreSQL database.
type PostgresMetricsRepository struct {
	db *sql.DB
}

func NewPostgresMetricsRepository(db *sql.DB) *PostgresMetricsRepository {
	return &PostgresMetricsRepository{db: db}
}

// Save inserts or updates an aggregated metric in the database.
func (r *PostgresMetricsRepository) Save(ctx context.Context, metric *model.AggregatedMetric) error {
	_, err := r.db.ExecContext(
		ctx,
		saveMetricQuery,
		metric.APIKey,
		metric.Timestamp,
		metric.Total,
		metric.Rejected,
	)
	if err != nil {
		return fmt.Errorf("save metric: %w", err)
	}

	return nil
}

// Get retrieves aggregated metrics for a given API key and time range from the database.
func (r *PostgresMetricsRepository) Get(
	ctx context.Context,
	apiKey string,
	from, to int64,
) ([]model.AggregatedMetric, error) {
	rows, err := r.db.QueryContext(ctx, getMetricsQuery, apiKey, from, to)
	if err != nil {
		return nil, fmt.Errorf("get metrics: %w", err)
	}
	defer rows.Close()

	// Iterate over the result set and scan each row into an AggregatedMetric struct, collecting them into a slice to return.
	var metrics []model.AggregatedMetric
	for rows.Next() {
		var metric model.AggregatedMetric
		if err := rows.Scan(&metric.APIKey, &metric.Timestamp, &metric.Total, &metric.Rejected); err != nil {
			return nil, fmt.Errorf("scan metric: %w", err)
		}
		metrics = append(metrics, metric)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate metrics: %w", err)
	}

	return metrics, nil
}
