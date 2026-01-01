# Configuration

Quantlete can be configured via environment variables, a config file, or command-line flags.

## Configuration Precedence

Configuration is loaded in this order (later sources override earlier):

1. Default values
2. Config file (YAML)
3. Environment variables
4. Command-line flags

## Quick Reference

| Setting | Environment Variable | Default |
|---------|---------------------|---------|
| Server port | `QUANTLETE_SERVER_PORT` | 8081 |
| Strava Client ID | `QUANTLETE_STRAVA_CLIENT_ID` | (required) |
| Strava Client Secret | `QUANTLETE_STRAVA_CLIENT_SECRET` | (required) |
| Data directory | `QUANTLETE_STORAGE_DATA_DIR` | ./data |
| Log level | `QUANTLETE_LOG_LEVEL` | info |

See [Environment Variables](environment.md) for the complete list.

## Minimal Configuration

The only required settings are Strava API credentials:

```bash
export QUANTLETE_STRAVA_CLIENT_ID=12345
export QUANTLETE_STRAVA_CLIENT_SECRET=your_secret_here
quantlete serve
```

## Using a Config File

For complex setups, use a YAML config file. See [Config File](config-file.md) for details.

```yaml
server:
  port: 8081

strava:
  client_id: "12345"
  client_secret: "your_secret_here"

storage:
  data_dir: "/var/lib/quantlete"
```

## Configuration Sections

- **[Environment Variables](environment.md)** - All available environment variables
- **[Config File](config-file.md)** - YAML configuration format and locations
