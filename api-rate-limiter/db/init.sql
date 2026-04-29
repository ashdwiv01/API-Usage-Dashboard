CREATE TABLE IF NOT EXISTS rate_limit_configs (
    api_key TEXT PRIMARY KEY,
    capacity INTEGER NOT NULL CHECK (capacity > 0),
    refill_rate DOUBLE PRECISION NOT NULL CHECK (refill_rate > 0)
);

CREATE TABLE IF NOT EXISTS rate_limit_metrics (
    api_key TEXT NOT NULL,
    bucket_ts BIGINT NOT NULL,
    total INTEGER NOT NULL CHECK (total >= 0),
    rejected INTEGER NOT NULL CHECK (rejected >= 0),
    PRIMARY KEY (api_key, bucket_ts)
);

INSERT INTO rate_limit_configs (api_key, capacity, refill_rate)
VALUES ('test123', 10, 2)
ON CONFLICT (api_key) DO UPDATE
SET capacity = EXCLUDED.capacity,
    refill_rate = EXCLUDED.refill_rate;
