package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"api-rate-limiter/internal/handler"
	"api-rate-limiter/internal/model"
	"api-rate-limiter/internal/service"
	redispkg "api-rate-limiter/pkg/redis"

	_ "github.com/lib/pq"
	redis "github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()
	redisAddr := getenv("REDIS_ADDR", "127.0.0.1:6379")
	port := getenv("PORT", "8080")
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL must be set")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Verify database connection
	if err := db.PingContext(ctx); err != nil {
		log.Fatal(err)
	}

	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	limiter := redispkg.NewRedisRateLimiter(redisClient)

	configSvc := service.NewPostgresConfigService(db)
	metricsRepo := service.NewPostgresMetricsRepository(db)
	metricsCh := make(chan model.MetricEvent, 1000)
	metricsAggregator := service.NewMetricsAggregator(metricsCh, metricsRepo, time.Second)
	metricsAggregator.Start(ctx)

	rateLimiterService := service.NewRateLimiterService(limiter, configSvc, metricsCh)
	h := handler.NewHandler(rateLimiterService, metricsRepo, configSvc)

	mux := http.NewServeMux()
	mux.HandleFunc("/check", h.CheckRateLimit)
	mux.HandleFunc("/metrics", h.GetMetrics)
	mux.HandleFunc("/configs", h.GetConfigs)
	mux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// For GET requests, we call GetConfig to retrieve the configuration for a specific API key.
			h.GetConfig(w, r)
		case http.MethodPut:
			// For PUT requests, we call UpsertConfig to create or update the rate limit configuration for an API key.
			h.UpsertConfig(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	addr := ":" + port
	log.Printf("server listening on %s (redis: %s, postgres configured)", addr, redisAddr)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

func getenv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, x-api-key")
		w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
