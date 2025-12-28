# Statistics for Strava - Implementation Plan

This document provides a phased implementation plan for the Statistics for Strava application. Each phase builds upon the previous one, delivering incremental value.

---

## Overview

| Phase | Focus | Deliverable | Status |
|-------|-------|-------------|--------|
| 0 | Project Setup | Repository, tooling, CI/CD | ✓ |
| 1 | Core Backend | Go server, Strava OAuth, basic import | ✓ |
| 2 | Database & Storage | DuckDB schema, activity storage | ✓ |
| 3 | Frontend Foundation | React shell, routing, shadcn setup | ✓ |
| 4 | Activities Feature | Activity list, filters, detail view | ✓ |
| 5 | Dashboard | Widget system, core widgets | ✓ |
| 6 | Charts & Visualizations | ECharts integration, all chart types | ✓ |
| 7 | Maps & Heatmap | Leaflet integration, route visualization | |
| 8 | Advanced Features | Segments, gear, maintenance, calendar | |
| 9 | Analytics | Eddington, best efforts, training load | |
| 10 | WASM Mode | Browser-only version with DuckDB-WASM | |
| 11 | Polish | PWA, i18n, settings, badges | |

---

## Phase 0: Project Setup ✓

**Goal:** Repository structure, tooling, and CI/CD pipeline ready.

### 0.1 Repository Initialization
- [x] Initialize Git repository
- [x] Create directory structure per ARCHITECTURE.md
- [x] Create `go.mod` with module path
- [x] Create `web/package.json` with dependencies
- [x] Create `justfile` with development commands
- [x] Create project `CLAUDE.md` with coding guidelines

### 0.2 Go Project Setup
- [x] Install Go dependencies:
  - [x] `github.com/go-chi/chi/v5`
  - [x] `github.com/spf13/cobra`
  - [x] `github.com/spf13/viper`
  - [x] `github.com/marcboeker/go-duckdb`
  - [x] `golang.org/x/oauth2`
  - [x] `github.com/robfig/cron/v3`
  - [x] `github.com/stretchr/testify`
- [x] Create basic `cmd/stata/main.go` entry point
- [x] Create `cmd/stata/root.go` with cobra root command
- [x] Create `cmd/stata/version.go` command
- [x] Verify `go build` works

### 0.3 React Project Setup
- [x] Initialize Vite React TypeScript project in `web/`
- [x] Configure TypeScript with strict mode
- [x] Install and configure Tailwind CSS
- [x] Install and configure shadcn/ui
- [x] Configure path aliases (`@/components`, etc.)
- [x] Install core dependencies:
  - [x] `zustand`
  - [x] `@tanstack/react-query`
  - [x] `@tanstack/react-router`
  - [x] `echarts` + `echarts-for-react`
  - [x] `react-leaflet` + `leaflet`
  - [x] `@tanstack/react-table`
  - [x] `@tanstack/react-virtual`
  - [x] `tailwind-merge` + `clsx`
  - [x] `date-fns`
- [x] Verify `yarn dev` works
- [x] Verify `yarn build` produces output

### 0.4 Linting & Formatting
- [x] Install and configure `golangci-lint` with config file
- [x] Configure ESLint for React/TypeScript
- [x] Configure Prettier
- [x] Add lint commands to justfile
- [x] Verify linting passes

### 0.5 Testing Setup
- [x] Configure Go test structure
- [ ] Configure Jest/Vitest for React unit tests
- [ ] Install Playwright for E2E tests
- [ ] Create initial placeholder tests
- [x] Add test commands to justfile

### 0.6 CI/CD Pipeline
- [x] Create `.github/workflows/ci.yml`:
  - [x] Go build and test
  - [x] Go lint
  - [x] React build and test
  - [x] React lint and typecheck
- [x] Create `.github/workflows/release.yml`:
  - [x] GoReleaser configuration
  - [x] Multi-platform builds (linux, darwin, windows × amd64, arm64)
  - [x] Docker image build and push
- [x] Create `.goreleaser.yaml`
- [x] Create `Dockerfile`

### 0.7 Documentation
- [x] Create README.md with:
  - [x] Project description
  - [x] Development setup instructions
  - [x] Build instructions
  - [x] Deployment instructions

---

## Phase 1: Core Backend ✓

**Goal:** Go server running with Strava OAuth working.

### 1.1 Configuration System
- [x] Create `internal/config/config.go` with config struct
- [x] Create `internal/config/loader.go` with Viper integration
- [x] Support configuration via:
  - [x] Environment variables
  - [x] Config file (YAML)
  - [x] Command-line flags
- [x] Configuration fields:
  - [x] Server port
  - [x] Data directory path
  - [x] Strava client ID/secret
  - [x] Strava redirect URI
  - [x] Log level

### 1.2 HTTP Server
- [x] Create `internal/api/router.go` with chi router
- [x] Create `internal/api/middleware.go`:
  - [x] Request logging middleware
  - [x] CORS middleware (for dev mode)
  - [x] Recovery middleware
- [x] Create `cmd/stata/serve.go` command
- [x] Implement graceful shutdown
- [ ] Serve static files from embedded React build
- [x] Verify server starts and serves placeholder page

### 1.3 Strava OAuth Flow
- [x] Create `internal/strava/oauth.go`:
  - [x] OAuth2 config setup
  - [x] Generate auth URL
  - [x] Exchange code for tokens
  - [x] Refresh expired tokens
- [x] Create `internal/api/handlers/auth.go`:
  - [x] `GET /api/v1/auth/strava` - redirect to Strava
  - [x] `GET /api/v1/auth/strava/callback` - handle callback
  - [x] `GET /api/v1/auth/status` - check auth status
  - [x] `POST /api/v1/auth/refresh` - force token refresh
- [x] Store tokens (initially in memory, later in DB)
- [x] Test OAuth flow end-to-end

### 1.4 Strava API Client
- [x] Create `internal/strava/client.go`:
  - [x] HTTP client with auth header injection
  - [x] Rate limit tracking from response headers
  - [ ] Automatic token refresh on 401
- [x] Create `internal/strava/types.go`:
  - [x] Activity struct matching Strava API
  - [x] Athlete struct
  - [x] Gear struct
  - [x] Stream types
- [x] Create `internal/strava/ratelimit.go`:
  - [x] Track 15-min and daily limits
  - [x] Expose rate limit status

### 1.5 Basic Activity Fetching
- [x] Create `internal/strava/activities.go`:
  - [x] `GetActivities(page, perPage)` - list activities
  - [x] `GetActivity(id)` - single activity detail
  - [x] `GetActivityStreams(id, types)` - stream data
- [x] Create `internal/api/handlers/activities.go`:
  - [x] `GET /api/v1/activities` - from local DB
  - [x] `GET /api/v1/activities/:id` - single activity
- [ ] Test fetching activities from Strava API

---

## Phase 2: Database & Storage ✓

**Goal:** DuckDB schema set up, activities persisted locally.

### 2.1 DuckDB Connection
- [x] Create `internal/storage/db.go`:
  - [x] Open/create DuckDB file
  - [x] Connection pool management
  - [x] Close on shutdown
- [x] Configure database path from config
- [ ] Test database creation

### 2.2 Schema Migrations
- [x] Create `internal/storage/migrations.go`:
  - [x] Migration tracking table
  - [x] Run pending migrations
  - [ ] Rollback support (optional)
- [x] Create `schema/migrations/001_initial.sql`:
  - [x] Athletes table
  - [x] Auth tokens table
- [x] Create `schema/migrations/002_activities.sql`:
  - [x] Activities table with all fields
  - [x] Indexes
- [x] Create `schema/migrations/003_streams.sql`:
  - [x] Activity streams table
- [x] Embed migrations using `go:embed`

### 2.3 Activity Repository
- [x] Create `internal/storage/activities.go`:
  - [x] `Insert(activity)` - upsert activity
  - [x] `GetByID(id)` - fetch single activity
  - [x] `List(filters, pagination)` - filtered list
  - [x] `GetTotals(filters)` - sum aggregates
- [ ] Create `internal/storage/queries/activities.sql`:
  - [ ] Parameterized queries
- [ ] Unit tests for repository

### 2.4 Importer Service
- [x] Create `internal/importer/importer.go`:
  - [x] Orchestrate full import
  - [x] Track import progress
  - [x] Handle rate limits gracefully
- [x] Create `internal/importer/activities.go`:
  - [x] Fetch activities in batches
  - [x] Transform Strava → domain model
  - [x] Persist to database
- [x] Create `cmd/stata/import.go` command:
  - [x] Trigger manual import
  - [x] Show progress
- [x] Test importing real activities

### 2.5 Athlete & Token Storage
- [x] Create `internal/storage/athletes.go`:
  - [x] Store athlete profile
  - [x] Store/retrieve OAuth tokens
- [x] Update OAuth flow to persist tokens
- [x] Retrieve tokens on server start

---

## Phase 3: Frontend Foundation ✓

**Goal:** React app shell with routing and basic layout.

### 3.1 App Shell
- [x] Create `web/src/main.tsx` with router setup
- [x] Create route definitions for all pages
- [x] Create `web/src/pages/` placeholder pages:
  - [x] Dashboard
  - [x] Activities
  - [x] Activity Detail
  - [x] Heatmap
  - [x] Calendar
  - [x] Segments
  - [x] Gear
  - [x] Eddington
  - [x] Best Efforts
  - [x] Settings

### 3.2 Layout Components
- [x] Create `web/src/components/layout/root-layout.tsx`:
  - [x] App title
  - [x] Navigation sidebar
  - [x] Content area
- [ ] Create breadcrumb component (deferred)
- [ ] Collapsible sidebar on mobile (deferred)

### 3.3 shadcn/ui Base Components
- [x] Initialize shadcn/ui with Tailwind v4
- [x] Add essential components via shadcn CLI:
  - [x] Button
  - [x] Card
  - [x] Table
  - [x] Badge
  - [x] Skeleton
  - [x] Separator
  - [ ] Dialog (add when needed)
  - [ ] Dropdown Menu (add when needed)
  - [ ] Input (add when needed)
  - [ ] Select (add when needed)

### 3.4 Data Layer Setup
- [x] Create `web/src/lib/api/types.ts` - TypeScript types
- [x] Create `web/src/lib/api/client.ts` - REST client
- [x] Create `web/src/lib/api/activities.ts` - Activity hooks
- [x] Create `web/src/lib/api/auth.ts` - Auth hooks

### 3.5 State Management
- [x] Create `web/src/stores/settings.ts`:
  - [x] Unit system (metric/imperial)
  - [x] Theme preference
- [x] Create `web/src/stores/activity-filters.ts`:
  - [x] Filter state
  - [ ] URL sync (deferred)

### 3.6 TanStack Query Setup
- [x] Configure QueryClient with defaults
- [x] Create query key factory (activityKeys, authKeys)
- [x] Create base hooks structure

### 3.7 Utility Functions
- [x] Create `web/src/lib/format.ts`:
  - [x] Distance conversion
  - [x] Speed conversion
  - [x] Elevation conversion
  - [x] Pace formatting
  - [x] Date/time formatting
  - [x] Duration formatting
- [x] Create `web/src/lib/utils.ts` - className helper (cn)
- [x] Create `web/src/lib/sport-types.ts`:
  - [x] Sport category mapping
  - [x] Sport colors
  - [x] Sport icons

### 3.8 Integration
- [ ] Embed React build in Go binary using `go:embed`
- [ ] Serve React from Go server
- [x] Configure Vite proxy for API in dev mode
- [x] Test full stack integration

---

## Phase 4: Activities Feature ✓

**Goal:** Complete activities list with filtering, sorting, and detail view.

### 4.1 Activities API Endpoints
- [x] Implement `GET /api/v1/activities`:
  - [x] Pagination
  - [x] All filter parameters
  - [x] Sorting
  - [x] Search
- [x] Implement `GET /api/v1/activities/:id`:
  - [x] Full activity detail
  - [ ] Include gear info
- [ ] Implement `GET /api/v1/activities/:id/streams`:
  - [ ] Return stream data
- [ ] Implement `GET /api/v1/activities/:id/photos`:
  - [ ] Return photo URLs

### 4.2 Activities List Page
- [x] Create `web/src/lib/api/activities.ts`:
  - [x] TanStack Query hook
  - [x] Filter state management
  - [x] Pagination state
- [ ] Create `web/src/components/activities/activity-filters.tsx`:
  - [ ] Sport type multi-select
  - [ ] Date range picker
  - [ ] Gear dropdown
  - [ ] Commute toggle
  - [ ] Search input
  - [ ] Clear filters button
- [x] Create `web/src/components/activities/activities-table.tsx`:
  - [x] Activity table with columns
  - [x] Loading skeleton
  - [x] Empty state
- [x] Create `web/src/components/activities/pagination.tsx`:
  - [x] Page navigation
  - [x] Results count
- [x] Create `web/src/pages/activities.tsx`:
  - [x] Table with real data
  - [x] Pagination
  - [ ] URL state sync (deferred)

### 4.3 Activity Detail Page
- [x] Create `web/src/pages/activity-detail.tsx`
- [x] Create `web/src/components/activities/activity-header.tsx`:
  - [x] Name, date, sport icon
  - [x] Link to Strava
  - [x] Tags (commute, indoor, private)
- [x] Create `web/src/components/activities/activity-stats.tsx`:
  - [x] Grid of key metrics
  - [x] Conditional display based on sport type
- [ ] Create `web/src/components/activities/activity-map.tsx`:
  - [ ] Route visualization (Phase 7)

### 4.4 Import Functionality
- [x] Create `internal/api/handlers/import.go`:
  - [x] Start import endpoint
  - [x] Progress endpoint
  - [x] Cancel endpoint
- [x] Create `web/src/lib/api/import.ts`:
  - [x] Import hooks with polling
- [x] Update Settings page:
  - [x] Strava connection status
  - [x] Import progress display
  - [x] Start/cancel import buttons

### 4.5 Activity Streams (Preparation)
- [ ] Create stream data hooks
- [ ] Create stream data types
- [ ] Prepare for chart integration (Phase 6)

---

## Phase 5: Dashboard ✓

**Goal:** Configurable widget-based dashboard.

### 5.1 Dashboard Infrastructure
- [x] Implement `GET /api/v1/dashboard`:
  - [x] Combined dashboard data endpoint
- [x] Implement `GET /api/v1/dashboard/stats`:
  - [x] Aggregated statistics (total, year, month)
- [x] Implement `GET /api/v1/dashboard/weekly`:
  - [x] Current week stats by sport
- [x] Implement `GET /api/v1/dashboard/recent`:
  - [x] Recent activities list
- [x] Implement `GET /api/v1/dashboard/sports`:
  - [x] Stats by sport type
- [ ] Implement `GET /api/v1/dashboard/config`:
  - [ ] Widget configuration (deferred)
- [ ] Implement `PUT /api/v1/dashboard/config`:
  - [ ] Update widget configuration (deferred)
- [x] Create `internal/storage/stats.go`:
  - [x] Aggregation queries

### 5.2 Widget System
- [x] Create `web/src/components/dashboard/widget-wrapper.tsx`:
  - [x] Card container
  - [x] Title, action buttons
  - [x] Loading state
- [ ] Create `web/src/components/dashboard/widget-grid.tsx`:
  - [ ] CSS Grid layout (using basic grid for now)
  - [ ] Widget sizing (33%, 50%, 66%, 100%)
  - [ ] Responsive behavior
- [ ] Create `web/src/stores/dashboard.ts`:
  - [ ] Widget order (deferred)
  - [ ] Widget configuration (deferred)
  - [ ] Persist to API (deferred)

### 5.3 Core Widgets (Text/Stats)
- [x] Create `stats-summary.tsx`:
  - [x] Key stats cards (activities, distance, time, elevation)
  - [x] Year/month/total breakdown
- [x] Create `recent-activities.tsx`:
  - [x] N recent activities
  - [x] "View all" link
- [x] Create `weekly-stats.tsx`:
  - [x] Current week by sport type
  - [x] Summary totals
- [x] Create `sport-breakdown.tsx`:
  - [x] Sport type distribution
  - [x] Progress bars with percentages
- [ ] Create `training-goals.tsx`:
  - [ ] Progress bars for goals (deferred)
  - [ ] Weekly/monthly/yearly/lifetime tabs (deferred)

### 5.4 Dashboard Page
- [x] Create `web/src/pages/dashboard.tsx`:
  - [x] Fetch dashboard data
  - [x] Render widget grid
  - [x] Handle loading/error states
  - [x] Handle unauthenticated state
  - [x] Handle empty state
- [x] Create `web/src/lib/api/dashboard.ts`:
  - [x] Dashboard API hooks
  - [x] TypeScript types

---

## Phase 6: Charts & Visualizations ✓

**Goal:** ECharts-based visualizations for dashboard and analytics.

### 6.1 ECharts Infrastructure
- [x] Create `web/src/components/charts/echarts-wrapper.tsx`:
  - [x] Tree-shaking with selective component imports
  - [x] Resize handling with event listener
  - [x] Loading state with skeleton
- [x] Create chart color constants
- [x] Create default grid/tooltip configurations

### 6.2 Line Charts
- [x] Create `web/src/components/charts/line-chart.tsx`:
  - [x] Generic line chart component
  - [x] Area style option
  - [x] Multi-series support
  - [x] Data zoom support
- [ ] Create FTP history chart (deferred)
- [ ] Create weight history chart (deferred)
- [ ] Create training load chart (deferred)
- [ ] Create Eddington history chart (deferred)

### 6.3 Bar Charts
- [x] Create `web/src/components/charts/bar-chart.tsx`:
  - [x] Generic bar chart component
  - [x] Horizontal option
  - [x] Value labels option
- [x] Create stacked bar chart variant
- [ ] Create heart rate zone chart (deferred - needs stream data)
- [ ] Create power zone chart (deferred - needs stream data)

### 6.4 Pie/Donut Charts
- [x] Create sport type distribution chart (donut style)
  - [x] Activity count mode
  - [x] Distance mode
  - [x] Sport-specific colors

### 6.5 Calendar Heatmap
- [x] Create `ActivityCalendarChart`:
  - [x] GitHub-style activity grid
  - [x] Intensity coloring based on activity count
  - [x] Tooltips with date and count

### 6.6 Activity Stream Charts
- [ ] Create combined stream profile chart (deferred - needs stream data)
- [ ] Create elevation profile chart (deferred - needs stream data)
- [ ] Create power curve chart (deferred - needs stream data)

### 6.7 Dashboard Chart Widgets
- [x] Create `MonthlyChart` widget:
  - [x] Year selector
  - [x] Metric toggle (distance/count/time)
- [x] Create `SportChart` widget:
  - [x] Metric toggle (count/distance)
- [x] Create `ActivityCalendar` widget:
  - [x] Year navigation
- [ ] Create `peak-power-outputs.tsx` widget (deferred - needs stream data)
- [ ] Create `heart-rate-zones.tsx` widget (deferred - needs stream data)
- [ ] Create `training-load.tsx` widget (deferred)

### 6.8 Chart Data API Endpoints
- [x] Implement `GET /api/v1/dashboard/monthly`:
  - [x] Monthly aggregated stats
  - [x] Optional year filter
- [x] Implement `GET /api/v1/dashboard/yearly`:
  - [x] Yearly aggregated stats
- [x] Implement `GET /api/v1/dashboard/calendar`:
  - [x] Daily activity counts for calendar heatmap

---

## Phase 7: Maps & Heatmap

**Goal:** Leaflet maps with route visualization.

### 7.1 Map Infrastructure
- [ ] Create `web/src/components/maps/base-map.tsx`:
  - [ ] react-leaflet setup
  - [ ] Tile layer configuration
  - [ ] Controls
- [ ] Create `web/src/lib/constants/tile-layers.ts`:
  - [ ] OpenStreetMap
  - [ ] Satellite options
  - [ ] Grayscale filter
- [ ] Create polyline decoder utility

### 7.2 Activity Map
- [ ] Create `web/src/components/maps/activity-map.tsx`:
  - [ ] Route polyline
  - [ ] Start/end markers
  - [ ] Elevation coloring (optional)
  - [ ] Fit bounds to route
- [ ] Integrate into activity detail page

### 7.3 Heatmap Page
- [ ] Implement `GET /api/v1/stats/heatmap`:
  - [ ] Return polylines with filters
  - [ ] Country statistics
- [ ] Create `web/src/pages/heatmap.tsx`:
  - [ ] Full-screen map
  - [ ] Filter sidebar
  - [ ] Route count display
  - [ ] Country coverage stats
- [ ] Create `web/src/components/maps/heatmap.tsx`:
  - [ ] Multi-polyline rendering
  - [ ] Configurable color
  - [ ] Performance optimization for many routes
- [ ] Add filters:
  - [ ] Sport type
  - [ ] Date range
  - [ ] Commute filter
  - [ ] Workout type

### 7.4 Segment Map
- [ ] Create `web/src/components/maps/segment-map.tsx`:
  - [ ] Segment route
  - [ ] Start/end points
  - [ ] Gradient visualization

### 7.5 Virtual World Maps (Future)
- [ ] Research Zwift map tiles
- [ ] Create virtual world tile layers
- [ ] Detect virtual activities
- [ ] Show appropriate map

---

## Phase 8: Advanced Features

**Goal:** Segments, gear, maintenance, calendar, photos, challenges.

### 8.1 Segments
- [ ] Create `schema/migrations/004_segments.sql`
- [ ] Create `internal/storage/segments.go`
- [ ] Create `internal/strava/segments.go`:
  - [ ] Fetch segment details
  - [ ] Fetch segment efforts
- [ ] Create `internal/importer/segments.go`
- [ ] Implement segment API endpoints
- [ ] Create `web/src/pages/segments.tsx`:
  - [ ] Segment list with filters
  - [ ] Search
  - [ ] Sorting
- [ ] Create segment detail modal:
  - [ ] Segment map
  - [ ] Personal efforts list
  - [ ] PR progression chart

### 8.2 Gear
- [ ] Create `schema/migrations/005_gear.sql`
- [ ] Create `internal/storage/gear.go`
- [ ] Create `internal/strava/gear.go`
- [ ] Implement gear API endpoints
- [ ] Create `web/src/pages/gear.tsx`:
  - [ ] Gear list with stats
  - [ ] Active vs retired toggle
  - [ ] Per-gear metrics
- [ ] Create custom gear functionality:
  - [ ] Create custom gear form
  - [ ] Hashtag linking

### 8.3 Gear Maintenance
- [ ] Create maintenance tables in schema
- [ ] Create `internal/storage/maintenance.go`
- [ ] Implement maintenance API endpoints
- [ ] Create `web/src/pages/gear-maintenance.tsx`:
  - [ ] Component list by gear
  - [ ] Progress indicators
  - [ ] Add component form
  - [ ] Log maintenance action
  - [ ] Maintenance history

### 8.4 Calendar
- [ ] Create `web/src/pages/calendar.tsx`
- [ ] Create `web/src/components/calendar/month-view.tsx`:
  - [ ] CSS Grid calendar
  - [ ] Activities on dates
  - [ ] Color by sport type
- [ ] Create month navigation
- [ ] Create monthly summary stats
- [ ] Create day detail modal

### 8.5 Photos
- [ ] Create `internal/storage/photos.go`
- [ ] Create `internal/strava/photos.go`
- [ ] Implement photos API endpoints
- [ ] Create `web/src/pages/photos.tsx`:
  - [ ] Masonry/flex grid
  - [ ] Lazy loading
  - [ ] Sport type filter
  - [ ] Country filter
- [ ] Integrate lightbox:
  - [ ] Add lightbox library
  - [ ] Slideshow mode
  - [ ] Activity link overlay

### 8.6 Challenges
- [ ] Create `schema/migrations/006_challenges.sql`
- [ ] Create `internal/storage/challenges.go`
- [ ] Create challenge scraping (from public profile)
- [ ] Implement challenges API endpoints
- [ ] Create `web/src/pages/challenges.tsx`:
  - [ ] Grouped by month
  - [ ] Badge images
  - [ ] Strava links
- [ ] Create `most-recent-challenges.tsx` dashboard widget
- [ ] Create `challenge-consistency.tsx` dashboard widget

---

## Phase 9: Analytics

**Goal:** Eddington, best efforts, training load, rewind.

### 9.1 Eddington Number
- [ ] Implement `GET /api/v1/stats/eddington`:
  - [ ] Calculate Eddington number
  - [ ] Days distribution
  - [ ] Progress tracking
- [ ] Create `web/src/pages/eddington.tsx`:
  - [ ] Current number display
  - [ ] Distribution chart
  - [ ] History chart
  - [ ] "Days needed" table
  - [ ] Sport type tabs
- [ ] Create `eddington.tsx` dashboard widget

### 9.2 Best Efforts
- [ ] Create best efforts tables in schema
- [ ] Calculate best efforts during import
- [ ] Implement `GET /api/v1/stats/best-efforts`:
  - [ ] PRs by distance
  - [ ] History per distance
- [ ] Create `web/src/pages/best-efforts.tsx`:
  - [ ] Sport type tabs
  - [ ] PR list by distance
  - [ ] Chart showing records over time
- [ ] Create distance detail modal:
  - [ ] All efforts for distance
  - [ ] Progression chart

### 9.3 Training Load
- [ ] Research training load algorithms (CTL/ATL/TSB)
- [ ] Create training load calculations
- [ ] Implement `GET /api/v1/stats/training-load`
- [ ] Create training load detail page:
  - [ ] Fitness curve
  - [ ] Fatigue curve
  - [ ] Form curve
  - [ ] Date range selection
- [ ] Create `training-load.tsx` dashboard widget

### 9.4 Strava Rewind
- [ ] Implement `GET /api/v1/stats/rewind/:year`:
  - [ ] All rewind statistics
- [ ] Create `web/src/pages/rewind.tsx`:
  - [ ] Year selector
  - [ ] Comparison mode
  - [ ] Metric sections:
    - [ ] Total activities by month
    - [ ] Distance by month
    - [ ] Elevation by month
    - [ ] Time by sport type
    - [ ] Active vs rest days
    - [ ] Start times distribution
    - [ ] PRs by month
    - [ ] Activity locations map
    - [ ] Streaks
    - [ ] Carbon saved
    - [ ] Kudos received
    - [ ] Biggest activities
    - [ ] Random photo

---

## Phase 10: WASM Mode

**Goal:** Browser-only version with DuckDB-WASM.

### 10.1 DuckDB-WASM Setup
- [ ] Install `@duckdb/duckdb-wasm`
- [ ] Create `web/src/lib/db/wasm-client.ts`:
  - [ ] Initialize DuckDB-WASM
  - [ ] OPFS persistence setup
  - [ ] Implement DataSource interface
- [ ] Create schema migration for browser:
  - [ ] Embed SQL as strings
  - [ ] Version tracking in DuckDB

### 10.2 Strava Direct Integration
- [ ] Create `web/src/lib/strava/client.ts`:
  - [ ] Direct Strava API calls from browser
  - [ ] OAuth implicit flow handling
- [ ] Create OAuth callback page for browser mode
- [ ] Handle token storage (localStorage)
- [ ] Rate limit tracking in browser

### 10.3 Browser Importer
- [ ] Create `web/src/lib/importer/`:
  - [ ] Activity importer (browser version)
  - [ ] Progress tracking
  - [ ] Rate limit awareness
- [ ] Create import UI:
  - [ ] Start import button
  - [ ] Progress display
  - [ ] Estimated time remaining
  - [ ] Pause/resume

### 10.4 Mode Detection & Switching
- [ ] Create build configuration for WASM mode
- [ ] Create `web/src/lib/mode.ts`:
  - [ ] Detect current mode
  - [ ] Feature flags per mode
- [ ] Update DataSource factory
- [ ] Hide server-only features in WASM mode:
  - [ ] Webhooks
  - [ ] Scheduled imports
  - [ ] Challenge scraping

### 10.5 Static Hosting Build
- [ ] Create separate Vite config for WASM build
- [ ] Configure for static hosting (GitHub Pages, etc.)
- [ ] Create deployment documentation

### 10.6 Token Exchange Service (Optional)
- [ ] Create Cloudflare Worker for token exchange
- [ ] Or document user-provided serverless option
- [ ] Handle client_secret securely

---

## Phase 11: Polish

**Goal:** PWA, internationalization, settings, badges, final polish.

### 11.1 PWA Support
- [ ] Create `web/public/manifest.json`
- [ ] Create service worker for offline support
- [ ] Add app icons (multiple sizes)
- [ ] Configure installable prompt
- [ ] Test PWA installation

### 11.2 Internationalization
- [ ] Install i18n library (react-i18next)
- [ ] Create translation structure
- [ ] Extract all strings
- [ ] Create locale files:
  - [ ] en_US
  - [ ] (others as contributed)
- [ ] Add locale selector
- [ ] Format numbers/dates per locale

### 11.3 Unit System
- [ ] Create unit conversion utilities
- [ ] Add unit system selector
- [ ] Apply unit formatting throughout app
- [ ] Persist preference

### 11.4 Settings Page
- [ ] Create `web/src/pages/settings.tsx`:
  - [ ] Profile section (photo, name)
  - [ ] Display settings (units, locale, theme)
  - [ ] Import settings
  - [ ] Heart rate zones configuration
  - [ ] FTP history management
  - [ ] Weight history management
  - [ ] Dashboard configuration
  - [ ] Notification settings
  - [ ] Data export

### 11.5 SVG Badges
- [ ] Create badge generation in Go:
  - [ ] Strava stats badge
  - [ ] PB badges per sport
  - [ ] Zwift badge
- [ ] Create badge API endpoints
- [ ] Create badge preview in settings
- [ ] Create embed code generator

### 11.6 Webhooks
- [ ] Implement Strava webhook subscription:
  - [ ] `GET /api/v1/webhooks/strava` - validation
  - [ ] `POST /api/v1/webhooks/strava` - receive events
- [ ] Handle activity create/update/delete events
- [ ] Trigger incremental import on webhook

### 11.7 Scheduler
- [ ] Create `internal/scheduler/scheduler.go`
- [ ] Configurable scheduled jobs:
  - [ ] Periodic full sync
  - [ ] Maintenance check notifications
  - [ ] App update check
- [ ] Create scheduler configuration in settings

### 11.8 Notifications
- [ ] Integrate Shoutrrr for notifications
- [ ] Create notification settings UI
- [ ] Send notifications for:
  - [ ] Import completion
  - [ ] Maintenance due
  - [ ] New features (optional)

### 11.9 Final Polish
- [ ] Accessibility audit:
  - [ ] ARIA labels
  - [ ] Keyboard navigation
  - [ ] Screen reader testing
- [ ] Performance audit:
  - [ ] Lighthouse scores
  - [ ] Bundle size optimization
  - [ ] Lazy loading verification
- [ ] Error handling review:
  - [ ] User-friendly error messages
  - [ ] Error boundaries
  - [ ] Retry mechanisms
- [ ] Mobile responsiveness testing
- [ ] Cross-browser testing

---

## Testing Milestones

Throughout implementation, maintain test coverage:

### Unit Tests
- [ ] Go: All storage repositories
- [ ] Go: All domain logic
- [ ] Go: API handlers
- [ ] React: All utility functions
- [ ] React: All hooks
- [ ] React: Key components

### Integration Tests
- [ ] Go: Full import flow
- [ ] Go: OAuth flow
- [ ] Go: API endpoint integration
- [ ] React: Page rendering with mock data

### E2E Tests (Playwright)
- [ ] Authentication flow
- [ ] Dashboard loading
- [ ] Activities list filtering
- [ ] Activity detail view
- [ ] Heatmap interaction
- [ ] Settings persistence

---

## Documentation Milestones

- [ ] README with quick start
- [ ] Deployment guide (Docker, binary)
- [ ] Configuration reference
- [ ] Development setup guide
- [ ] API documentation
- [ ] WASM mode documentation
- [ ] Contributing guide

---

*This TODO list should be updated as implementation progresses. Check off items as they are completed and add new items as scope is refined.*
