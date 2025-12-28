# Stata - Project Guidelines

This is the Stata project - a self-hosted analytics dashboard for Strava activities.

## Project Structure

- `cmd/stata/` - Go CLI entry points (main, serve, import commands)
- `internal/` - Go internal packages (api, storage, strava client, importer)
- `web/` - React frontend (Vite + TypeScript)
- `schema/migrations/` - SQLite database migrations
- `docs/` - Documentation (SPECIFICATION.md, ARCHITECTURE.md, TODO.md)

## Build Commands

Use `just` for all tasks:

```bash
just dev          # Run dev servers (Go + React)
just build        # Build everything
just test         # Run all tests
just lint         # Run all linters
just fmt          # Format all code
```

### Go Commands

```bash
just dev-api      # Run Go server in dev mode
just build-go     # Build Go binary
just test-go      # Run Go tests
just lint-go      # Lint Go code
```

### React Commands

```bash
just dev-web      # Run Vite dev server
just build-web    # Build React app
just test-web     # Run React unit tests
just lint-web     # Lint React code
cd web && yarn typecheck  # TypeScript check
```

## Code Style

### Go
- Use `slog` for structured logging
- Follow standard Go project layout
- Pure Go (no CGO required)
- Run `golangci-lint` before committing

### TypeScript/React
- Strict TypeScript mode enabled
- Use path alias `@/` for imports from `src/`
- Use Tailwind CSS with shadcn/ui components
- Use Zustand for state management
- Use TanStack Query for data fetching

## Database

- SQLite for storage and analytics
- Migrations in `schema/migrations/`
- Pure Go driver: `modernc.org/sqlite`

## Testing

- Go: `go test` with testify
- React: Jest/Vitest for unit tests, Playwright for E2E
- Always run tests before pushing

## Key Dependencies

### Go
- chi (HTTP router)
- cobra (CLI)
- viper (config)
- modernc.org/sqlite (database)

### React
- React 18+ with TypeScript
- Tailwind CSS v4
- shadcn/ui components
- ECharts for charts
- react-leaflet for maps
- TanStack (Query, Table, Virtual, Router)

## Do Not Modify

These files are manually configured and should not be changed:

- `.golangci.yml` - Linting rules are intentionally configured

## Important Notes

- Single binary distribution goal - React is embedded via go:embed
- Browser-only WASM mode is planned (sql.js + OPFS)
- Strava OAuth required for data import
- Rate limit awareness for Strava API (15-min and daily limits)
