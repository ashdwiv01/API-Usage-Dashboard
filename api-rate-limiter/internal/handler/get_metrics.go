// this file contains the handler for the /metrics endpoint,
// which retrieves aggregated metrics for a given API key and time range.
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// GetMetrics handles GET requests to the /metrics endpoint, validating query parameters and
// returning aggregated metrics in JSON format.
func (h *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	apiKey := r.URL.Query().Get("apiKey")
	if apiKey == "" {
		http.Error(w, "apiKey is required", http.StatusBadRequest)
		return
	}

	// Parse 'from' and 'to' query parameters as unix timestamps, returning errors for invalid formats.
	from, err := strconv.ParseInt(r.URL.Query().Get("from"), 10, 64) // 10, 64 means base 10 and 64-bit integer
	if err != nil {
		http.Error(w, "from must be a valid unix timestamp", http.StatusBadRequest)
		return
	}

	to, err := strconv.ParseInt(r.URL.Query().Get("to"), 10, 64)
	if err != nil {
		http.Error(w, "to must be a valid unix timestamp", http.StatusBadRequest)
		return
	}

	if from > to {
		// reject with 400 Bad Request if 'from' timestamp is greater than 'to' timestamp
		http.Error(w, "from must be less than or equal to to", http.StatusBadRequest)
		return
	}

	if to-from > 3600 {
		// reject with 400 Bad Request if the requested time range exceeds 24 hours
		http.Error(w, "time range cannot exceed 24 hours", http.StatusBadRequest)
		return
	}

	// Retrieve aggregated metrics from the MetricsRepository based on the provided API key and time range,
	//  handling any errors that occur during retrieval.
	metrics, err := h.metricsReader.Get(r.Context(), apiKey, from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the response content type to JSON and encode the retrieved metrics into the response body
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
