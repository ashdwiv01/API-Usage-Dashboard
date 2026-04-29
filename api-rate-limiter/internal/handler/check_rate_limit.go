package handler

import "net/http"

func (h *Handler) CheckRateLimit(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get("x-api-key")

	allowed, err := h.service.Allow(r.Context(), apiKey)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if !allowed {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"allowed": false}`))
		return
	}

	_, _ = w.Write([]byte(`{"allowed": true}`))
}
