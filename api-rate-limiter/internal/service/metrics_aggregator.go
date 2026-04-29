/*
What does a MetricsAggregator do?

A MetricsAggregator is responsible for consuming MetricEvents, aggregating them into AggregatedMetrics,
 and periodically flushing the aggregated metrics to a storage repository.
 It maintains an in-memory buffer to count total and rejected requests by API key and timestamp,
  and uses a ticker to trigger regular flushes to the storage.
*/

package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"api-rate-limiter/internal/model"
)

// MetricsAggregator aggregates MetricEvents into AggregatedMetrics and periodically flushes them to storage.
type MetricsAggregator struct {
	ch       <-chan model.MetricEvent // Channel to receive MetricEvents
	storage  MetricsRepository        // Repository to save aggregated metrics
	interval time.Duration            // Flush interval

	mu          sync.Mutex                         // Buffer to hold aggregated metrics before flushing
	buffer      map[string]*model.AggregatedMetric // Keyed by
	consumeDone chan struct{}
}

// NewMetricsAggregator creates a new MetricsAggregator.
func NewMetricsAggregator(
	ch <-chan model.MetricEvent,
	storage MetricsRepository,
	interval time.Duration,
) *MetricsAggregator {
	return &MetricsAggregator{
		ch:          ch,
		storage:     storage,
		interval:    interval,
		buffer:      make(map[string]*model.AggregatedMetric),
		consumeDone: make(chan struct{}),
	}
}

// Start begins consuming MetricEvents and flushing aggregated metrics at intervals.
func (m *MetricsAggregator) Start(ctx context.Context) {
	go func() {
		defer close(m.consumeDone)
		m.consume(ctx)
	}()
	go m.flushWorker(ctx)
}

func (m *MetricsAggregator) drain() {
	for {
		select {
		case event, ok := <-m.ch:
			if !ok {
				// Channel is closed, nothing left to drain
				return
			}
			m.bufferEvent(event)
		default:
			// No more events immediately available in the channel
			return
		}
	}
}

// consume reads MetricEvents from the channel and buffers them for aggregation.
func (m *MetricsAggregator) consume(ctx context.Context) {
	// Go worker loop
	for {
		select {
		// Exit if context is canceled
		case <-ctx.Done():
			// Drain any remaining events in the channel before exiting
			m.drain()
			return
		// Process incoming MetricEvent
		case event, ok := <-m.ch:
			if !ok {
				return
			}
			// Buffer the event for aggregation
			m.bufferEvent(event)
		}
	}
}

// bufferEvent updates the in-memory buffer with the incoming MetricEvent, aggregating counts by API key and timestamp.
func (m *MetricsAggregator) bufferEvent(event model.MetricEvent) {
	ts := event.Timestamp.Unix() // Round to the nearest second for aggregation
	key := fmt.Sprintf("%s:%d", event.APIKey, ts)

	// Lock the buffer for safe concurrent access
	m.mu.Lock()
	defer m.mu.Unlock()

	// Retrieve or create the aggregated metric for the given API key and timestamp
	metric, exists := m.buffer[key]
	if !exists {
		metric = &model.AggregatedMetric{
			APIKey:    event.APIKey,
			Timestamp: ts,
		}
		m.buffer[key] = metric
	}

	// Update the total and rejected counts based on the MetricEvent
	metric.Total++
	if !event.Allowed {
		metric.Rejected++
	}
}

// flushWorker periodically flushes the aggregated metrics to storage until the context is canceled.
func (m *MetricsAggregator) flushWorker(ctx context.Context) {
	// Use a ticker to trigger flushes at regular intervals
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		// Exit if context is canceled
		case <-ctx.Done():
			// Wait for the consume worker to finish processing any remaining events before flushing
			<-m.consumeDone
			m.flush(ctx)
			return
		// Flush the aggregated metrics at each tick
		case <-ticker.C:
			m.flush(ctx)
		}
	}
}

// flush saves the current buffer of aggregated metrics to storage and resets the buffer.
func (m *MetricsAggregator) flush(ctx context.Context) {
	// -- Lock the buffer -- swap it with a new empty buffer to allow new events to be buffered while flushing
	m.mu.Lock()
	data := m.buffer
	m.buffer = make(map[string]*model.AggregatedMetric)
	m.mu.Unlock()

	for _, metric := range data {
		if err := m.storage.Save(ctx, metric); err != nil {
			log.Printf(
				"flush metric failed: api_key=%s bucket_ts=%d total=%d rejected=%d err=%v",
				metric.APIKey,
				metric.Timestamp,
				metric.Total,
				metric.Rejected,
				err,
			)
		}
	}
}
