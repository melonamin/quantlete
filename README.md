# Quantlete - Statistics for Strava

A self-hosted analytics dashboard for Strava athletes. Get comprehensive statistics, visualizations, and insights into your training data.

## Features

- Import activities from Strava API
- Interactive dashboard with customizable widgets
- Activity maps and heatmaps
- Detailed charts (ECharts)
- Segment analysis
- Gear tracking and maintenance
- Eddington numbers
- Best efforts and personal records
- Year in review (Strava Rewind style)
- Browser-only mode (WASM) - no server required

## Quick Start

### Using Docker

```bash
docker run -d \
  -p 8081:8081 \
  -v quantlete-data:/data \
  -e STRAVA_CLIENT_ID=your_client_id \
  -e STRAVA_CLIENT_SECRET=your_client_secret \
  ghcr.io/user/quantlete:latest
```

### Using Binary

Download the latest release for your platform from the [Releases](https://github.com/user/quantlete/releases) page.

```bash
# Start the server
./quantlete serve --port 8081
```

### From Source

Requirements:
- Go 1.23+
- Node.js 22+
- just (task runner)

```bash
# Clone the repository
git clone https://github.com/user/quantlete.git
cd quantlete

# Install dependencies
just setup

# Run development servers
just dev

# Build for production
just build
```

## Configuration

Create a `.env` file or set environment variables:

```bash
# Required: Strava API credentials
STRAVA_CLIENT_ID=your_client_id
STRAVA_CLIENT_SECRET=your_client_secret

# Optional
QUANTLETE_PORT=8081
QUANTLETE_DATA_DIR=/path/to/data
```

### Getting Strava API Credentials

1. Go to [Strava API Settings](https://www.strava.com/settings/api)
2. Create a new application
3. Set the callback URL to `http://localhost:8081/api/v1/auth/strava/callback`
4. Copy the Client ID and Client Secret

## Development

```bash
# Run development servers
just dev

# Run tests
just test

# Lint code
just lint

# Format code
just fmt
```

## Architecture

See [ARCHITECTURE.md](docs/ARCHITECTURE.md) for technical details.

## License

MIT
