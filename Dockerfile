# Build stage
FROM golang:1.23-bookworm AS builder

RUN apt-get update && apt-get install -y \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o stata ./cmd/stata

# Runtime stage
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN useradd -r -u 1000 stata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/stata /app/stata

# Create data directory
RUN mkdir -p /data && chown stata:stata /data

USER stata

VOLUME /data
ENV STATA_DATA_DIR=/data

EXPOSE 8081

ENTRYPOINT ["/app/stata"]
CMD ["serve", "--port", "8081"]
