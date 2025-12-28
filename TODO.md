# Statistics for Strava - Implementation Plan

This document provides a phased implementation plan for the Statistics for Strava application. Each phase builds upon the previous one, delivering incremental value.

---

## Overview

| Phase | Focus | Deliverable |
|-------|-------|-------------|
| 0 | Project Setup | Repository, tooling, CI/CD |
| 1 | Core Backend | Go server, Strava OAuth, basic import |
| 2 | Database & Storage | DuckDB schema, activity storage |
| 3 | Frontend Foundation | React shell, routing, shadcn setup |
| 4 | Activities Feature | Activity list, filters, detail view |
| 5 | Dashboard | Widget system, core widgets |
| 6 | Charts & Visualizations | ECharts integration, all chart types |
| 7 | Maps & Heatmap | Leaflet integration, route visualization |
| 8 | Advanced Features | Segments, gear, maintenance, calendar |
| 9 | Analytics | Eddington, best efforts, training load |
| 10 | WASM Mode | Browser-only version with DuckDB-WASM |
| 11 | Polish | PWA, i18n, settings, badges |

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

## Phase 1: Core Backend

**Goal:** Go server running with Strava OAuth working.

### 1.1 Configuration System
- [ ] Create `internal/config/config.go` with config struct
- [ ] Create `internal/config/loader.go` with Viper integration
- [ ] Support configuration via:
  - [ ] Environment variables
  - [ ] Config file (YAML)
  - [ ] Command-line flags
- [ ] Configuration fields:
  - [ ] Server port
  - [ ] Data directory path
  - [ ] Strava client ID/secret
  - [ ] Strava redirect URI
  - [ ] Log level

### 1.2 HTTP Server
- [ ] Create `internal/api/router.go` with chi router
- [ ] Create `internal/api/middleware.go`:
  - [ ] Request logging middleware
  - [ ] CORS middleware (for dev mode)
  - [ ] Recovery middleware
- [ ] Create `cmd/stata/serve.go` command
- [ ] Implement graceful shutdown
- [ ] Serve static files from embedded React build
- [ ] Verify server starts and serves placeholder page

### 1.3 Strava OAuth Flow
- [ ] Create `internal/strava/oauth.go`:
  - [ ] OAuth2 config setup
  - [ ] Generate auth URL
  - [ ] Exchange code for tokens
  - [ ] Refresh expired tokens
- [ ] Create `internal/api/handlers/auth.go`:
  - [ ] `GET /api/v1/auth/strava` - redirect to Strava
  - [ ] `GET /api/v1/auth/strava/callback` - handle callback
  - [ ] `GET /api/v1/auth/status` - check auth status
  - [ ] `POST /api/v1/auth/refresh` - force token refresh
- [ ] Store tokens (initially in memory, later in DB)
- [ ] Test OAuth flow end-to-end

### 1.4 Strava API Client
- [ ] Create `internal/strava/client.go`:
  - [ ] HTTP client with auth header injection
  - [ ] Rate limit tracking from response headers
  - [ ] Automatic token refresh on 401
- [ ] Create `internal/strava/types.go`:
  - [ ] Activity struct matching Strava API
  - [ ] Athlete struct
  - [ ] Gear struct
  - [ ] Stream types
- [ ] Create `internal/strava/ratelimit.go`:
  - [ ] Track 15-min and daily limits
  - [ ] Expose rate limit status

### 1.5 Basic Activity Fetching
- [ ] Create `internal/strava/activities.go`:
  - [ ] `GetActivities(page, perPage)` - list activities
  - [ ] `GetActivity(id)` - single activity detail
  - [ ] `GetActivityStreams(id, types)` - stream data
- [ ] Create `internal/api/handlers/activities.go`:
  - [ ] `GET /api/v1/activities` - proxy to Strava (temporary)
- [ ] Test fetching activities from Strava API

---

## Phase 2: Database & Storage

**Goal:** DuckDB schema set up, activities persisted locally.

### 2.1 DuckDB Connection
- [ ] Create `internal/storage/db.go`:
  - [ ] Open/create DuckDB file
  - [ ] Connection pool management
  - [ ] Close on shutdown
- [ ] Configure database path from config
- [ ] Test database creation

### 2.2 Schema Migrations
- [ ] Create `internal/storage/migrations.go`:
  - [ ] Migration tracking table
  - [ ] Run pending migrations
  - [ ] Rollback support (optional)
- [ ] Create `schema/migrations/001_initial.sql`:
  - [ ] Athletes table
  - [ ] Auth tokens table
- [ ] Create `schema/migrations/002_activities.sql`:
  - [ ] Activities table with all fields
  - [ ] Indexes
- [ ] Create `schema/migrations/003_streams.sql`:
  - [ ] Activity streams table
- [ ] Embed migrations using `go:embed`

### 2.3 Activity Repository
- [ ] Create `internal/storage/activities.go`:
  - [ ] `Insert(activity)` - upsert activity
  - [ ] `GetByID(id)` - fetch single activity
  - [ ] `List(filters, pagination)` - filtered list
  - [ ] `GetTotals(filters)` - sum aggregates
- [ ] Create `internal/storage/queries/activities.sql`:
  - [ ] Parameterized queries
- [ ] Unit tests for repository

### 2.4 Importer Service
- [ ] Create `internal/importer/importer.go`:
  - [ ] Orchestrate full import
  - [ ] Track import progress
  - [ ] Handle rate limits gracefully
- [ ] Create `internal/importer/activities.go`:
  - [ ] Fetch activities in batches
  - [ ] Transform Strava → domain model
  - [ ] Persist to database
- [ ] Create `cmd/stata/import.go` command:
  - [ ] Trigger manual import
  - [ ] Show progress
- [ ] Test importing real activities

### 2.5 Athlete & Token Storage
- [ ] Create `internal/storage/athletes.go`:
  - [ ] Store athlete profile
  - [ ] Store/retrieve OAuth tokens
- [ ] Update OAuth flow to persist tokens
- [ ] Retrieve tokens on server start

---

## Phase 3: Frontend Foundation

**Goal:** React app shell with routing and basic layout.

### 3.1 App Shell
- [ ] Create `web/src/app.tsx` with router setup
- [ ] Create route definitions for all pages
- [ ] Create `web/src/pages/` placeholder pages:
  - [ ] Dashboard
  - [ ] Activities
  - [ ] Segments
  - [ ] Heatmap
  - [ ] Calendar
  - [ ] Best Efforts
  - [ ] Eddington
  - [ ] Gear
  - [ ] Photos
  - [ ] Challenges
  - [ ] Rewind
  - [ ] Settings

### 3.2 Layout Components
- [ ] Create `web/src/components/layout/header.tsx`:
  - [ ] App title
  - [ ] Navigation menu
  - [ ] Settings link
- [ ] Create `web/src/components/layout/sidebar.tsx`:
  - [ ] Navigation links
  - [ ] Active state
  - [ ] Collapsible on mobile
- [ ] Create `web/src/components/layout/breadcrumb.tsx`
- [ ] Create `web/src/components/layout/page-layout.tsx`:
  - [ ] Combines header + sidebar + content area

### 3.3 shadcn/ui Base Components
- [ ] Add essential components via shadcn CLI:
  - [ ] Button
  - [ ] Card
  - [ ] Dialog
  - [ ] Dropdown Menu
  - [ ] Input
  - [ ] Label
  - [ ] Select
  - [ ] Table
  - [ ] Tabs
  - [ ] Toast
  - [ ] Tooltip
  - [ ] Badge
  - [ ] Skeleton
  - [ ] Switch
  - [ ] Calendar
  - [ ] Popover

### 3.4 Data Layer Setup
- [ ] Create `web/src/lib/db/types.ts` - DataSource interface
- [ ] Create `web/src/lib/db/api-client.ts` - REST implementation
- [ ] Create `web/src/lib/db/index.ts` - factory function
- [ ] Create `web/src/types/` - shared TypeScript types:
  - [ ] Activity
  - [ ] Segment
  - [ ] Gear
  - [ ] Athlete
  - [ ] Common types (pagination, filters)

### 3.5 State Management
- [ ] Create `web/src/stores/settings.ts`:
  - [ ] Unit system (metric/imperial)
  - [ ] Theme preference
  - [ ] Locale
- [ ] Create `web/src/stores/filters.ts`:
  - [ ] Global filter state
  - [ ] URL sync

### 3.6 TanStack Query Setup
- [ ] Configure QueryClient with defaults
- [ ] Create query key factory
- [ ] Create base hooks structure

### 3.7 Utility Functions
- [ ] Create `web/src/lib/utils/units.ts`:
  - [ ] Distance conversion (m ↔ km ↔ mi)
  - [ ] Speed conversion (m/s ↔ km/h ↔ mph)
  - [ ] Elevation conversion (m ↔ ft)
  - [ ] Pace formatting
- [ ] Create `web/src/lib/utils/dates.ts`:
  - [ ] Date formatting
  - [ ] Duration formatting
- [ ] Create `web/src/lib/utils/cn.ts` - className helper
- [ ] Create `web/src/lib/constants/sport-types.ts`
- [ ] Create `web/src/lib/constants/colors.ts`

### 3.8 Integration
- [ ] Embed React build in Go binary using `go:embed`
- [ ] Serve React from Go server
- [ ] Configure Vite proxy for API in dev mode
- [ ] Test full stack integration

---

## Phase 4: Activities Feature

**Goal:** Complete activities list with filtering, sorting, and detail view.

### 4.1 Activities API Endpoints
- [ ] Implement `GET /api/v1/activities`:
  - [ ] Pagination
  - [ ] All filter parameters
  - [ ] Sorting
  - [ ] Search
- [ ] Implement `GET /api/v1/activities/:id`:
  - [ ] Full activity detail
  - [ ] Include gear info
- [ ] Implement `GET /api/v1/activities/:id/streams`:
  - [ ] Return stream data
- [ ] Implement `GET /api/v1/activities/:id/photos`:
  - [ ] Return photo URLs

### 4.2 Activities List Page
- [ ] Create `web/src/hooks/use-activities.ts`:
  - [ ] TanStack Query hook
  - [ ] Filter state management
  - [ ] Pagination state
- [ ] Create `web/src/components/activities/activity-filters.tsx`:
  - [ ] Sport type multi-select
  - [ ] Date range picker
  - [ ] Country dropdown
  - [ ] Gear dropdown
  - [ ] Device dropdown
  - [ ] Commute toggle
  - [ ] Workout type
  - [ ] Search input
  - [ ] Clear filters button
- [ ] Create `web/src/components/activities/activity-table.tsx`:
  - [ ] TanStack Table integration
  - [ ] Virtual scrolling for large lists
  - [ ] Sortable columns
  - [ ] Sticky header
  - [ ] Totals row
- [ ] Create `web/src/pages/activities.tsx`:
  - [ ] Combine filters + table
  - [ ] URL state sync

### 4.3 Activity Detail Page
- [ ] Create `web/src/pages/activity-detail.tsx`
- [ ] Create `web/src/components/activities/activity-header.tsx`:
  - [ ] Name, date, sport icon
  - [ ] Link to Strava
- [ ] Create `web/src/components/activities/activity-stats.tsx`:
  - [ ] Grid of key metrics
  - [ ] Conditional display based on sport type
- [ ] Create `web/src/components/activities/activity-map.tsx`:
  - [ ] Route visualization (placeholder, full in Phase 7)
- [ ] Create `web/src/components/activities/activity-weather.tsx`:
  - [ ] Weather conditions display

### 4.4 Activity Streams (Preparation)
- [ ] Create stream data hooks
- [ ] Create stream data types
- [ ] Prepare for chart integration (Phase 6)

---

## Phase 5: Dashboard

**Goal:** Configurable widget-based dashboard.

### 5.1 Dashboard Infrastructure
- [ ] Implement `GET /api/v1/dashboard/stats`:
  - [ ] Aggregated statistics
- [ ] Implement `GET /api/v1/dashboard/weekly`:
  - [ ] Current week stats by sport
- [ ] Implement `GET /api/v1/dashboard/config`:
  - [ ] Widget configuration
- [ ] Implement `PUT /api/v1/dashboard/config`:
  - [ ] Update widget configuration
- [ ] Create `internal/storage/stats.go`:
  - [ ] Aggregation queries

### 5.2 Widget System
- [ ] Create `web/src/components/dashboard/widget-grid.tsx`:
  - [ ] CSS Grid layout
  - [ ] Widget sizing (33%, 50%, 66%, 100%)
  - [ ] Responsive behavior
- [ ] Create `web/src/components/dashboard/widget-wrapper.tsx`:
  - [ ] Card container
  - [ ] Title, action buttons
  - [ ] Loading state
  - [ ] Error state
- [ ] Create `web/src/stores/dashboard.ts`:
  - [ ] Widget order
  - [ ] Widget configuration
  - [ ] Persist to API

### 5.3 Core Widgets (Text/Stats)
- [ ] Create `most-recent-activities.tsx`:
  - [ ] N recent activities
  - [ ] "View all" link
- [ ] Create `intro-text.tsx`:
  - [ ] Total stats summary
- [ ] Create `weekly-stats.tsx`:
  - [ ] Current week by sport type
- [ ] Create `training-goals.tsx`:
  - [ ] Progress bars for goals
  - [ ] Weekly/monthly/yearly/lifetime tabs

### 5.4 Dashboard Page
- [ ] Create `web/src/pages/dashboard.tsx`:
  - [ ] Fetch widget config
  - [ ] Render widget grid
  - [ ] Handle loading/error states
- [ ] Create `web/src/hooks/use-dashboard-stats.ts`

---

## Phase 6: Charts & Visualizations

**Goal:** All ECharts-based visualizations working.

### 6.1 ECharts Infrastructure
- [ ] Create `web/src/components/charts/echarts-wrapper.tsx`:
  - [ ] Lazy loading
  - [ ] Theme integration
  - [ ] Resize handling
  - [ ] Loading state
- [ ] Create chart theme configuration
- [ ] Create base chart option builders

### 6.2 Line Charts
- [ ] Create `web/src/components/charts/line-chart.tsx`:
  - [ ] Generic line chart component
- [ ] Create monthly stats chart (multi-year overlay)
- [ ] Create yearly stats chart
- [ ] Create FTP history chart
- [ ] Create weight history chart
- [ ] Create training load chart
- [ ] Create Eddington history chart

### 6.3 Bar Charts
- [ ] Create `web/src/components/charts/bar-chart.tsx`:
  - [ ] Generic bar chart component
- [ ] Create Eddington distribution chart
- [ ] Create heart rate zone chart (stacked)
- [ ] Create power zone chart (stacked)
- [ ] Create best effort bars

### 6.4 Pie/Donut Charts
- [ ] Create `web/src/components/charts/pie-chart.tsx`:
  - [ ] Generic pie/donut component
- [ ] Create weekday distribution chart
- [ ] Create daytime distribution chart
- [ ] Create gear usage chart
- [ ] Create sport type distribution chart

### 6.5 Calendar Heatmap
- [ ] Create `web/src/components/charts/calendar-heatmap.tsx`:
  - [ ] GitHub-style activity grid
  - [ ] Intensity coloring
  - [ ] Tooltips

### 6.6 Activity Stream Charts
- [ ] Create combined stream profile chart:
  - [ ] Heart rate line
  - [ ] Power line
  - [ ] Cadence line
  - [ ] Elevation area
  - [ ] Sync X-axis (time/distance toggle)
- [ ] Create elevation profile chart
- [ ] Create power curve chart (best efforts)

### 6.7 Dashboard Chart Widgets
- [ ] Create `peak-power-outputs.tsx` widget
- [ ] Create `heart-rate-zones.tsx` widget
- [ ] Create `activity-grid.tsx` widget
- [ ] Create `monthly-stats.tsx` widget (chart version)
- [ ] Create `training-load.tsx` widget
- [ ] Create `weekday-stats.tsx` widget
- [ ] Create `daytime-stats.tsx` widget
- [ ] Create `gear-stats.tsx` widget
- [ ] Create `ftp-history.tsx` widget
- [ ] Create `weight-history.tsx` widget

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
