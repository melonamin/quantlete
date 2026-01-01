# serve

Start the Quantlete web server.

## Usage

```bash
quantlete serve [OPTIONS]
```

## Options

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--port` | `-p` | Port to listen on | 8081 |
| `--dev` | | Run in development mode | false |

## Description

The `serve` command starts the HTTP server that hosts the Quantlete dashboard and API. It:

- Serves the React frontend
- Exposes the REST API at `/api/v1/`
- Handles Strava OAuth callbacks
- Processes webhook notifications
- Runs background sync and maintenance tasks

## Examples

### Basic Usage

```bash
quantlete serve
```

Server starts at http://localhost:8081

### Custom Port

```bash
quantlete serve --port 3000
```

Server starts at http://localhost:3000

### Development Mode

```bash
quantlete serve --dev
```

In development mode, the server expects the React dev server to be running separately on port 5173 and proxies frontend requests to it. This enables hot reloading during development.

## Environment Variables

The server respects these environment variables (see [Configuration](../configuration/environment.md)):

- `QUANTLETE_SERVER_PORT` - Default port (overridden by `--port`)
- `QUANTLETE_SERVER_HOST` - Bind address
- `QUANTLETE_SERVER_DEV_MODE` - Default dev mode (overridden by `--dev`)
- `QUANTLETE_STRAVA_CLIENT_ID` - Strava OAuth client ID
- `QUANTLETE_STRAVA_CLIENT_SECRET` - Strava OAuth client secret

## Shutdown

The server handles graceful shutdown on:
- `SIGINT` (Ctrl+C)
- `SIGTERM`

It stops accepting new connections, finishes active requests, and closes the database cleanly.
