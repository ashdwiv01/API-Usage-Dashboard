package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"api-rate-limiter/internal/model"
)

// GetConfigs handles the GET /configs endpoint to retrieve all rate limit configurations.
func (h *Handler) GetConfigs(w http.ResponseWriter, r *http.Request) {
	configs, err := h.configManager.ListConfigs(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, configs)
}

// GetConfig handles the GET /config endpoint to retrieve the rate limit configuration for a specific API key.
func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	apiKey := r.URL.Query().Get("apiKey")
	if apiKey == "" {
		http.Error(w, "apiKey is required", http.StatusBadRequest)
		return
	}

	cfg, err := h.configManager.GetConfig(r.Context(), apiKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, cfg)
}

// UpsertConfig handles the PUT /config endpoint to create or update a rate limit configuration for an API key.
func (h *Handler) UpsertConfig(w http.ResponseWriter, r *http.Request) {
	var cfg model.RateLimitConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	cfg.APIKey = strings.TrimSpace(cfg.APIKey)
	if cfg.APIKey == "" {
		http.Error(w, "apiKey is required", http.StatusBadRequest)
		return
	}

	if cfg.Capacity <= 0 {
		http.Error(w, "capacity must be greater than 0", http.StatusBadRequest)
		return
	}

	if cfg.RefillRate <= 0 {
		http.Error(w, "refillRate must be greater than 0", http.StatusBadRequest)
		return
	}

	saved, err := h.configManager.UpsertConfig(r.Context(), cfg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, saved)
}

// writeJSON is a helper function to write a JSON response with the given status code and value.
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
