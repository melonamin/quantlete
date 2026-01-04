# Environment Variables

All configuration options can be set via environment variables with the `QUANTLETE_` prefix.

## Server Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `QUANTLETE_SERVER_PORT` | HTTP server port | 8081 |
| `QUANTLETE_SERVER_HOST` | Bind address (empty = all interfaces) | "" |
| `QUANTLETE_SERVER_DEV_MODE` | Enable development mode | false |
| `QUANTLETE_SERVER_READ_TIMEOUT` | HTTP read timeout | 15s |
| `QUANTLETE_SERVER_WRITE_TIMEOUT` | HTTP write timeout | 15s |
| `QUANTLETE_SERVER_IDLE_TIMEOUT` | HTTP idle timeout | 60s |

## Strava Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `QUANTLETE_STRAVA_CLIENT_ID` | Strava OAuth Client ID | (required) |
| `QUANTLETE_STRAVA_CLIENT_SECRET` | Strava OAuth Client Secret | (required) |
| `QUANTLETE_STRAVA_REDIRECT_URI` | OAuth callback URL | http://localhost:8081/api/v1/auth/strava/callback |
| `QUANTLETE_STRAVA_WEBHOOK_VERIFY_TOKEN` | Webhook verification token (min 16 chars) | "" |
| `QUANTLETE_STRAVA_WEBHOOK_SUBSCRIPTION_ID` | Webhook subscription ID (set manually after creating the Strava webhook) | 0 |

## Storage Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `QUANTLETE_STORAGE_DATA_DIR` | Directory for database and files | ./data |
| `QUANTLETE_STORAGE_DB_FILE` | SQLite database filename | quantlete.db |

The database will be created at `{data_dir}/{db_file}`.

## Logging Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `QUANTLETE_LOG_LEVEL` | Log verbosity (debug, info, warn, error) | info |
| `QUANTLETE_LOG_FORMAT` | Log format (text, json) | text |

## Examples

### Basic Setup

```bash
export QUANTLETE_STRAVA_CLIENT_ID=12345
export QUANTLETE_STRAVA_CLIENT_SECRET=abcdef123456
quantlete serve
```

### Custom Port and Data Directory

```bash
export QUANTLETE_SERVER_PORT=3000
export QUANTLETE_STORAGE_DATA_DIR=/var/lib/quantlete
export QUANTLETE_STRAVA_CLIENT_ID=12345
export QUANTLETE_STRAVA_CLIENT_SECRET=abcdef123456
quantlete serve
```

### Production with Webhooks

```bash
export QUANTLETE_SERVER_PORT=8081
export QUANTLETE_STORAGE_DATA_DIR=/var/lib/quantlete
export QUANTLETE_STRAVA_CLIENT_ID=12345
export QUANTLETE_STRAVA_CLIENT_SECRET=abcdef123456
export QUANTLETE_STRAVA_REDIRECT_URI=https://quantlete.example.com/api/v1/auth/strava/callback
export QUANTLETE_STRAVA_WEBHOOK_VERIFY_TOKEN=your_secure_random_token
export QUANTLETE_STRAVA_WEBHOOK_SUBSCRIPTION_ID=123456
export QUANTLETE_LOG_LEVEL=info
export QUANTLETE_LOG_FORMAT=json
quantlete serve
```

### Debug Logging

```bash
export QUANTLETE_LOG_LEVEL=debug
quantlete serve
```

## Using .env Files

You can store environment variables in a `.env` file:

```bash
# .env
QUANTLETE_STRAVA_CLIENT_ID=12345
QUANTLETE_STRAVA_CLIENT_SECRET=abcdef123456
QUANTLETE_STORAGE_DATA_DIR=./data
```

Then use with Docker:

```bash
docker run --env-file .env ghcr.io/melonamin/quantlete:latest
```

Or source it in your shell:

```bash
set -a && source .env && set +a
quantlete serve
```
