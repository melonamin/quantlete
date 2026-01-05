# Quantlete

**Self-hosted analytics dashboard for Strava athletes.**

Quantlete gives you comprehensive statistics, visualizations, and insights from your Strava data—all running locally or in your browser. No cloud subscription required.

**[Live Demo](https://demo.quantlete.fit)** · **[Documentation](https://docs.quantlete.fit)**

---

## Features

- **Dashboard** — Customizable widgets showing your key metrics
- **Activity Maps** — Route visualization with interactive heatmaps
- **Charts & Analytics** — Distance, elevation, pace, power, heart rate trends
- **Segment Analysis** — Track your segment efforts and PRs
- **Gear Tracking** — Monitor usage and schedule maintenance
- **Eddington Numbers** — Calculate your cycling/running Eddington number
- **Best Efforts** — Personal records across standard distances
- **Wrapped** — Year in review summaries
- **Training Load** — CTL/ATL/TSB fitness and fatigue tracking

## Quick Start

### Browser-Only (No Server)

Run entirely in your browser using WebAssembly. Your data stays on your device.

1. Visit **[quantlete.fit](https://quantlete.fit)**
2. Connect your Strava account
3. Import your activities

### Self-Hosted

```bash
docker run -d \
  -p 8081:8081 \
  -v quantlete-data:/data \
  -e QUANTLETE_STRAVA_CLIENT_ID=your_client_id \
  -e QUANTLETE_STRAVA_CLIENT_SECRET=your_client_secret \
  ghcr.io/melonamin/quantlete:latest
```

Or with Docker Compose:

```yaml
services:
  quantlete:
    image: ghcr.io/melonamin/quantlete:latest
    restart: unless-stopped
    ports:
      - "8081:8081"
    volumes:
      - quantlete-data:/data
    env_file:
      - .env

volumes:
  quantlete-data:
```

Create a `.env` file with your Strava credentials:

```bash
QUANTLETE_STRAVA_CLIENT_ID=your_client_id
QUANTLETE_STRAVA_CLIENT_SECRET=your_client_secret
```

> **Note:** Never commit `.env` files or credentials to version control.

See the **[deployment guide](https://docs.quantlete.fit/deployment)** for binary downloads and building from source.

## Documentation

Full documentation is available at **[docs.quantlete.fit](https://docs.quantlete.fit)**:

- [Getting Started](https://docs.quantlete.fit/getting-started)
- [Deployment Options](https://docs.quantlete.fit/deployment)
- [Configuration](https://docs.quantlete.fit/configuration)
- [Strava API Setup](https://docs.quantlete.fit/strava-setup)
- [Architecture](https://docs.quantlete.fit/architecture)

## Tech Stack

**Backend:** Go, Chi, SQLite, Cobra
**Frontend:** React, TypeScript, Tailwind CSS, shadcn/ui, ECharts, Leaflet
**WASM:** Go → WebAssembly, sql.js, OPFS

## Contributing

Contributions are welcome! Please see the [contributing guide](https://docs.quantlete.fit/contributing) for details.

## License

Quantlete is released under the [O'Saasy License](LICENSE.md).
