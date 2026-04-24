# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.21-bullseye AS builder

# CGO required for go-sqlite3
ENV CGO_ENABLED=1

WORKDIR /app

COPY go.mod go.sum ./
COPY vendor/ ./vendor/
COPY . .

RUN go build -mod=vendor -ldflags="-s -w" -o /corren .

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /data /etc/corren

COPY --from=builder /corren /usr/local/bin/corren
COPY docker/corren.yaml /etc/corren/corren.yaml
COPY docker/entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh

VOLUME ["/data"]
EXPOSE 3068

ENV CORREN_STORAGE_DIR=/data
ENV CORREN_SERVER_HTTP_BIND_ADDRESS=0.0.0.0:3068

ENTRYPOINT ["/entrypoint.sh"]
