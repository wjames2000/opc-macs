# ===== Stage 1: Build =====
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /src

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build main binary
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /build/opc-agent ./cmd/opc-agent/

# Build plugins (for Linux .so deployment)
RUN mkdir -p /build/plugins && \
    cd /src/plugins/copywriter && CGO_ENABLED=1 go build -buildmode=plugin -o /build/plugins/copywriter.so . 2>/dev/null || true && \
    cd /src/plugins/email_sorter && CGO_ENABLED=1 go build -buildmode=plugin -o /build/plugins/email_sorter.so . 2>/dev/null || true && \
    cd /src/plugins/xhs_poster && CGO_ENABLED=1 go build -buildmode=plugin -o /build/plugins/xhs_poster.so . 2>/dev/null || true

# ===== Stage 2: Runtime =====
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary and plugins
COPY --from=builder /build/opc-agent .
COPY --from=builder /build/plugins/ ./plugins/
COPY config.yaml ./config.yaml

# Create data directory for memory persistence
RUN mkdir -p /data

# Health check
EXPOSE 9090
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD ./opc-agent --health || exit 1

# Run
ENTRYPOINT ["./opc-agent"]
CMD ["--config", "config.yaml"]
