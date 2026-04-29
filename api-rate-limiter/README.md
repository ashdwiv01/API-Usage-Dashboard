# API Rate Limiter

A small Go backend that checks whether an API key is allowed to make a request.

It uses:

- Postgres to store API key rate-limit configuration
- Redis to store token-bucket state
- Go `net/http` for the `/check` endpoint

## How It Works

When a request hits `/check`, the app:

1. Reads the `x-api-key` header
2. Looks up that key's config in Postgres
3. Uses Redis to apply token-bucket rate limiting
4. Returns whether the request is allowed

## Local Run With Docker

From the `api-rate-limiter` directory:

```bash
docker compose up --build
```

This starts:

- The Go app on `localhost:8080`
- Postgres on `localhost:5432`
- Redis on `localhost:6379`

The compose setup is defined in [docker-compose.yml](./docker-compose.yml), and the app image is built from [Dockerfile](./Dockerfile).

## Seeded API Key

On first Postgres startup, the init script in [db/init.sql](./db/init.sql) creates the config table and inserts:

- `api_key`: `test123`
- `capacity`: `10`
- `refill_rate`: `2`

If you already have an existing Postgres volume and want the init script to run again:

```bash
docker compose down -v
docker compose up --build
```

## Test The Endpoint

Allowed request:

```bash
curl -i -H "x-api-key: test123" http://localhost:8080/check
```

Expected response:

```http
HTTP/1.1 200 OK
{"allowed": true}
```

Unknown API key:

```bash
curl -i -H "x-api-key: unknown" http://localhost:8080/check
```

Expected response: `500 Internal Server Error`

Rate-limit exhaustion:

```bash
for i in {1..12}; do curl -s -o /dev/null -w "%{http_code}\n" -H "x-api-key: test123" http://localhost:8080/check; done
```

Expected behavior:

- Early requests return `200`
- After the bucket is exhausted, requests return `429`

## Run Without Docker

You can also run the backend directly if Postgres and Redis are already available.

Required environment variables:

- `DATABASE_URL`
- `REDIS_ADDR`

Optional:

- `PORT` default: `8080`

Example:

```bash
DATABASE_URL="postgres://postgres:postgres@localhost:5432/rate_limiter?sslmode=disable" \
REDIS_ADDR="127.0.0.1:6379" \
PORT="8080" \
/usr/local/go/bin/go run ./cmd/server
```

## Database Schema

The config table schema is defined in [db/schema.sql](./db/schema.sql):

```sql
CREATE TABLE IF NOT EXISTS rate_limit_configs (
    api_key TEXT PRIMARY KEY,
    capacity INTEGER NOT NULL CHECK (capacity > 0),
    refill_rate DOUBLE PRECISION NOT NULL CHECK (refill_rate > 0)
);
```

## Tests

Run the Go tests with:

```bash
/usr/local/go/bin/go test ./...
```
