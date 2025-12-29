# Statistics for Strava - Implementation Plan

This document provides a phased implementation plan for the Statistics for Strava application. Each phase builds upon the previous one, delivering incremental value.

---

## Overview

| Phase | Focus                   | Deliverable                              | Status      |
| ----- | ----------------------- | ---------------------------------------- | ----------- |
| 0     | Project Setup           | Repository, tooling, CI/CD               | ✓           |
| 1     | Core Backend            | Go server, Strava OAuth, basic import    | ✓           |
| 2     | Database & Storage      | SQLite schema, activity storage          | ✓           |
| 3     | Frontend Foundation     | React shell, routing, shadcn setup       | ✓           |
| 4     | Activities Feature      | Activity list, filters, detail view      | ✓           |
| 5     | Dashboard               | Widget system, core widgets              | ✓           |
| 6     | Charts & Visualizations | ECharts integration, all chart types     | ✓           |
| 7     | Maps & Heatmap          | Leaflet integration, route visualization | ✓           |
| 8     | Advanced Features       | Segments, gear, maintenance, calendar    | ✓           |
| 9     | Analytics               | Eddington, best efforts, training load   | ✓           |
| 10    | Feature Parity & Beyond | Gap features + unique enhancements       | **NEXT**    |
| 11    | WASM Mode               | Browser-only version with SQLite-WASM    | ✓           |
| 12    | Polish                  | PWA, i18n, settings, badges              |             |

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
  - [x] `modernc.org/sqlite`
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

**Goal:** SQLite schema set up, activities persisted locally.

### 2.1 SQLite Connection

- [x] Create `internal/storage/db.go`:
  - [x] Open/create SQLite file
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
- [x] Implement `GET /api/v1/dashboard/config`:
  - [x] Return current widget configuration
  - [x] Widget order list
  - [x] Per-widget settings
- [x] Implement `PUT /api/v1/dashboard/config`:
  - [x] Update widget order
  - [x] Update widget sizes
  - [x] Update per-widget configuration
- [x] Create `internal/storage/stats.go`:
  - [x] Aggregation queries

### 5.2 Widget System

- [x] Create `web/src/components/dashboard/widget-wrapper.tsx`:
  - [x] Card container
  - [x] Title, action buttons
  - [x] Loading state
- [x] Create `web/src/components/dashboard/widget-grid.tsx`:
  - [x] CSS Grid layout with drag-and-drop reordering
  - [x] Widget sizing (33%, 50%, 66%, 100%)
  - [x] Responsive behavior (stack on mobile)
  - [x] Widget show/hide toggle
- [x] Create `web/src/stores/dashboard.ts`:
  - [x] Widget order state
  - [x] Widget visibility state
  - [x] Per-widget configuration state
  - [x] Persist to API on change
  - [x] Load from API on mount

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
- [x] Create `training-goals.tsx`:
  - [x] Goal types: distance, elevation, moving time
  - [x] Weekly/monthly/yearly/lifetime tabs
  - [x] Configurable per sport type
  - [x] Visual progress indicators (horizontal bars)
  - [x] Percentage completion display
  - [x] Create `GET /api/v1/goals` endpoint
  - [x] Create `PUT /api/v1/goals` endpoint
  - [x] Goal configuration UI

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
- [x] Create FTP history chart:
  - [x] Line chart showing FTP evolution over time
  - [x] Date-keyed data points
  - [x] Separate series for cycling/running FTP
  - [x] Create `GET /api/v1/athlete/ftp` endpoint
- [x] Create weight history chart:
  - [x] Line chart showing body weight changes
  - [x] Date-keyed data points
  - [x] Create `GET /api/v1/athlete/weight` endpoint
- [x] Create training load chart:
  - [x] Fitness and fatigue curves over time
  - [x] Daily training stress scores
  - [x] Form (fitness - fatigue) line
- [x] Create Eddington history chart:
  - [x] Eddington number progression over time
  - [x] Show when each new number was achieved

### 6.3 Bar Charts

- [x] Create `web/src/components/charts/bar-chart.tsx`:
  - [x] Generic bar chart component
  - [x] Horizontal option
  - [x] Value labels option
- [x] Create stacked bar chart variant
- [x] Create heart rate zone chart:
  - [x] Time distribution across 5 HR zones
  - [x] Stacked horizontal bars
  - [x] Configurable zones (relative % or absolute BPM)
  - [x] Zone colors (Recovery: light blue, Aerobic: green, Tempo: yellow, Threshold: orange, Anaerobic: red)
- [x] Create power zone chart:
  - [x] Time spent in each power zone
  - [x] Stacked horizontal bars
  - [x] FTP-based zone calculation

### 6.4 Pie/Donut Charts

- [x] Create sport type distribution chart (donut style)
  - [x] Activity count mode
  - [x] Distance mode
  - [x] Sport-specific colors
- [x] Create daytime stats chart:
  - [x] Activity distribution by time of day (morning/afternoon/evening/night)
  - [x] Donut visualization
- [x] Create weekday stats chart:
  - [x] Activity distribution by day of week
  - [x] Donut visualization

### 6.5 Calendar Heatmap

- [x] Create `ActivityCalendarChart`:
  - [x] GitHub-style activity grid
  - [x] Intensity coloring based on activity count
  - [x] Tooltips with date and count

### 6.6 Activity Stream Charts

- [x] Create combined stream profile chart:
  - [x] Heart rate over distance/time
  - [x] Power over distance/time
  - [x] Cadence over distance/time
  - [x] Elevation over distance/time
  - [x] Synchronized tooltips across series
  - [x] Toggle visibility of each stream
- [x] Create elevation profile chart:
  - [x] Elevation vs distance
  - [x] Gradient coloring option
  - [x] Min/max/avg elevation stats
- [x] Create power curve chart:
  - [x] Best power efforts at different durations
  - [x] Duration intervals: 5s, 10s, 30s, 1m, 5m, 8m, 20m, 1h
  - [x] Comparison with previous periods

### 6.7 Dashboard Chart Widgets

- [x] Create `MonthlyChart` widget:
  - [x] Year selector
  - [x] Metric toggle (distance/count/time)
- [x] Create `SportChart` widget:
  - [x] Metric toggle (count/distance)
- [x] Create `ActivityCalendar` widget:
  - [x] Year navigation
- [x] Create `peak-power-outputs.tsx` widget:
  - [x] All-time best power outputs
  - [x] Duration intervals: 5s, 10s, 30s, 1m, 5m, 8m, 20m, 1h
  - [x] Bar chart visualization
  - [x] "View details" link to power analysis page
- [x] Create `heart-rate-zones.tsx` widget:
  - [x] Time distribution across 5 heart rate zones
  - [x] Configurable zones (relative % or absolute BPM)
  - [x] Zone ranges customizable by date and sport type
  - [x] Horizontal stacked bar visualization
- [x] Create `training-load.tsx` widget:
  - [x] Fitness and fatigue tracking over time
  - [x] Based on activity intensity calculations
  - [x] "View details" link to training load page

### 6.8 Chart Data API Endpoints

- [x] Implement `GET /api/v1/dashboard/monthly`:
  - [x] Monthly aggregated stats
  - [x] Optional year filter
- [x] Implement `GET /api/v1/dashboard/yearly`:
  - [x] Yearly aggregated stats
- [x] Implement `GET /api/v1/dashboard/calendar`:
  - [x] Daily activity counts for calendar heatmap
- [x] Implement `GET /api/v1/stats/power`:
  - [x] Peak power outputs for various durations
  - [x] Best efforts history
- [x] Implement `GET /api/v1/stats/hr-zones`:
  - [x] Time in each heart rate zone
  - [x] Filterable by date range and sport type
- [x] Implement `GET /api/v1/stats/training-load`:
  - [x] Daily training stress scores
  - [x] Fitness and fatigue calculations

---

## Phase 7: Maps & Heatmap ✓

**Goal:** Leaflet maps with route visualization.

### 7.1 Map Infrastructure

- [x] Create `web/src/components/maps/base-map.tsx`:
  - [x] react-leaflet setup
  - [x] Tile layer configuration
  - [x] Controls
- [x] Create `web/src/lib/maps/tile-layers.ts`:
  - [x] OpenStreetMap
  - [x] CartoDB options (light/dark)
  - [x] OpenTopoMap
- [x] Create polyline decoder utility

### 7.2 Activity Map

- [x] Create `web/src/components/maps/activity-map.tsx`:
  - [x] Route polyline
  - [x] Start/end markers
  - [x] Elevation coloring:
    - [x] Color polyline segments by elevation/gradient
    - [x] Legend showing color scale
    - [x] Toggle between flat color and elevation mode
  - [x] Fit bounds to route
- [x] Integrate into activity detail page

### 7.3 Heatmap Page

- [x] Implement `GET /api/v1/stats/heatmap`:
  - [x] Return polylines with filters
  - [x] Country statistics:
    - [x] List of countries with activities
    - [x] Activity count per country
    - [x] Percentage of world coverage
- [x] Create `web/src/pages/heatmap.tsx`:
  - [x] Full-screen map
  - [x] Filter sidebar:
    - [x] Sport type radio buttons
    - [x] Date range picker (from/to)
    - [x] Commute filter (yes/no/all)
    - [x] Workout type filter
    - [x] Clear filters button
    - [x] Collapsible on mobile
  - [x] Route count display
  - [x] Country coverage stats:
    - [x] Number of countries visited
    - [x] Percentage of world coverage
    - [x] Country flags display
- [x] Create `web/src/components/maps/heatmap.tsx`:
  - [x] Multi-polyline rendering
  - [x] Configurable color
  - [x] Sport-type coloring
- [x] Add filters (backend support):
  - [x] Sport type
  - [x] Date range
  - [x] Commute filter

### 7.4 Segment Map

- [x] Create `web/src/components/maps/segment-map.tsx`:
  - [x] Segment route visualization
  - [x] Start/end markers
  - [x] Gradient indicators for climb segments
  - [x] Fit bounds to segment

### 7.5 Virtual World Maps

- [x] Research Zwift map tiles:
  - [x] Watopia map tiles
  - [x] London map tiles
  - [x] New York map tiles
  - [x] Other Zwift worlds
- [x] Research Rouvy map integration
- [x] Research MyWhoosh map integration
- [x] Create virtual world tile layers:
  - [x] Custom tile layer per virtual world
  - [x] Coordinate transformation if needed
- [x] Detect virtual activities:
  - [x] Check for VirtualRide/VirtualRun sport types
  - [x] Identify world from activity metadata
- [x] Show appropriate map:
  - [x] Auto-switch to virtual world map for virtual activities
  - [x] Fall back to real-world map if detection fails

---

## Phase 8: Advanced Features ✓

**Goal:** Segments, gear, maintenance, calendar, photos, challenges.

### 8.1 Segments

- [x] Create `schema/migrations/005_segments.sql`:
  - [x] Segments table (id, name, distance, avg_grade, max_grade, climb_category, start_latlng, end_latlng, starred, polyline)
  - [x] Segment efforts table (segment_id, activity_id, elapsed_time, moving_time, start_date, pr_rank, avg_watts, avg_hr)
  - [x] Indexes for efficient queries
- [x] Create `internal/storage/segments.go`:
  - [x] Insert/update segment
  - [x] Insert segment effort
  - [x] Get segment by ID
  - [x] List segments with filters
  - [x] Get segment efforts
  - [x] Get best effort per segment
- [x] Create `internal/strava/segments.go`:
  - [x] Parse segment data from activity response
  - [x] Fetch segment details API call
- [x] Update importer to import segment efforts:
  - [x] Extract segment_efforts from activity detail
  - [x] Store segments and efforts
- [x] Implement segment API endpoints:
  - [x] `GET /api/v1/segments` - list with filters
  - [x] `GET /api/v1/segments/:id` - single segment
  - [x] `GET /api/v1/segments/:id/efforts` - efforts history
- [x] Create `web/src/pages/segments.tsx`:
  - [x] Searchable segment list
  - [x] Filter by sport type (radio buttons)
  - [x] Filter by country (radio buttons with flags)
  - [x] Filter by starred/favorite segments
  - [x] Filter by KOM status
  - [x] Sortable columns (name, distance, gradient, ride count, last effort)
  - [x] Display: name, distance, max gradient, climb category, times completed, last effort, best time
- [x] Create segment detail modal:
  - [x] Segment map visualization
  - [x] All personal efforts history
  - [x] Best time progression chart
  - [x] Link to Strava segment page

### 8.2 Gear

- [x] Create `schema/migrations/004_gear.sql`
- [x] Create `internal/storage/gear.go`
- [x] Implement gear API endpoints
- [x] Create `web/src/pages/gear.tsx`:
  - [x] Gear list with stats
  - [x] Active vs retired toggle
  - [x] Per-gear metrics
- [x] Create custom gear functionality:
  - [x] User-defined gear not trackable in Strava
  - [x] Examples: skateboards, kayaks, snowboards
  - [x] Linked via hashtags in activity titles (#gear-name)
  - [x] Same statistics as Strava gear
  - [x] CRUD UI for custom gear
  - [x] `GET /api/v1/gear/custom` endpoint
  - [x] `POST /api/v1/gear/custom` endpoint
  - [x] `PUT /api/v1/gear/custom/:id` endpoint
  - [x] `DELETE /api/v1/gear/custom/:id` endpoint
- [x] Create gear statistics charts:
  - [x] Distance per month per gear (stacked bar)
  - [x] Distance over time per gear (cumulative line)
  - [x] Moving time per gear (donut)
- [x] Add purchase price tracking:
  - [x] Purchase price field
  - [x] Relative cost per hour calculation
  - [x] Relative cost per activity calculation

### 8.3 Gear Maintenance

- [x] Create `schema/migrations/006_maintenance.sql`:
  - [x] Components table (id, gear_id, name, image_url, created_at)
  - [x] Maintenance rules table (component_id, type, threshold_value)
  - [x] Maintenance log table (component_id, activity_id, completed_at)
- [x] Create `internal/storage/maintenance.go`:
  - [x] CRUD for components
  - [x] CRUD for maintenance rules
  - [x] Log maintenance completion
  - [x] Calculate component wear status
- [x] Implement maintenance API endpoints:
  - [x] `GET /api/v1/gear/:id/components` - list components
  - [x] `POST /api/v1/gear/:id/components` - add component
  - [x] `PUT /api/v1/components/:id` - update component
  - [x] `DELETE /api/v1/components/:id` - remove component
  - [x] `POST /api/v1/components/:id/maintenance` - log maintenance
  - [x] `GET /api/v1/maintenance/due` - components needing maintenance
- [x] Create maintenance UI:
  - [x] Define components (chain, cassette, brake pads, etc.)
  - [x] Attach components to specific gear
  - [x] Set maintenance intervals:
    - [x] Distance-based (every X km/mi)
    - [x] Time-based (every X hours used)
    - [x] Calendar-based (every X days)
  - [x] Visual progress indicators (linear progress bars)
  - [x] Component images support
  - [x] Maintenance history accordion
- [x] Hashtag-based maintenance tracking:
  - [x] Detect maintenance hashtags in activity titles
  - [x] Auto-reset component counters
  - [x] Log maintenance date and trigger activity

### 8.4 Calendar

- [x] Create `web/src/pages/calendar.tsx`
- [x] Create `web/src/components/calendar/month-view.tsx`:
  - [x] CSS Grid calendar
  - [x] Activities on dates
  - [x] Color by sport type
- [x] Create month navigation
- [x] Create `GET /api/v1/dashboard/calendar/activities` endpoint
- [x] Create monthly summary stats:
  - [x] Total distance
  - [x] Total elevation
  - [x] Total time
  - [x] Number of challenges completed
  - [x] Total calories
  - [x] Number of workouts
  - [x] Header display in calendar page
- [x] Create day detail modal:
  - [x] Show all activities for clicked date
  - [x] Activity name, distance, time, elevation
  - [x] Link to activity detail
  - [x] Daily totals

### 8.5 Photos

- [x] Create `schema/migrations/007_photos.sql`:
  - [x] Photos table (id, activity_id, url, thumbnail_url, caption, location)
- [x] Create `internal/strava/photos.go`:
  - [x] Fetch activity photos API call
  - [x] Parse photo response
- [x] Update importer to import photos:
  - [x] Fetch photos for each activity
  - [x] Store photo URLs and metadata
- [x] Implement photo API endpoints:
  - [x] `GET /api/v1/photos` - list with filters
  - [x] `GET /api/v1/activities/:id/photos` - activity photos
- [x] Create `web/src/pages/photos.tsx`:
  - [x] Masonry-style photo wall (flexbox-based)
  - [x] Lazy loading for performance
  - [x] Lightbox slideshow mode (LightGallery integration)
  - [x] Filter by sport type (multi-select)
  - [x] Filter by country
  - [x] Photo count display
  - [x] Hover overlay with activity name and date
  - [x] Click-through to source activity

### 8.6 Challenges

- [x] Create `schema/migrations/008_challenges.sql`:
  - [x] Challenges table (id, name, slug, badge_url, completion_date, month)
- [x] Create challenge scraping:
  - [x] Scrape visible challenges from public profile
  - [x] Parse trophy case HTML export (manual import option)
- [x] Implement challenge API endpoints:
  - [x] `GET /api/v1/challenges` - list with filters
  - [x] `POST /api/v1/challenges/import` - trigger import
- [x] Create `web/src/pages/challenges.tsx`:
  - [x] Grouped by completion month
  - [x] Challenge badge images
  - [x] Challenge names
  - [x] Links to Strava challenge pages
  - [x] Challenge count per month
- [x] Create challenge consistency widget:
  - [x] Monthly challenge completion tracking
  - [x] Configurable challenges:
    - [x] Distance goals (total or single activity)
    - [x] Elevation goals
    - [x] Moving time goals
    - [x] Number of activities
    - [x] Calories burned
  - [x] Per sport type filtering
- [x] Create most recent challenges dashboard widget:
  - [x] Recently completed Strava challenges
  - [x] Challenge badges with completion dates
  - [x] "View all" link

---

## Phase 9: Analytics ✓

**Goal:** Eddington, best efforts, training load, rewind.

### 9.1 Eddington Number

- [x] Implement `GET /api/v1/stats/eddington`:
  - [x] Calculate Eddington number
  - [x] Days distribution
  - [x] Next steps tracking
- [x] Create `web/src/pages/eddington.tsx`:
  - [x] Current number display
  - [x] Top distance days table
  - [x] "Next goals" table
  - [x] Sport type filter (All/Rides/Runs)
- [x] Create `eddington.tsx` dashboard widget:
  - [x] Compact current Eddington numbers display
  - [x] Configurable sport type groups
  - [x] "View details" link
- [x] Create Eddington history chart:
  - [x] Eddington progression over time
  - [x] Show when each new number was achieved
  - [x] Line chart visualization
- [x] Configurable Eddington definitions:
  - [x] Multiple Eddington definitions in settings
  - [x] Sport type groupings per definition
  - [x] NavBar visibility toggle
  - [x] Dashboard widget visibility toggle

### 9.2 Best Efforts

- [x] Create `internal/strava/best_efforts.go`:
  - [x] Parse best_efforts from activity response
  - [x] Calculate best efforts for standard distances
- [x] Create `schema/migrations/009_best_efforts.sql`:
  - [x] Best efforts table (activity_id, distance_type, elapsed_time, start_index, end_index)
- [x] Create `internal/storage/best_efforts.go`:
  - [x] Store best efforts per activity
  - [x] Query personal records per distance
  - [x] Query all efforts per distance
- [x] Update importer to import best efforts:
  - [x] Extract best_efforts from activity detail
  - [x] Store in database
- [x] Implement best efforts API endpoints:
  - [x] `GET /api/v1/stats/best-efforts` - all PRs
  - [x] `GET /api/v1/stats/best-efforts/:distance` - efforts for distance
- [x] Create `web/src/pages/best-efforts.tsx`:
  - [x] Running distances: 400m, 1/2 mile, 1km, 1 mile, 2 mile, 5km, 10km, 15km, 10 mile, 20km, Half Marathon, 30km, Marathon, 50km, 100km
  - [x] Grouped by activity type tabs (Run, Ride, etc.)
  - [x] Chart showing records over time
  - [x] Table with date, time, and activity link
- [x] Create distance efforts modal:
  - [x] Click-through from best efforts table
  - [x] All efforts for that distance
  - [x] Progression chart over time

### 9.3 Training Load

- [x] Create training load calculations:
  - [x] Training Stress Score (TSS) per activity
  - [x] Based on FTP and power data (cycling)
  - [x] Based on pace and heart rate (running)
  - [x] Intensity Factor (IF) calculation
  - [x] Normalized Power calculation
- [x] Create `internal/storage/training_load.go`:
  - [x] Store daily TSS
  - [x] Calculate Chronic Training Load (CTL/Fitness)
  - [x] Calculate Acute Training Load (ATL/Fatigue)
  - [x] Calculate Training Stress Balance (TSB/Form)
- [x] Implement training load API endpoints:
  - [x] `GET /api/v1/stats/training-load` - fitness/fatigue curves
  - [x] `GET /api/v1/stats/training-load/daily` - daily TSS values
- [x] Create `web/src/pages/training-load.tsx`:
  - [x] Fitness (CTL) line over time
  - [x] Fatigue (ATL) line over time
  - [x] Form (TSB) line over time
  - [x] Daily training stress bars
  - [x] Date range selector
  - [x] Tooltip with values on hover
- [x] Create training load dashboard widget:
  - [x] Current fitness/fatigue/form values
  - [x] Mini chart preview
  - [x] "View details" link

### 9.4 Strava Rewind (Year in Review)

- [x] Implement `GET /api/v1/stats/rewind`:
  - [x] Year parameter
  - [x] All rewind metrics
- [x] Create `web/src/pages/rewind.tsx`:
  - [x] Year selector
  - [x] Compare years feature (side-by-side)
- [x] Implement rewind metrics:
  - [x] Total activities count by month (bar chart)
  - [x] Distance per month (bar chart)
  - [x] Elevation per month (bar chart)
  - [x] Moving time per sport type (pie/donut)
  - [x] Active days vs rest days (pie chart)
  - [x] Activity start times by hour (line chart)
  - [x] Personal records per month (line chart)
  - [x] Activity locations world map (ECharts effectScatter)
  - [x] Streaks (consecutive active days, rest days)
  - [x] Carbon saved (estimated CO2 reduction from cycling commutes)
  - [x] Socials (kudos received total)
  - [x] Biggest activities:
    - [x] Longest distance
    - [x] Most elevation
    - [x] Longest duration
  - [x] Random photo from the year
- [x] Create comparison view:
  - [x] Compare any two years
  - [x] Compare year vs. all-time
  - [x] Side-by-side metric display
  - [x] Percentage change indicators

### 9.5 Settings Robustness

- [x] Ensure `/api/v1/settings` always returns normalized settings (including `virtual_world_tile_layers`)
- [x] Ensure `PUT /api/v1/settings` returns the normalized stored settings (not the raw payload)
- [x] Ensure `/api/v1/zones/hr` is backward-compatible when the `hr_zone_definitions` table is missing
- [x] Guard Settings UI against missing/empty settings maps

---

## Phase 10: Feature Parity & Beyond

**Goal:** Match and exceed the reference Statistics for Strava implementation with missing features and unique enhancements.

### Gap Analysis Summary

Compared against reference project (statistics-for-strava PHP implementation):

| Category | Missing Features | Priority |
|----------|-----------------|----------|
| Dashboard Widgets | 6 widgets | High |
| Pages | 2 pages | Medium |
| Heatmap Features | 4 features | Medium |
| UX Enhancements | 5 features | Medium |
| Beyond Reference | 8 features | High |

---

### 10.1 Missing Dashboard Widgets

#### 10.1.1 Yearly Stats Widget ✓
Year-over-year comparison with delta indicators showing improvement/regression.

- [x] Create `web/src/components/dashboard/yearly-stats.tsx`:
  - [x] Multi-year comparison table
  - [x] Columns: year, distance, elevation, time, activities
  - [x] Delta arrows (↑/↓) with color coding (green/red)
  - [x] Percentage change display
- [x] Uses existing `GET /api/v1/dashboard/yearly` endpoint
- [x] Add widget to dashboard registry

#### 10.1.2 Distance Breakdown Widget ✓
Activity statistics grouped by distance categories/zones.

- [x] Create `web/src/components/dashboard/distance-breakdown.tsx`:
  - [x] Distance zone categories (0-5km, 5-10km, 10-20km, 20-50km, 50-100km, 100km+)
  - [x] Per-zone stats: count, total distance, avg distance
  - [x] Bar chart and table view toggle
- [x] Uses client-side aggregation from activities data
- [x] Add widget to dashboard registry

#### 10.1.3 Zwift Stats Widget ✓
Virtual cycling platform statistics per world.

- [x] Create `web/src/components/dashboard/zwift-stats.tsx`:
  - [x] Table with Zwift world rows (Watopia, London, New York, France, Paris, etc.)
  - [x] Columns: workouts, distance, elevation, time
  - [x] World detection from location_city/country metadata
  - [x] Support for Zwift, Rouvy, MyWhoosh platforms
- [x] Uses client-side aggregation from activities data
- [x] Add widget to dashboard registry

#### 10.1.4 Recent Challenges Widget ✓
Horizontal scrolling display of recently completed challenges.

- [x] Update `web/src/components/dashboard/recent-challenges.tsx`:
  - [x] Horizontal scroll container with overflow
  - [x] Challenge badge images
  - [x] Scroll buttons (ChevronLeft/ChevronRight)
  - [x] Challenge name on hover
  - [x] Link to Strava challenge page
  - [x] "View all" link to challenges page
- [x] Already in dashboard registry

#### 10.1.5 Intro Text Widget ✓
Custom introductory content widget for dashboard personalization.

- [x] Create `web/src/components/dashboard/intro-text.tsx`:
  - [x] Dynamic welcome message based on activity count
  - [x] Motivational quote rotation
  - [x] Lifetime stats grid (activities, distance, elevation, time)
  - [x] Achievement badges based on milestones
- [x] Add widget to dashboard registry

#### 10.1.6 Challenge Consistency Grid Widget ✓
Visual grid showing monthly goal achievement with checkmarks.

- [x] Create `web/src/components/dashboard/challenge-consistency-grid.tsx`:
  - [x] Challenge names as rows
  - [x] Months as columns (vertical text orientation)
  - [x] Green checkmark (✓) for goal reached
  - [x] Red X (✗) for goal missed
  - [x] Color-coded cells (green/red background)
  - [x] Year selector
- [x] Uses existing goals data
- [x] Add widget to dashboard registry

---

### 10.2 Missing Pages

#### 10.2.1 Monthly Stats Page ✓
Dedicated page with accordion-style monthly breakdown tables.

- [x] Create `web/src/pages/monthly-stats.tsx`:
  - [x] Accordion rows by month (expandable/collapsible)
  - [x] Per-month columns: workouts, distance, elevation, moving time, calories
  - [x] Per-activity-type breakdown when expanded
  - [x] Year selector
  - [x] Export to CSV option
  - [x] Yearly summary cards
- [x] Uses existing `GET /api/v1/dashboard/monthly` endpoint
- [x] Add route and navigation link

#### 10.2.2 Badge Display Page
SVG badge generation for profile embedding.

- [ ] Create `web/src/pages/badges.tsx`:
  - [ ] Three tabs: User Badge, PB Badges, Virtual Badges
  - [ ] User Badge:
    - [ ] Total stats summary badge
    - [ ] Customizable colors
    - [ ] SVG preview
    - [ ] Embed code generator (Markdown, HTML)
  - [ ] PB Badges:
    - [ ] Personal best badges per sport type
    - [ ] Distance/time achievements
  - [ ] Virtual Badges:
    - [ ] Zwift level/achievements
    - [ ] Virtual world stats
- [ ] Create `GET /api/v1/badges/user` endpoint
- [ ] Create `GET /api/v1/badges/pb` endpoint
- [ ] Create `GET /api/v1/badges/virtual` endpoint
- [ ] SVG generation in Go (`internal/badges/`)
- [ ] Add route and navigation link

---

### 10.3 Heatmap Enhancements

#### 10.3.1 Click-to-Explore Nearby Routes ✓
Discover activities near a clicked point on the map.

- [x] Add click handler to heatmap:
  - [x] On map click, find activities within 500m radius (configurable)
  - [x] Show popup with matching activities (up to 10)
  - [x] Activity name, date, distance, sport type
  - [x] Link to activity detail
- [x] Uses client-side Haversine distance calculation
- [x] CircleMarker visualization at click point

#### 10.3.2 Country View Switching with FlyTo ✓
Quick navigation to activities by country.

- [x] Add country selector to CountryPanel:
  - [x] List countries with activity counts
  - [x] Country flags
  - [x] Visual selection state
- [x] Implement flyTo animation:
  - [x] Calculate bounds for selected country's activities
  - [x] Animate map to fit bounds using flyToBounds
  - [x] Smooth transition with padding

#### 10.3.3 Workout Type Filter ✓
Filter heatmap activities by workout type.

- [x] Workout type filter already exists in heatmap sidebar (lines 246-257)
- [x] Includes Race, Workout types
- [x] Backend already supports workout_type filter

#### 10.3.4 Activity Route Preview on Hover
Show activity details when hovering over a route.

- [ ] Add hover handlers to polylines:
  - [ ] Highlight hovered route (increase opacity/width)
  - [ ] Show tooltip with activity name, date, stats
- [ ] Implement efficient hover detection for overlapping routes

---

### 10.4 UX Enhancements

#### 10.4.1 Connected Charts (Cross-Filtering)
Synchronized selection across multiple dashboard charts.

- [ ] Create ECharts connection manager:
  - [ ] Register chart instances by ID
  - [ ] Broadcast selection events
  - [ ] Highlight corresponding data across charts
- [ ] Implement in monthly/weekly/yearly charts:
  - [ ] Click on month → highlight in all charts
  - [ ] Show activities for selected period
- [ ] Create connected chart wrapper component

#### 10.4.2 Chart Detail Modals ✓
Deep-dive popups when clicking chart elements.

- [x] Create `web/src/components/charts/chart-detail-modal.tsx`:
  - [x] Reusable modal component with Dialog
  - [x] Title, subtitle, stats grid
  - [x] Custom content slot
  - [x] `useChartDetailModal` hook for state management
- [x] Ready for integration with chart click handlers

#### 10.4.3 Multi-Field Search in Activities ✓
Advanced search across multiple activity fields.

- [x] Create `web/src/components/activities/activity-filters.tsx`:
  - [x] Debounced search input with icon
  - [x] Sport type quick filters (All, Ride, VirtualRide, Run, VirtualRun, Walk)
  - [x] Advanced filters panel (toggle)
  - [x] Date range filters (from/to)
  - [x] Commute/Trainer toggles
  - [x] Clear filters button
- [x] Search syntax help tooltip
- [x] Integrated with activities page

#### 10.4.4 Accordion Tables ✓
Collapsible row groups for dense data display.

- [x] Create `web/src/components/ui/accordion-table.tsx`:
  - [x] Expandable row groups with generic types
  - [x] Summary row with totals
  - [x] Expand/collapse all button
  - [x] Animated transitions (ChevronRight rotation)
  - [x] Configurable column definitions
- [x] Applied to Monthly stats page

#### 10.4.5 Clustered Table Rendering ✓
Performance optimization for large tables.

- [x] Implement virtual scrolling with @tanstack/react-virtual:
  - [x] Render only visible rows with ROW_HEIGHT constant
  - [x] Smooth scroll handling with overscan
  - [x] Alternating row colors for readability
- [x] Create `VirtualizedActivitiesTable` component:
  - [x] Flexbox-based layout for proper column sizing
  - [x] Fixed header with sticky positioning
  - [x] Footer showing activity count
- [x] Apply to Activities page:
  - [x] Toggle between paginated and virtual "View All" mode
  - [x] Fetches up to 2000 activities in View All mode

---

### 10.5 Beyond Reference - Unique Features

These features go beyond the reference implementation to make Stata superior.

#### 10.5.1 AI-Powered Activity Insights ✓
Automatic insights generated from activity data.

- [x] Create `web/src/components/dashboard/activity-insights.tsx`:
  - [x] Detect patterns (streaks, milestones, trends)
  - [x] Compare to historical averages
  - [x] Generate natural language insights
- [x] Insight types implemented:
  - [x] Streak detection (consecutive active days)
  - [x] Monthly distance trend analysis (+/-20%)
  - [x] Big effort recognition (2x average distance)
  - [x] Time-of-day patterns (morning/evening athlete)
  - [x] Location variety (countries visited)
  - [x] Power trend analysis
  - [x] Distance/activity milestones
- [x] Add widget to dashboard registry

#### 10.5.2 Weather Overlay for Activities
Show weather conditions during activities.

- [ ] Integrate weather API (Open-Meteo):
  - [ ] Historical weather lookup by date/location
  - [ ] Temperature, wind, precipitation
- [ ] Create `web/src/components/activities/weather-badge.tsx`:
  - [ ] Weather icon
  - [ ] Temperature range
  - [ ] Wind speed/direction
  - [ ] Conditions (sunny, cloudy, rainy)
- [ ] Add weather to activity detail page
- [ ] Store weather data in database (cache)

#### 10.5.3 Route Similarity Finder
Find similar routes to a given activity.

- [ ] Create route similarity algorithm:
  - [ ] Compare polyline shapes
  - [ ] Distance matching
  - [ ] Start/end location proximity
- [ ] Create `GET /api/v1/activities/:id/similar` endpoint
- [ ] Add "Similar Routes" section to activity detail
- [ ] Show comparison stats (faster/slower times)

#### 10.5.4 Training Plan Suggestions
Basic training recommendations based on history.

- [ ] Create `internal/training/planner.go`:
  - [ ] Analyze training load trends
  - [ ] Suggest recovery days
  - [ ] Recommend target distances
- [ ] Create `GET /api/v1/training/suggestions` endpoint
- [ ] Create training suggestions widget:
  - [ ] "Consider a rest day tomorrow"
  - [ ] "Good time for a long ride"
  - [ ] "Your fitness is peaking"

#### 10.5.5 Social Comparison (Anonymous)
Compare your stats to aggregated anonymous data.

- [ ] Create opt-in data sharing:
  - [ ] Anonymous aggregate submission
  - [ ] No PII included
- [ ] Create comparison display:
  - [ ] "Your average distance is higher than 75% of cyclists"
  - [ ] Age group comparisons
  - [ ] Regional comparisons

#### 10.5.6 Goal Recommendations ✓
Smart goal suggestions based on historical data.

- [x] Analyze past performance:
  - [x] Calculate averages from last 6 months
  - [x] Compare against best month achievements
  - [x] Weekly/monthly targets derived from patterns
- [x] Create `web/src/components/dashboard/goal-recommendations.tsx`:
  - [x] Data-driven recommendations based on history
  - [x] One-click goal creation ("Set Goal" button)
  - [x] Difficulty indicators (easy/moderate/stretch)
  - [x] Distance, elevation, and time metrics
- [x] Add widget to dashboard registry

#### 10.5.7 Strava Comments & Kudos Integration
Display social interactions from Strava.

- [ ] Import kudos count per activity
- [ ] Import comments (text only)
- [ ] Create `web/src/components/activities/social-stats.tsx`:
  - [ ] Kudos count with heart icon
  - [ ] Comments list
  - [ ] "Most kudos'd activities" widget
- [ ] Add to activity detail page

#### 10.5.8 Export & Backup Features ✓
Comprehensive data export capabilities.

- [x] Create `GET /api/v1/export/csv` (CSV format)
- [x] Create `GET /api/v1/export/json` (JSON format)
- [x] Create `GET /api/v1/export/stats` (export statistics)
- [x] Create `web/src/pages/export.tsx`:
  - [x] Export format selection (CSV/JSON)
  - [x] Sport type filter
  - [x] Date range filter
  - [x] Download button with preview
- [x] Add route and navigation link

---

### 10.6 Implementation Priority

**Sprint 1 - Dashboard Gaps (High Priority):** ✓
1. ✓ Yearly Stats Widget
2. ✓ Distance Breakdown Widget
3. ✓ Recent Challenges Widget (horizontal scroll)
4. ✓ Challenge Consistency Grid Widget
5. ✓ Zwift Stats Widget
6. ✓ Intro Text Widget

**Sprint 2 - Pages & Heatmap:** ✓
1. ✓ Monthly Stats Page
2. ✓ Click-to-Explore on Heatmap
3. ✓ Country View Switching with FlyTo
4. ✓ Workout Type Filter (already existed)
5. ✓ Export Page

**Sprint 3 - UX Polish:** ✓
1. ✓ Chart Detail Modals
2. ✓ Multi-Field Search
3. ✓ Accordion Tables
4. ✓ Clustered Table Rendering

**Sprint 4 - Beyond Reference:** (Partial)
1. ✓ AI-Powered Insights
2. [ ] Weather Overlay
3. ✓ Goal Recommendations
4. [ ] Route Similarity Finder

**Sprint 5 - Advanced Features:**
1. [ ] Badge Display Page
2. [ ] Connected Charts
3. [ ] Training Plan Suggestions
4. [ ] Social Comparison

---

## Phase 11: WASM Mode

**Goal:** Browser-only version with SQLite-WASM.

### 11.1 SQLite-WASM Setup

- [ ] Install `sql.js` or `@aspect-build/sqlite3-wasm`
- [ ] Create `web/src/lib/db/wasm-client.ts`:
  - [ ] Initialize SQLite-WASM
  - [ ] OPFS persistence setup
  - [ ] Implement DataSource interface
- [ ] Create schema migration for browser:
  - [ ] Embed SQL as strings
  - [ ] Version tracking in SQLite

### 11.2 Strava Direct Integration

- [ ] Create `web/src/lib/strava/client.ts`:
  - [ ] Direct Strava API calls from browser
  - [ ] OAuth implicit flow handling
- [ ] Create OAuth callback page for browser mode
- [ ] Handle token storage (localStorage)
- [ ] Rate limit tracking in browser

### 11.3 Browser Importer

- [ ] Create `web/src/lib/importer/`:
  - [ ] Activity importer (browser version)
  - [ ] Progress tracking
  - [ ] Rate limit awareness
- [ ] Create import UI:
  - [ ] Start import button
  - [ ] Progress display
  - [ ] Estimated time remaining
  - [ ] Pause/resume

### 11.4 Mode Detection & Switching

- [ ] Create build configuration for WASM mode
- [ ] Create `web/src/lib/mode.ts`:
  - [ ] Detect current mode
  - [ ] Feature flags per mode
- [ ] Update DataSource factory
- [ ] Hide server-only features in WASM mode:
  - [ ] Webhooks
  - [ ] Scheduled imports
  - [ ] Challenge scraping

### 11.5 Static Hosting Build

- [ ] Create separate Vite config for WASM build
- [ ] Configure for static hosting (GitHub Pages, etc.)
- [ ] Create deployment documentation

### 11.6 Token Exchange Service (Optional)

- [ ] Create Cloudflare Worker for token exchange
- [ ] Or document user-provided serverless option
- [ ] Handle client_secret securely

---

## Phase 12: Polish

**Goal:** PWA, internationalization, settings, badges, final polish.

### 12.1 PWA Support

- [ ] Create `web/public/manifest.json`
- [ ] Create service worker for offline support
- [ ] Add app icons (multiple sizes)
- [ ] Configure installable prompt
- [ ] Test PWA installation

### 12.2 Unit System

- [ ] Create unit conversion utilities
- [ ] Add unit system selector
- [ ] Apply unit formatting throughout app
- [ ] Persist preference

### 12.3 Settings Page

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

### 12.4 SVG Badges

- [ ] Create badge generation in Go:
  - [ ] Strava stats badge
  - [ ] PB badges per sport
  - [ ] Zwift badge
- [ ] Create badge API endpoints
- [ ] Create badge preview in settings
- [ ] Create embed code generator

### 12.5 Webhooks

- [ ] Implement Strava webhook subscription:
  - [ ] `GET /api/v1/webhooks/strava` - validation
  - [ ] `POST /api/v1/webhooks/strava` - receive events
- [ ] Handle activity create/update/delete events
- [ ] Trigger incremental import on webhook

### 12.6 Scheduler

- [ ] Create `internal/scheduler/scheduler.go`
- [ ] Configurable scheduled jobs:
  - [ ] Periodic full sync
  - [ ] Maintenance check notifications
  - [ ] App update check
- [ ] Create scheduler configuration in settings

### 12.7 Notifications

- [ ] Integrate Shoutrrr for notifications
- [ ] Create notification settings UI
- [ ] Send notifications for:
  - [ ] Import completion
  - [ ] Maintenance due
  - [ ] New features (optional)

### 12.8 Final Polish

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

_This TODO list should be updated as implementation progresses. Check off items as they are completed and add new items as scope is refined._
