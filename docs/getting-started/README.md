# Installation

Choose your preferred installation method.

## Docker (Recommended)

The easiest way to run Quantlete:

```bash
docker run -d \
  --name quantlete \
  -p 8081:8081 \
  -v quantlete-data:/data \
  -e QUANTLETE_STRAVA_CLIENT_ID=your_client_id \
  -e QUANTLETE_STRAVA_CLIENT_SECRET=your_client_secret \
  ghcr.io/melonamin/quantlete:latest
```

Then open http://localhost:8081 in your browser.

See [Docker Deployment](../deployment/docker.md) for advanced configuration.

## Pre-built Binary

Download the latest release for your platform:

```bash
# Linux (amd64)
curl -LO https://github.com/melonamin/quantlete/releases/latest/download/quantlete-linux-amd64
chmod +x quantlete-linux-amd64
mv quantlete-linux-amd64 quantlete

# macOS (Apple Silicon)
curl -LO https://github.com/melonamin/quantlete/releases/latest/download/quantlete-darwin-arm64
chmod +x quantlete-darwin-arm64
mv quantlete-darwin-arm64 quantlete

# macOS (Intel)
curl -LO https://github.com/melonamin/quantlete/releases/latest/download/quantlete-darwin-amd64
chmod +x quantlete-darwin-amd64
mv quantlete-darwin-amd64 quantlete
```

Run the server:

```bash
export QUANTLETE_STRAVA_CLIENT_ID=your_client_id
export QUANTLETE_STRAVA_CLIENT_SECRET=your_client_secret
./quantlete serve
```

See [Binary Deployment](../deployment/binary.md) for running as a service.

## From Source

Requirements:
- Go 1.23+
- Node.js 22+
- just (task runner)

```bash
# Clone the repository
git clone https://github.com/melonamin/quantlete.git
cd quantlete

# Install dependencies
just setup

# Build
just build

# Run
./bin/quantlete serve
```

## Next Steps

1. [Set up Strava API credentials](strava-setup.md)
2. [Quick start guide](quick-start.md)
