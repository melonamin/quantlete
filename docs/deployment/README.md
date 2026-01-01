# Deployment

Quantlete can be deployed in several ways depending on your infrastructure.

## Deployment Options

| Method | Best For | Complexity |
|--------|----------|------------|
| [Docker](docker.md) | Most users | Low |
| [Binary](binary.md) | Servers without Docker | Medium |
| From Source | Developers | High |

## Requirements

All deployment methods need:

- Strava API credentials ([setup guide](../getting-started/strava-setup.md))
- Persistent storage for the database
- Network access to Strava API

## Quick Comparison

### Docker

```bash
docker run -d -p 8081:8081 \
  -v quantlete-data:/data \
  -e QUANTLETE_STRAVA_CLIENT_ID=xxx \
  -e QUANTLETE_STRAVA_CLIENT_SECRET=xxx \
  ghcr.io/melonamin/quantlete:latest
```

**Pros:** Easy setup, isolated environment, easy updates
**Cons:** Requires Docker

### Binary

```bash
./quantlete serve
```

**Pros:** No dependencies, single file
**Cons:** Manual service setup

## Network Access

Quantlete needs outbound HTTPS access to:

- `www.strava.com` - OAuth and API
- `api.open-meteo.com` - Weather data (optional)

For webhooks (real-time sync), your server must be accessible from the internet.

## Storage

The SQLite database is stored in the data directory (default: `./data`). Ensure:

- Sufficient disk space (grows with activity history)
- Regular backups of the database file
- Persistent storage in containerized deployments

## Security Considerations

- Run behind a reverse proxy with HTTPS in production
- Keep Strava credentials secure (use environment variables, not files)
- The dashboard is currently single-user (no authentication beyond Strava OAuth)
- Consider network-level access controls if exposed to the internet
