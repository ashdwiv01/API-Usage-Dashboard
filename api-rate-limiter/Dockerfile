FROM golang:1.22 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /bin/api-rate-limiter ./cmd/server

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=build /bin/api-rate-limiter /usr/local/bin/api-rate-limiter

EXPOSE 8080

CMD ["api-rate-limiter"]
