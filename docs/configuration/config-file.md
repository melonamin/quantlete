# Config File

For complex configurations, use a YAML config file instead of environment variables.

## File Locations

Quantlete searches for `quantlete.yaml` in these locations (in order):

1. `./quantlete.yaml` - Current directory
2. `./config/quantlete.yaml` - Config subdirectory
3. `$HOME/.config/quantlete/quantlete.yaml` - User config directory
4. `/etc/quantlete/quantlete.yaml` - System-wide config

The first file found is used.

## Complete Example

```yaml
# Server configuration
server:
  port: 8081
  host: ""                    # Empty = all interfaces, or "127.0.0.1" for localhost only
  dev_mode: false
  read_timeout: 15s
  write_timeout: 15s
  idle_timeout: 60s

# Strava API configuration
strava:
  client_id: "12345"
  client_secret: "your_client_secret_here"
  redirect_uri: "http://localhost:8081/api/v1/auth/strava/callback"
  webhook_verify_token: ""    # Optional, min 16 chars if set
  webhook_subscription_id: 0  # Filled automatically after registration

# Storage configuration
storage:
  data_dir: "./data"
  db_file: "quantlete.db"

# Logging configuration
log:
  level: "info"               # debug, info, warn, error
  format: "text"              # text, json
```

## Minimal Example

Only Strava credentials are required:

```yaml
strava:
  client_id: "12345"
  client_secret: "your_client_secret_here"
```

## Production Example

```yaml
server:
  port: 8081
  host: "127.0.0.1"           # Localhost only (reverse proxy handles external)

strava:
  client_id: "12345"
  client_secret: "your_client_secret_here"
  redirect_uri: "https://quantlete.example.com/api/v1/auth/strava/callback"
  webhook_verify_token: "a_secure_random_token_here"

storage:
  data_dir: "/var/lib/quantlete"
  db_file: "quantlete.db"

log:
  level: "info"
  format: "json"              # JSON for log aggregation
```

## Combining with Environment Variables

Environment variables override config file values. This is useful for secrets:

```yaml
# quantlete.yaml - Safe to commit
server:
  port: 8081

storage:
  data_dir: "/var/lib/quantlete"

log:
  level: "info"
  format: "json"

# Strava credentials set via environment
# strava:
#   client_id: set via QUANTLETE_STRAVA_CLIENT_ID
#   client_secret: set via QUANTLETE_STRAVA_CLIENT_SECRET
```

Then set secrets in your environment or `.env` file:

```bash
export QUANTLETE_STRAVA_CLIENT_ID=12345
export QUANTLETE_STRAVA_CLIENT_SECRET=your_secret
```

## Docker with Config File

Mount your config file into the container:

```bash
docker run -d \
  -v /path/to/quantlete.yaml:/app/quantlete.yaml:ro \
  -v quantlete-data:/data \
  -p 8081:8081 \
  ghcr.io/melonamin/quantlete:latest
```
