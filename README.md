# Quantlete

**Self-hosted analytics dashboard for Strava athletes.**


Quantlete gives you comprehensive statistics, visualizations, and insights from your Strava data—all running locally or in your browser. No cloud subscription required.

**[Live Demo](https://demo.quantlete.fit)** · **[Documentation](ARCHITECTURE.md)**

<!-- TODO: Add screenshot
![Quantlete Dashboard](docs/screenshot.png)
-->

---

## Features

- **Dashboard** — Customizable widgets showing your key metrics
- **Activity Maps** — Route visualization with interactive heatmaps
- **Charts & Analytics** — Distance, elevation, pace, power, heart rate trends
- **Segment Analysis** — Track your segment efforts and PRs
- **Gear Tracking** — Monitor usage and schedule maintenance
- **Eddington Numbers** — Calculate your cycling/running Eddington number
- **Best Efforts** — Personal records across standard distances
- **Year in Review** — Strava Rewind-style annual summaries
- **Training Load** — CTL/ATL/TSB fitness and fatigue tracking

## Deployment Options

### Browser-Only Mode (No Server)

Run entirely in your browser using WebAssembly. Your data stays on your device using browser storage (OPFS).

1. Visit [quantlete.fit](https://quantlete.fit) (or self-host the static files)
2. Connect your Strava account
3. Import your activities

No server required—everything runs client-side.

### Self-Hosted Server

Run the Go binary for a traditional server setup with SQLite storage.

#### Docker

```bash
docker run -d \
  -p 8081:8081 \
  -v quantlete-data:/data \
  -e STRAVA_CLIENT_ID=your_client_id \
  -e STRAVA_CLIENT_SECRET=your_client_secret \
  ghcr.io/melonamin/quantlete:latest
```

#### Binary

Download from [Releases](https://github.com/melonamin/quantlete/releases), then:

```bash
./quantlete serve --port 8081
```

#### From Source

Requirements: Go 1.23+, Node.js 22+, [just](https://github.com/casey/just)

```bash
git clone https://github.com/melonamin/quantlete.git
cd quantlete
just setup
just build
./bin/quantlete serve
```

## Configuration

### Environment Variables

```bash
# Required for self-hosted mode
STRAVA_CLIENT_ID=your_client_id
STRAVA_CLIENT_SECRET=your_client_secret

# Optional
QUANTLETE_PORT=8081
QUANTLETE_DATA_DIR=/path/to/data
```

### Getting Strava API Credentials

1. Go to [Strava API Settings](https://www.strava.com/settings/api)
2. Create a new application
3. Set the callback URL:
   - Self-hosted: `http://localhost:8081/api/v1/auth/strava/callback`
   - Browser-only: Your domain's callback URL
4. Copy the Client ID and Client Secret

## Development

```bash
# Install dependencies
just setup

# Run dev servers (Go API + Vite)
just dev

# Run tests
just test

# Lint and format
just lint
just fmt

# Build for production
just build
```

See [ARCHITECTURE.md](ARCHITECTURE.md) for technical details on the codebase structure, WASM bridge, and code generation.

## Tech Stack

**Backend:** Go, Chi, SQLite (modernc.org/sqlite), Cobra

**Frontend:** React, TypeScript, Tailwind CSS, shadcn/ui, ECharts, Leaflet

**WASM:** Go compiled to WebAssembly, sql.js, OPFS

## License

