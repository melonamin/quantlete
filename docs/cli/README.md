# CLI Reference

Quantlete provides a command-line interface for managing your Strava data.

## Commands

| Command | Description |
|---------|-------------|
| [serve](serve.md) | Start the web server |
| [import](import.md) | Import activities from Strava |
| [demo](demo.md) | Generate demo data |
| version | Display version information |

## Global Behavior

All commands:
- Load configuration from environment variables and config files
- Use the configured data directory for database storage
- Respect log level settings

## Quick Examples

```bash
# Start the server
quantlete serve

# Start on a different port
quantlete serve --port 3000

# Import all activities
quantlete import

# Full re-sync
quantlete import --full

# Generate demo data
quantlete demo

# Show version
quantlete version
```

## Getting Help

Use `--help` with any command:

```bash
quantlete --help
quantlete serve --help
quantlete import --help
```
