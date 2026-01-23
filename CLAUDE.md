# Quantlete - Project Guidelines

This is the Quantlete project - a self-hosted analytics dashboard for Strava activities.

## Architecture Overview

Quantlete operates in two modes sharing the same core Go logic:
1. **Server Mode:** Traditional Go HTTP server (REST API) + React SPA. Uses `modernc.org/sqlite` on disk.
2. **WASM Mode:** Browser-only. Go logic compiled to WebAssembly (`cmd/wasm`), running in a web worker. Uses `sql.js` (SQLite) + OPFS for storage.

Core business logic resides in `internal/services` and is decoupled from the transport layer (HTTP or WASM), ensuring 100% parity between modes.

## Project Structure

- `cmd/quantlete/` - Go CLI entry points (main, serve, import commands)
- `cmd/wasm/` - WASM entry points and bridge code
- `internal/` - Go internal packages
    - `analysis/` - Activity analysis functions (splits, HR zones, pace distribution)
    - `api/` - HTTP handlers (Server mode)
    - `services/` - Core business logic (Shared)
    - `storage/` - Database repositories (Shared)
    - `strava/` - Strava API client
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
just generate     # Regenerate SQL/types/adapters after service or schema changes
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

> **Build order matters:** `just build` runs generators, builds the React bundle, then builds the Go binary so `go:embed` picks up fresh assets. If you build targets manually, run `just build-web` before `just build-go`.

## Code Style

### Go
- Use `slog` for structured logging
- Follow standard Go project layout
- Pure Go (no CGO required) to ensure WASM compatibility
- Run `golangci-lint` before committing

### Go Error Handling

Use the established error patterns in `internal/services/errors.go`:

**Sentinel Errors:**
```go
// Use predefined errors for consistent categorization
services.ErrNotFound      // 404
services.ErrUnauthorized  // 401
services.ErrForbidden     // 403
services.ErrBadRequest    // 400
services.ErrConflict      // 409
services.ErrInternal      // 500
```

**Wrapping with Context:**
```go
// Good: wrap with context using helper functions
return services.NotFound("athlete")
return services.BadRequest("invalid date range")

// Good: wrap underlying errors
return services.Wrapf(err, "failed to fetch activities")

// Bad: generic errors without context
return fmt.Errorf("not found")
return errors.New("something went wrong")
```

**Client Message Masking:**
```go
// Use ClientMessage() to return safe error messages to API clients
// Internal details are masked; only ErrBadRequest messages are exposed
msg := services.ClientMessage(err)
```

**Repository Pattern:**
```go
// Repositories return (nil, nil) for not-found, not (nil, error)
func (r *GearRepository) GetByID(ctx context.Context, id string) (*Gear, error) {
    // ...
    if err == sql.ErrNoRows {
        return nil, nil  // Not found is not an error
    }
    return nil, fmt.Errorf("scanning gear: %w", err)
}
```

### Go Logging

Always use structured logging with `slog`:

```go
// Good: structured with key-value pairs
slog.Info("starting server", "addr", addr, "mode", mode)
slog.Error("failed to save", "error", err, "athlete_id", id)
slog.Warn("rate limit approaching", "remaining", remaining)

// Bad: embedding values in message string
slog.Info(fmt.Sprintf("starting server on %s", addr))
slog.Error("failed to save: " + err.Error())
```

**Anti-patterns to avoid:**
- Never embed error text in log message string
- Never use `fmt.Sprintf` for log messages
- Never log sensitive data (tokens, passwords)

### API Response Envelope

- Every handler should respond via `shared.Response` helpers (`shared.SuccessResponse`, `shared.ErrorMessage`, `shared.WriteJSONResponse`) so the frontend `ApiClient` can unwrap `{ ok, data, message }`.
- Stick to that envelope even for WASM-path functions; the TS bridge (`goStorage`) expects to parse the same structure before surfacing `data` to React Query.

### TypeScript/React
- Strict TypeScript mode enabled
- Use path alias `@/` for imports from `src/`
- Use Tailwind CSS v4 with shadcn/ui components
- Use Zustand for state management
- Use TanStack Query for data fetching
- All data access flows through `DataProviderWrapper`; never hit `fetch` directly in components or hooks.
- React Query hooks must gate on `useDataProviderStatus()` (initialized/no error) and reuse the `['data', ...]` key prefix so the event bus can invalidate them.
- Mutations have to invalidate related caches (see `useUpdateDashboardConfig` for the canonical pattern).
- Global side effects (theme toggle, sync protection, data events) are mounted once inside `SyncEffects` in `web/src/main.tsx`; add new global hooks there.

#### Import Conventions

**ALWAYS use `@/` path aliases, never relative imports:**
```typescript
// Good
import { Button } from '@/components/ui/button'
import { useActivities } from '@/lib/data/hooks'
import { formatDistance } from '@/lib/format'

// Bad - never use relative imports
import { Button } from '../../../components/ui/button'
import { useActivities } from '../../lib/data/hooks'
```

**Import grouping order:**
1. React and React-related (`react`, `react-dom`)
2. Third-party libraries (`@tanstack/*`, `zustand`, etc.)
3. Local modules (`@/lib/*`, `@/hooks/*`)
4. Components (`@/components/*`)
5. Types (if separate)

#### Component Organization

- Organize by feature in `components/` (e.g., `activities/`, `charts/`, `dashboard/`)
- Use barrel exports via `index.ts` for each feature folder
- Named exports only (no default exports)
- Define prop interfaces directly above components:

```typescript
interface ActivityRowProps {
  activity: Activity
  isSelected?: boolean
  onSelect?: (id: number) => void
}

export function ActivityRow({ activity, isSelected, onSelect }: ActivityRowProps) {
  // ...
}
```

#### State Management (Zustand)

```typescript
// Always define typed interface for store state
interface ActivityFiltersState {
  filters: ActivityFilters
  setFilters: (filters: Partial<ActivityFilters>) => void
  resetFilters: () => void
}

// Create store with typed interface
export const useActivityFiltersStore = create<ActivityFiltersState>((set) => ({
  filters: defaultFilters,
  setFilters: (filters) => set((s) => ({ filters: { ...s.filters, ...filters } })),
  resetFilters: () => set({ filters: defaultFilters }),
}))
```

**Rules:**
- Use helper functions for complex state transformations (keep immutable)
- Only use `persist` middleware for user preferences (settings, theme)
- Use selectors to avoid unnecessary re-renders: `useStore((s) => s.specificField)`

#### Data Fetching (TanStack Query)

```typescript
export function useActivities(filters: ActivityFilters) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    // Query key pattern: ['data', feature, ...specifics]
    queryKey: ['data', 'activities', filters],
    queryFn: () => provider.getActivities(filters),
    // Always check provider state before enabling
    enabled: initialized && !error && !!provider,
    staleTime: 1000 * 60, // 1 minute for most data
  })
}
```

**Rules:**
- Query keys follow pattern: `['data', feature, ...specifics]`
- Always check `initialized && !error && !!provider` before enabling queries
- Mutations must invalidate related queries in `onSuccess`
- Use shorter `staleTime` for auth (30s), longer for activity data (1-5min)

#### Hook Patterns

```typescript
/**
 * Hook to track sync progress and prevent accidental page close.
 * Shows browser's native confirmation dialog during active sync.
 */
export function useSyncProtection() {
  const { data: progress } = useImportProgress()
  const isRunning = progress?.status === 'running'

  useEffect(() => {
    // Return early if conditions aren't met
    if (!isRunning) return

    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      e.preventDefault()
      e.returnValue = 'Sync in progress. Leave anyway?'
      return e.returnValue
    }

    window.addEventListener('beforeunload', handleBeforeUnload)
    // Always return cleanup function
    return () => window.removeEventListener('beforeunload', handleBeforeUnload)
  }, [isRunning])
}
```

**Rules:**
- Include JSDoc comment explaining the hook's purpose
- Return early from effects if conditions aren't met
- Always return cleanup functions from useEffect
- Each hook should focus on a single concern

## Database

- SQLite for storage and analytics
- Migrations in `schema/migrations/`
- Pure Go driver: `modernc.org/sqlite`
- **WASM Support:** Uses `sql.js` (SQLite compiled to WASM) in the browser, sharing the same schema.

### Pagination & Sorting

- Pagination defaults live in both `internal/pagination/pagination.go` and `web/src/lib/constants.ts`; update them together.
- Repositories must call `Normalize()` on pagination structs before querying so user input cannot bypass bounds.
- Sortable columns must be whitelisted (e.g., `allowedActivityOrderColumns` in `internal/storage/activities.go`). Extend those maps whenever you expose a new `order_by` field.

### SQL Queries

**NEVER write raw SQL in Go or TypeScript code.** All queries must go through the codegen system:

1. Define queries in `schema/queries/*.sql` using sqlc-style annotations
2. Run `just generate-sql` to generate `queries.gen.go` and `queries.gen.ts`
3. Use the generated `Queries` struct methods in storage layer code

Raw SQL requires explicit confirmation from Sasha. This ensures:
- Parity between Go backend and TypeScript WASM provider
- Single source of truth for all queries
- Type-safe query methods with proper row types

### Service Layer Rules

- Input structs need `adapter:"..."` tags so generated HTTP/WASM handlers map fields from context/query/path/body correctly.
- Every exported service method requires both `//adapter:http` and `//adapter:wasm` annotations; run `just check-adapter-parity` after edits.
- Authorization stays in the service: fetch the entity, verify `AthleteID` (or other owner field), and return `services.ErrForbidden` if mismatched.
- Enforce payload caps like `maxStreamDataSize` when accepting bulk arrays to avoid blowing up WASM/browser memory.

## Code Generation

The project uses extensive code generation for type safety and platform parity (Go server + WASM browser).

### Generation Order (dependencies matter)

Run `just generate` or execute in order:

```bash
just generate-sql              # 1. SQL queries → Go methods (queries.gen.go)
just generate-schema           # 2. Browser WASM schema
just generate-ts-types         # 3. Go structs → TypeScript types
just generate-adapters         # 4. Service annotations → HTTP/WASM adapters
just generate-wasm-registration # 5. WASM exports registration
just generate-go-storage       # 6. goStorage TypeScript interface
```

**Each step depends on previous outputs.** Running out of order will cause build failures.

### Service Adapter Annotations

Services require dual annotations for HTTP and WASM parity:

```go
//adapter:wasm getActivities category=Activities
//adapter:http GET /api/v1/activities
func (s *ActivityService) List(ctx context.Context, in ListInput) (*ListOutput, error) {
    // ...
}
```

**Input struct tags** specify parameter sources:
```go
type ListInput struct {
    AthleteID  int64    `json:"athlete_id" adapter:"context"`        // From auth context
    SportTypes []string `json:"sport_types" adapter:"query,name=sport_type,split=,"` // Query param
    Page       int      `json:"page" adapter:"query"`                // Query param
    ActivityID int64    `json:"activity_id" adapter:"path"`          // URL path param
}
```

**After modifying services:**
1. Add both `//adapter:wasm` and `//adapter:http` annotations
2. Run `just check-adapter-parity` to verify consistency
3. Run `just generate` to regenerate adapters

### Handler Patterns

Handlers use constructor injection with dependencies:

```go
type AthleteHandler struct {
    metrics *storage.AthleteMetricsRepository
    strava  *strava.Client
}

func NewAthleteHandler(metrics *storage.AthleteMetricsRepository, strava *strava.Client) *AthleteHandler {
    return &AthleteHandler{metrics: metrics, strava: strava}
}

func (h *AthleteHandler) GetFTP(w http.ResponseWriter, r *http.Request) {
    athlete := h.strava.GetAthlete()
    if athlete == nil {
        shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
        return
    }

    ftp, err := h.metrics.GetCurrentFTP(r.Context(), athlete.ID)
    if err != nil {
        handleServiceError(w, err)
        return
    }

    shared.WriteSuccess(w, ftp)
}
```

**Response helpers** (from `internal/shared/response.go`):
- `shared.WriteSuccess(w, data)` - 200 with data
- `shared.WriteJSONResponse(w, status, response)` - Custom status
- `shared.ErrorMessage(msg)` - Error response struct

## Security

- **CSRF:** State-changing requests (POST/PUT/DELETE) are protected by verifying the `Origin` header matches the `Host` (or an allowlist in dev).
- **CSP:** Strict Content Security Policy is enforced via `securityHeaders` middleware.
- **Secrets:** Never commit credentials. Use `.env` files or environment variables.
- **Rate Limiting:** Strava API limits (15-min and daily) are tracked. Logic handles `429` responses gracefully.

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
- Browser-only WASM mode is implemented using a shared Go storage layer.
- Strava OAuth required for data import
- Detect runtime mode (server / wasm / demo) exclusively via `web/src/lib/mode.ts`; never inline environment checks in components.
- After adding `//wasm:export` functions, rerun `just generate-wasm-registration` and `just generate-go-storage` so TS bindings stay in sync.
- Importer changes must update `ImportState`, per-phase counters, SSE payloads, and `useDataEvents` invalidation map (`activities`, `streams`, `segments`, `gear`, `photos`, `all`).
- New Strava calls must update the shared rate limiter (`internal/strava/ratelimit.go`) so progress ETA/wait messaging remains accurate in both server and WASM builds.

## Environment Variables

All environment variables use `QUANTLETE_` prefix with nested underscores:

| Variable | Description | Default |
|----------|-------------|---------|
| `QUANTLETE_SERVER_PORT` | Server port | `8081` |
| `QUANTLETE_SERVER_DEV_MODE` | Enable dev mode | `false` |
| `QUANTLETE_STRAVA_CLIENT_ID` | Strava OAuth client ID | - |
| `QUANTLETE_STRAVA_CLIENT_SECRET` | Strava OAuth client secret | - |
| `QUANTLETE_STORAGE_DATA_DIR` | Data directory | `./data` |
| `QUANTLETE_STORAGE_DB_FILE` | Database filename | `quantlete.db` |
| `QUANTLETE_LOG_LEVEL` | Log level (debug, info, warn, error) | `info` |

Config file search paths (YAML format):
- `.` (current directory)
- `./config`
- `$HOME/.config/quantlete`
- `/etc/quantlete`