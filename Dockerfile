# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

RUN apk --no-cache add git

WORKDIR /app

# Download deps in a separate layer so Docker cache isn't busted on every code change
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build both binaries
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /app/bin/erp-server  ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /app/bin/erp-migrate ./cmd/migrate

# ── Stage 2: Run ──────────────────────────────────────────────────────────────
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy compiled binaries
COPY --from=builder /app/bin/erp-server  ./erp-server
COPY --from=builder /app/bin/erp-migrate ./erp-migrate

# Copy config and migration files (needed at runtime)
COPY --from=builder /app/config ./config
COPY --from=builder /app/internal/infrastructure/database/migrations \
                    ./internal/infrastructure/database/migrations

# Copy entrypoint script
COPY scripts/docker-entrypoint.sh ./docker-entrypoint.sh
RUN chmod +x ./docker-entrypoint.sh

EXPOSE 3000

ENTRYPOINT ["./docker-entrypoint.sh"]
