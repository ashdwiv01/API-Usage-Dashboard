package model

import "time"

type MetricEvent struct {
	APIKey    string
	Allowed   bool
	Timestamp time.Time
}

type AggregatedMetric struct {
	APIKey    string `json:"apiKey"`
	Timestamp int64  `json:"timestamp"`
	Total     int    `json:"total"`
	Rejected  int    `json:"rejected"`
}
