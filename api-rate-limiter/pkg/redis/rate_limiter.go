package redis

import (
	"context"
	"fmt"
	"time"

	redis "github.com/redis/go-redis/v9"
)

const luaScript = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

local data = redis.call("HMGET", key, "tokens", "last_refill_ts")

local tokens = tonumber(data[1])
local last_refill = tonumber(data[2])

if tokens == nil then
  tokens = capacity
  last_refill = now
end

local delta = math.max(0, now - last_refill)
local refill = delta * refill_rate
tokens = math.min(capacity, tokens + refill)

local allowed = tokens >= 1
if allowed then
  tokens = tokens - 1
end

redis.call("HMSET", key,
  "tokens", tokens,
  "last_refill_ts", now
)

redis.call("EXPIRE", key, 3600)

if allowed then
  return 1
end

return 0
`

type RedisRateLimiter struct {
	client *redis.Client
	script *redis.Script
}

func NewRedisRateLimiter(client *redis.Client) *RedisRateLimiter {
	script := redis.NewScript(luaScript)
	return &RedisRateLimiter{
		client: client,
		script: script,
	}
}

func (r *RedisRateLimiter) Allow(
	ctx context.Context,
	apiKey string,
	capacity int,
	refillRate float64,
) (bool, error) {
	key := "rate_limit:" + apiKey
	now := time.Now().Unix()

	res, err := r.script.Run(ctx, r.client,
		[]string{key},
		capacity,
		refillRate,
		now,
	).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}

	allowed, ok := res.(int64)
	if !ok {
		return false, fmt.Errorf("invalid response")
	}

	return allowed == 1, nil
}
