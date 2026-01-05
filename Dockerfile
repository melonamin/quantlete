# Build stage - Web assets
FROM node:22-alpine AS web-builder

WORKDIR /app/web
COPY web/package.json web/yarn.lock ./
RUN yarn install --frozen-lockfile
COPY web/ ./
RUN yarn build:server

# Build stage - Go binary
FROM golang:1.25-alpine AS go-builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=web-builder /app/web/dist ./web/dist

ARG VERSION=dev
ARG BUILD_DATE
RUN CGO_ENABLED=0 go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.buildDate=${BUILD_DATE}" \
    -o quantlete ./cmd/quantlete

# Runtime stage
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -u 1000 quantlete

COPY --from=go-builder /app/quantlete /usr/local/bin/quantlete

# Create data directory
RUN mkdir -p /data && chown quantlete:quantlete /data

USER quantlete

WORKDIR /data
VOLUME /data
ENV QUANTLETE_STORAGE_DATA_DIR=/data

EXPOSE 8081

ENTRYPOINT ["quantlete"]
CMD ["serve", "--port", "8081"]
