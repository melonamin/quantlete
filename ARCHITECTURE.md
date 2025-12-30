# Statistics for Strava - Technical Architecture

This document describes the technical architecture for implementing the Statistics for Strava application as specified in `SPECIFICATION.md`.

---

## 1. System Overview

### 1.1 Design Goals

| Goal | Approach |
|------|----------|
| **Single binary distribution** | Go with embedded React assets via `go:embed` |
| **CLI + Web modes** | Cobra CLI with `serve` and utility commands |
| **Rich UI** | React 18 + shadcn/ui + Tailwind CSS |
| **Analytics performance** | SQLite for reliable local storage |
| **Browser-only option** | SQLite-WASM with OPFS persistence |
| **Self-hosted simplicity** | SQLite-like deployment (single file database) |

### 1.2 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         DEPLOYMENT MODES                                │
├─────────────────────────────────┬───────────────────────────────────────┤
│       Self-Hosted Binary        │         Browser-Only (WASM)           │
├─────────────────────────────────┼───────────────────────────────────────┤
│  ┌───────────────────────────┐  │  ┌─────────────────────────────────┐  │
│  │      Go Binary            │  │  │       Static Hosting            │  │
│  │  ┌─────────┐ ┌─────────┐  │  │  │   (CDN / GitHub Pages)          │  │
│  │  │   CLI   │ │  HTTP   │  │  │  └─────────────────────────────────┘  │
│  │  │ (cobra) │ │ Server  │  │  │                 │                     │
│  │  └────┬────┘ └────┬────┘  │  │                 ▼                     │
│  │       │           │       │  │  ┌─────────────────────────────────┐  │
│  │       └─────┬─────┘       │  │  │         React SPA               │  │
│  │             │             │  │  │  ┌───────────┐ ┌─────────────┐  │  │
│  │             ▼             │  │  │  │  UI Layer │ │ SQLite-WASM │  │  │
│  │  ┌─────────────────────┐  │  │  │  └─────┬─────┘ └──────┬──────┘  │  │
│  │  │   Embedded React    │  │  │  │        │              │         │  │
│  │  │   (go:embed dist/)  │  │  │  │        └──────┬───────┘         │  │
│  │  └─────────────────────┘  │  │  │               │                 │  │
│  │             │             │  │  │               ▼                 │  │
│  │             ▼             │  │  │  ┌─────────────────────────┐    │  │
│  │  ┌─────────────────────┐  │  │  │  │   OPFS (Browser FS)     │    │  │
│  │  │       SQLite        │  │  │  │  │   quantlete.db              │    │  │
│  │  │  (pure Go driver)   │  │  │  │  └─────────────────────────┘    │  │
│  │  └─────────────────────┘  │  │  └─────────────────────────────────┘  │
│  │             │             │  │                 │                     │
│  │             ▼             │  │                 │                     │
│  │  ┌─────────────────────┐  │  │                 │                     │
│  │  │     quantlete.db        │  │  │                 │                     │
│  │  │   (local file)      │  │  │                 │                     │
│  │  └─────────────────────┘  │  │                 │                     │
│  └───────────────────────────┘  │                 │                     │
│              │                  │                 │                     │
└──────────────┼──────────────────┴─────────────────┼─────────────────────┘
               │                                    │
               ▼                                    ▼
        ┌─────────────────────────────────────────────────────┐
        │                   Strava API                        │
        │  OAuth 2.0 │ Activities │ Streams │ Webhooks        │
        └─────────────────────────────────────────────────────┘
```

---

## 2. Technology Stack

### 2.1 Backend (Go)

| Component | Technology | Version | Rationale |
|-----------|------------|---------|-----------|
| Language | Go | 1.23+ | Single binary, excellent concurrency |
| HTTP Router | chi | v5 | Stdlib-compatible, middleware support |
| CLI Framework | cobra | v1.8+ | Industry standard for Go CLIs |
| Configuration | viper | v1.18+ | Multi-source config, cobra integration |
| Database | modernc.org/sqlite | latest | Pure Go SQLite driver (no CGO) |
| OAuth | golang.org/x/oauth2 | latest | Standard OAuth2 implementation |
| Scheduler | robfig/cron | v3 | Cron expression support |
| Logging | slog | stdlib | Structured logging (Go 1.21+) |
| Embed | go:embed | stdlib | Embed React build in binary |
| Testing | testify | v1.9+ | Assertions and mocking |

### 2.2 Frontend (React)

| Component | Technology | Version | Rationale |
|-----------|------------|---------|-----------|
| Framework | React | 18.3+ | Component model, ecosystem |
| Build Tool | Vite | 5.x | Fast builds, good DX |
| Language | TypeScript | 5.x | Type safety |
| Styling | Tailwind CSS | 3.4+ | Utility-first, spec requirement |
| Components | shadcn/ui | latest | Owned code, Radix primitives |
| Charts | ECharts | 5.x | Spec requirement, powerful |
| Maps | react-leaflet | 4.x | Spec requirement |
| Tables | TanStack Table | 8.x | Headless, virtualization support |
| Virtualization | @tanstack/react-virtual | 3.x | Large list performance |
| State | Zustand | 4.x | Simple, performant |
| Data Fetching | TanStack Query | 5.x | Caching, background refresh |
| Router | TanStack Router | 1.x | Type-safe routing |
| Forms | React Hook Form | 7.x | Performance, validation |
| SQLite WASM | @aspect-build/sqlite3-wasm | latest | Browser-only mode |

### 2.3 Build & Tooling

| Component | Technology | Rationale |
|-----------|------------|-----------|
| Task Runner | just | User preference, simple |
| Go Releases | GoReleaser | Multi-platform CGO builds |
| Node Package | yarn | User preference (from js.md) |
| Linting (Go) | golangci-lint | Comprehensive linting |
| Linting (TS) | ESLint + Prettier | Standard tooling |
| Testing (E2E) | Playwright | User preference |
| CI/CD | GitHub Actions | Standard, GoReleaser integration |

### 2.4 Database: SQLite

**Why SQLite:**

SQLite was chosen for its simplicity and reliability:

- **Pure Go driver** - `modernc.org/sqlite` requires no CGO, simplifying cross-compilation
- **Battle-tested** - Most widely deployed database engine
- **Single file** - Simple backup and portability
- **Zero configuration** - No server setup required
- **WASM support** - Available for browser-only mode via sql.js or @aspect-build/sqlite3-wasm
- **Adequate performance** - For a single-user analytics dashboard, SQLite provides sufficient performance

---

## 3. Project Structure

```
quantlete/
├── .github/
│   └── workflows/
│       ├── ci.yml              # Test and lint on PR
│       └── release.yml         # GoReleaser on tag
│
├── .goreleaser.yaml            # Multi-platform build config
├── justfile                    # Task runner commands
├── go.mod
├── go.sum
├── CLAUDE.md                   # Project-specific AI instructions
│
├── cmd/
│   └── quantlete/
│       ├── main.go             # Entry point
│       ├── root.go             # Root cobra command
│       ├── serve.go            # `quantlete serve` - start web server
│       ├── import.go           # `quantlete import` - manual import
│       ├── export.go           # `quantlete export` - export data
│       └── version.go          # `quantlete version`
│
├── internal/
│   ├── api/
│   │   ├── router.go           # Chi router setup
│   │   ├── middleware.go       # Auth, logging, CORS
│   │   ├── handlers/
│   │   │   ├── activities.go
│   │   │   ├── dashboard.go
│   │   │   ├── segments.go
│   │   │   ├── gear.go
│   │   │   ├── auth.go         # OAuth flow handlers
│   │   │   └── webhooks.go     # Strava webhook receiver
│   │   └── responses.go        # Standard response helpers
│   │
│   ├── config/
│   │   ├── config.go           # Configuration struct
│   │   └── loader.go           # Viper loading logic
│   │
│   ├── storage/
│   │   ├── db.go               # SQLite connection management
│   │   ├── migrations.go       # Schema migrations
│   │   ├── activities.go       # Activity repository
│   │   ├── segments.go         # Segment repository
│   │   ├── gear.go             # Gear repository
│   │   ├── stats.go            # Aggregation queries
│   │   └── queries/            # Embedded SQL files
│   │       ├── activities.sql
│   │       ├── dashboard.sql
│   │       └── ...
│   │
│   ├── strava/
│   │   ├── client.go           # Strava API client
│   │   ├── oauth.go            # OAuth2 flow
│   │   ├── activities.go       # Activity fetching
│   │   ├── streams.go          # Stream data fetching
│   │   ├── webhooks.go         # Webhook handling
│   │   ├── ratelimit.go        # Rate limit tracking
│   │   └── types.go            # Strava API types
│   │
│   ├── importer/
│   │   ├── importer.go         # Import orchestration
│   │   ├── activities.go       # Activity import logic
│   │   ├── streams.go          # Stream import logic
│   │   ├── weather.go          # Open-Meteo enrichment
│   │   └── challenges.go       # Challenge scraping
│   │
│   ├── scheduler/
│   │   ├── scheduler.go        # Cron job management
│   │   └── jobs.go             # Individual job definitions
│   │
│   └── domain/
│       ├── activity.go         # Activity domain model
│       ├── segment.go          # Segment domain model
│       ├── gear.go             # Gear domain model
│       └── athlete.go          # Athlete domain model
│
├── embed.go                    # //go:embed web/dist/*
│
├── web/                        # React application
│   ├── package.json
│   ├── yarn.lock
│   ├── tsconfig.json
│   ├── vite.config.ts
│   ├── tailwind.config.ts
│   ├── postcss.config.js
│   ├── components.json         # shadcn/ui config
│   ├── index.html
│   │
│   ├── public/
│   │   ├── manifest.json       # PWA manifest
│   │   └── icons/
│   │
│   ├── src/
│   │   ├── main.tsx            # React entry point
│   │   ├── app.tsx             # App component with router
│   │   ├── vite-env.d.ts
│   │   │
│   │   ├── components/
│   │   │   ├── ui/             # shadcn/ui components
│   │   │   │   ├── button.tsx
│   │   │   │   ├── card.tsx
│   │   │   │   ├── dialog.tsx
│   │   │   │   └── ...
│   │   │   │
│   │   │   ├── layout/
│   │   │   │   ├── header.tsx
│   │   │   │   ├── sidebar.tsx
│   │   │   │   ├── nav.tsx
│   │   │   │   └── breadcrumb.tsx
│   │   │   │
│   │   │   ├── dashboard/
│   │   │   │   ├── widget-grid.tsx
│   │   │   │   ├── widgets/
│   │   │   │   │   ├── recent-activities.tsx
│   │   │   │   │   ├── training-goals.tsx
│   │   │   │   │   ├── weekly-stats.tsx
│   │   │   │   │   ├── heart-rate-zones.tsx
│   │   │   │   │   ├── activity-grid.tsx
│   │   │   │   │   ├── monthly-stats.tsx
│   │   │   │   │   ├── gear-stats.tsx
│   │   │   │   │   ├── eddington.tsx
│   │   │   │   │   └── ...
│   │   │   │   └── widget-config.tsx
│   │   │   │
│   │   │   ├── charts/
│   │   │   │   ├── echarts-wrapper.tsx
│   │   │   │   ├── line-chart.tsx
│   │   │   │   ├── bar-chart.tsx
│   │   │   │   ├── pie-chart.tsx
│   │   │   │   ├── calendar-heatmap.tsx
│   │   │   │   └── ...
│   │   │   │
│   │   │   ├── maps/
│   │   │   │   ├── activity-map.tsx
│   │   │   │   ├── heatmap.tsx
│   │   │   │   ├── segment-map.tsx
│   │   │   │   └── tile-layers.tsx
│   │   │   │
│   │   │   ├── activities/
│   │   │   │   ├── activity-table.tsx
│   │   │   │   ├── activity-filters.tsx
│   │   │   │   ├── activity-card.tsx
│   │   │   │   └── activity-detail.tsx
│   │   │   │
│   │   │   ├── segments/
│   │   │   ├── gear/
│   │   │   ├── photos/
│   │   │   └── calendar/
│   │   │
│   │   ├── pages/
│   │   │   ├── dashboard.tsx
│   │   │   ├── activities.tsx
│   │   │   ├── activity-detail.tsx
│   │   │   ├── segments.tsx
│   │   │   ├── heatmap.tsx
│   │   │   ├── calendar.tsx
│   │   │   ├── best-efforts.tsx
│   │   │   ├── eddington.tsx
│   │   │   ├── gear.tsx
│   │   │   ├── photos.tsx
│   │   │   ├── challenges.tsx
│   │   │   ├── rewind.tsx
│   │   │   └── settings.tsx
│   │   │
│   │   ├── hooks/
│   │   │   ├── use-activities.ts
│   │   │   ├── use-dashboard-stats.ts
│   │   │   ├── use-segments.ts
│   │   │   ├── use-filters.ts
│   │   │   └── use-theme.ts
│   │   │
│   │   ├── lib/
│   │   │   ├── db/
│   │   │   │   ├── index.ts          # DataSource factory
│   │   │   │   ├── types.ts          # Shared interfaces
│   │   │   │   ├── api-client.ts     # REST API implementation
│   │   │   │   └── wasm-client.ts    # SQLite-WASM implementation
│   │   │   │
│   │   │   ├── strava/
│   │   │   │   ├── client.ts         # Direct Strava API (WASM mode)
│   │   │   │   └── oauth.ts          # OAuth helpers
│   │   │   │
│   │   │   ├── utils/
│   │   │   │   ├── units.ts          # Metric/imperial conversion
│   │   │   │   ├── dates.ts          # Date formatting
│   │   │   │   ├── sport-types.ts    # Sport type utilities
│   │   │   │   └── cn.ts             # tailwind-merge helper
│   │   │   │
│   │   │   └── constants/
│   │   │       ├── sport-types.ts
│   │   │       ├── colors.ts
│   │   │       └── hr-zones.ts
│   │   │
│   │   ├── stores/
│   │   │   ├── settings.ts           # User preferences
│   │   │   ├── filters.ts            # Global filter state
│   │   │   └── dashboard.ts          # Dashboard widget config
│   │   │
│   │   └── types/
│   │       ├── activity.ts
│   │       ├── segment.ts
│   │       ├── gear.ts
│   │       └── api.ts
│   │
│   └── tests/
│       ├── unit/
│       └── integration/              # Playwright tests
│
├── schema/
│   └── migrations/
│       ├── 001_initial.sql
│       ├── 002_activities.sql
│       ├── 003_streams.sql
│       ├── 004_segments.sql
│       ├── 005_gear.sql
│       └── 006_challenges.sql
│
└── docs/
    ├── SPECIFICATION.md              # Feature specification
    ├── ARCHITECTURE.md               # This document
    └── TODO.md                       # Implementation plan
```

---

## 4. Database Schema

### 4.1 Core Tables

```sql
-- Athletes (single user, but structured for potential multi-user)
CREATE TABLE athletes (
    id BIGINT PRIMARY KEY,              -- Strava athlete ID
    username VARCHAR,
    firstname VARCHAR,
    lastname VARCHAR,
    profile_url VARCHAR,
    access_token VARCHAR,
    refresh_token VARCHAR,
    token_expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Activities
CREATE TABLE activities (
    id BIGINT PRIMARY KEY,              -- Strava activity ID
    athlete_id BIGINT REFERENCES athletes(id),
    sport_type VARCHAR NOT NULL,
    name VARCHAR NOT NULL,
    description VARCHAR,
    start_date TIMESTAMP NOT NULL,
    start_date_local TIMESTAMP NOT NULL,
    timezone VARCHAR,

    -- Metrics
    distance DOUBLE,                    -- meters
    moving_time INTEGER,                -- seconds
    elapsed_time INTEGER,               -- seconds
    total_elevation_gain DOUBLE,        -- meters
    calories INTEGER,

    -- Speed
    average_speed DOUBLE,               -- m/s
    max_speed DOUBLE,                   -- m/s

    -- Heart rate
    average_heartrate DOUBLE,
    max_heartrate INTEGER,

    -- Power
    average_watts DOUBLE,
    max_watts INTEGER,
    weighted_average_watts DOUBLE,      -- Normalized power

    -- Cadence
    average_cadence DOUBLE,
    max_cadence INTEGER,

    -- Best power efforts (calculated from streams)
    best_power_5s INTEGER,
    best_power_10s INTEGER,
    best_power_30s INTEGER,
    best_power_60s INTEGER,
    best_power_300s INTEGER,
    best_power_480s INTEGER,
    best_power_1200s INTEGER,
    best_power_3600s INTEGER,

    -- Location
    start_lat DOUBLE,
    start_lng DOUBLE,
    end_lat DOUBLE,
    end_lng DOUBLE,
    polyline VARCHAR,                   -- Encoded polyline
    country VARCHAR,

    -- Metadata
    gear_id VARCHAR,
    device_name VARCHAR,
    is_commute BOOLEAN DEFAULT FALSE,
    workout_type INTEGER,
    kudos_count INTEGER DEFAULT 0,
    photo_count INTEGER DEFAULT 0,

    -- Weather (from Open-Meteo)
    weather_temp DOUBLE,
    weather_condition VARCHAR,
    weather_humidity INTEGER,
    weather_wind_speed DOUBLE,

    -- Virtual world detection
    world_type VARCHAR DEFAULT 'real',  -- real, zwift, rouvy, mywhoosh

    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    -- Indexing
    INDEX idx_activities_athlete (athlete_id),
    INDEX idx_activities_date (start_date_local),
    INDEX idx_activities_sport (sport_type),
    INDEX idx_activities_gear (gear_id)
);

-- Activity streams (time-series data)
CREATE TABLE activity_streams (
    activity_id BIGINT REFERENCES activities(id),
    time_offset INTEGER NOT NULL,       -- seconds from start
    distance DOUBLE,                    -- meters
    latitude DOUBLE,
    longitude DOUBLE,
    altitude DOUBLE,                    -- meters
    heartrate INTEGER,
    cadence INTEGER,
    watts INTEGER,
    temp DOUBLE,
    velocity DOUBLE,                    -- m/s
    grade DOUBLE,                       -- percent
    moving BOOLEAN,

    PRIMARY KEY (activity_id, time_offset)
);

-- Segments
CREATE TABLE segments (
    id BIGINT PRIMARY KEY,              -- Strava segment ID
    name VARCHAR NOT NULL,
    sport_type VARCHAR NOT NULL,
    distance DOUBLE,                    -- meters
    average_grade DOUBLE,
    maximum_grade DOUBLE,
    elevation_high DOUBLE,
    elevation_low DOUBLE,
    climb_category INTEGER,
    start_lat DOUBLE,
    start_lng DOUBLE,
    end_lat DOUBLE,
    end_lng DOUBLE,
    polyline VARCHAR,
    country VARCHAR,
    is_starred BOOLEAN DEFAULT FALSE,
    is_kom BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Segment efforts
CREATE TABLE segment_efforts (
    id BIGINT PRIMARY KEY,              -- Strava effort ID
    segment_id BIGINT REFERENCES segments(id),
    activity_id BIGINT REFERENCES activities(id),
    athlete_id BIGINT REFERENCES athletes(id),
    start_date TIMESTAMP NOT NULL,
    elapsed_time INTEGER,               -- seconds
    moving_time INTEGER,
    distance DOUBLE,
    average_watts DOUBLE,
    average_heartrate DOUBLE,
    max_heartrate INTEGER,
    pr_rank INTEGER,                    -- 1, 2, 3 or NULL

    INDEX idx_efforts_segment (segment_id),
    INDEX idx_efforts_activity (activity_id)
);

-- Gear
CREATE TABLE gear (
    id VARCHAR PRIMARY KEY,             -- Strava gear ID (b12345)
    athlete_id BIGINT REFERENCES athletes(id),
    name VARCHAR NOT NULL,
    gear_type VARCHAR,                  -- bike, shoes, etc.
    brand VARCHAR,
    model VARCHAR,
    distance DOUBLE DEFAULT 0,          -- meters (from Strava)
    is_primary BOOLEAN DEFAULT FALSE,
    is_retired BOOLEAN DEFAULT FALSE,

    -- Custom fields
    purchase_date DATE,
    purchase_price DOUBLE,
    notes VARCHAR,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Custom gear (not in Strava, tracked via hashtags)
CREATE TABLE custom_gear (
    id VARCHAR PRIMARY KEY,             -- Generated UUID
    athlete_id BIGINT REFERENCES athletes(id),
    name VARCHAR NOT NULL,
    hashtag VARCHAR NOT NULL UNIQUE,    -- e.g., #skateboard
    gear_type VARCHAR,
    is_retired BOOLEAN DEFAULT FALSE,
    purchase_date DATE,
    purchase_price DOUBLE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Gear components (for maintenance tracking)
CREATE TABLE gear_components (
    id VARCHAR PRIMARY KEY,
    gear_id VARCHAR,                    -- References gear or custom_gear
    name VARCHAR NOT NULL,
    component_type VARCHAR,

    -- Maintenance thresholds
    distance_threshold DOUBLE,          -- meters
    time_threshold INTEGER,             -- hours
    calendar_threshold INTEGER,         -- days

    -- Current state
    current_distance DOUBLE DEFAULT 0,
    current_time INTEGER DEFAULT 0,
    installed_date DATE,
    last_maintenance_date DATE,

    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Maintenance log
CREATE TABLE maintenance_log (
    id VARCHAR PRIMARY KEY,
    component_id VARCHAR REFERENCES gear_components(id),
    activity_id BIGINT,                 -- Activity that triggered (if any)
    maintenance_date DATE NOT NULL,
    notes VARCHAR,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Challenges
CREATE TABLE challenges (
    id VARCHAR PRIMARY KEY,
    athlete_id BIGINT REFERENCES athletes(id),
    name VARCHAR NOT NULL,
    slug VARCHAR,
    badge_url VARCHAR,
    completed_at DATE,
    month INTEGER,                      -- For grouping
    year INTEGER,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Activity photos
CREATE TABLE activity_photos (
    id VARCHAR PRIMARY KEY,
    activity_id BIGINT REFERENCES activities(id),
    url VARCHAR NOT NULL,
    thumbnail_url VARCHAR,
    caption VARCHAR,
    location_lat DOUBLE,
    location_lng DOUBLE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Athlete settings/history
CREATE TABLE athlete_weight_history (
    athlete_id BIGINT REFERENCES athletes(id),
    date DATE NOT NULL,
    weight DOUBLE NOT NULL,             -- kg
    PRIMARY KEY (athlete_id, date)
);

CREATE TABLE athlete_ftp_history (
    athlete_id BIGINT REFERENCES athletes(id),
    date DATE NOT NULL,
    ftp INTEGER NOT NULL,               -- watts
    sport_type VARCHAR DEFAULT 'cycling',
    PRIMARY KEY (athlete_id, date, sport_type)
);

CREATE TABLE athlete_hr_zones (
    athlete_id BIGINT REFERENCES athletes(id),
    effective_date DATE NOT NULL,
    sport_type VARCHAR,
    zone_mode VARCHAR DEFAULT 'relative', -- relative or absolute
    zone1_min INTEGER,
    zone1_max INTEGER,
    zone2_max INTEGER,
    zone3_max INTEGER,
    zone4_max INTEGER,
    zone5_max INTEGER,
    PRIMARY KEY (athlete_id, effective_date, sport_type)
);

-- Best efforts (running distances)
CREATE TABLE best_efforts (
    id VARCHAR PRIMARY KEY,
    activity_id BIGINT REFERENCES activities(id),
    athlete_id BIGINT REFERENCES athletes(id),
    name VARCHAR NOT NULL,              -- "400m", "5K", "Marathon", etc.
    distance DOUBLE NOT NULL,           -- meters
    elapsed_time INTEGER NOT NULL,      -- seconds
    moving_time INTEGER,
    start_date TIMESTAMP NOT NULL,
    pr_rank INTEGER,

    INDEX idx_efforts_athlete_name (athlete_id, name)
);

-- Dashboard widget configuration
CREATE TABLE dashboard_config (
    athlete_id BIGINT REFERENCES athletes(id),
    widget_order JSON,                  -- Array of widget IDs
    widget_settings JSON,               -- Per-widget configuration
    updated_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (athlete_id)
);
```

### 4.2 Useful Views / Materialized Queries

```sql
-- Monthly aggregates (materialized for dashboard performance)
CREATE VIEW monthly_stats AS
SELECT
    athlete_id,
    sport_type,
    DATE_TRUNC('month', start_date_local) AS month,
    COUNT(*) AS activity_count,
    SUM(distance) AS total_distance,
    SUM(total_elevation_gain) AS total_elevation,
    SUM(moving_time) AS total_moving_time,
    SUM(calories) AS total_calories,
    AVG(average_heartrate) AS avg_heartrate,
    AVG(average_watts) AS avg_watts
FROM activities
GROUP BY athlete_id, sport_type, DATE_TRUNC('month', start_date_local);

-- Weekly aggregates
CREATE VIEW weekly_stats AS
SELECT
    athlete_id,
    sport_type,
    DATE_TRUNC('week', start_date_local) AS week,
    COUNT(*) AS activity_count,
    SUM(distance) AS total_distance,
    SUM(total_elevation_gain) AS total_elevation,
    SUM(moving_time) AS total_moving_time
FROM activities
GROUP BY athlete_id, sport_type, DATE_TRUNC('week', start_date_local);

-- Gear usage stats
CREATE VIEW gear_stats AS
SELECT
    g.id,
    g.name,
    g.gear_type,
    g.is_retired,
    COUNT(a.id) AS activity_count,
    SUM(a.distance) AS total_distance,
    SUM(a.total_elevation_gain) AS total_elevation,
    SUM(a.moving_time) AS total_moving_time,
    SUM(a.calories) AS total_calories,
    AVG(a.average_speed) AS avg_speed
FROM gear g
LEFT JOIN activities a ON a.gear_id = g.id
GROUP BY g.id, g.name, g.gear_type, g.is_retired;
```

---

## 5. API Design

### 5.1 REST Endpoints

```
Base URL: /api/v1

Authentication:
  GET  /auth/strava              # Initiate OAuth flow
  GET  /auth/strava/callback     # OAuth callback
  POST /auth/refresh             # Refresh access token
  GET  /auth/status              # Check auth status

Activities:
  GET  /activities               # List activities (paginated, filterable)
  GET  /activities/:id           # Single activity detail
  GET  /activities/:id/streams   # Activity stream data
  GET  /activities/:id/photos    # Activity photos

Dashboard:
  GET  /dashboard/stats          # Aggregated dashboard data
  GET  /dashboard/weekly         # Current week stats
  GET  /dashboard/monthly        # Monthly comparison data
  GET  /dashboard/config         # Widget configuration
  PUT  /dashboard/config         # Update widget configuration

Segments:
  GET  /segments                 # List segments (filterable)
  GET  /segments/:id             # Segment detail
  GET  /segments/:id/efforts     # Personal efforts on segment

Gear:
  GET  /gear                     # List all gear
  GET  /gear/:id                 # Gear detail with stats
  GET  /gear/:id/activities      # Activities with this gear
  POST /gear/custom              # Create custom gear
  PUT  /gear/custom/:id          # Update custom gear

Maintenance:
  GET  /maintenance/components   # List components
  POST /maintenance/components   # Create component
  PUT  /maintenance/components/:id
  POST /maintenance/log          # Log maintenance action

Stats:
  GET  /stats/eddington          # Eddington numbers
  GET  /stats/best-efforts       # Personal records
  GET  /stats/heatmap            # Route data for heatmap
  GET  /stats/training-load      # Fitness/fatigue data
  GET  /stats/rewind/:year       # Year in review data

Challenges:
  GET  /challenges               # List completed challenges
  POST /challenges/import        # Import challenges

Photos:
  GET  /photos                   # All photos (filterable)

Settings:
  GET  /settings                 # User settings
  PUT  /settings                 # Update settings
  GET  /settings/hr-zones        # Heart rate zone config
  PUT  /settings/hr-zones
  GET  /settings/ftp-history
  POST /settings/ftp-history
  GET  /settings/weight-history
  POST /settings/weight-history

Import:
  POST /import/trigger           # Trigger manual import
  GET  /import/status            # Import progress

Webhooks (Strava):
  GET  /webhooks/strava          # Webhook validation
  POST /webhooks/strava          # Receive webhook events
```

### 5.2 Query Parameters

Activities list supports:
```
GET /activities?
  sport_type=Ride,Run           # Comma-separated sport types
  from=2024-01-01               # Start date
  to=2024-12-31                 # End date
  country=US                    # Country filter
  gear_id=b12345                # Gear filter
  device=Garmin                 # Device filter (partial match)
  is_commute=false              # Commute filter
  workout_type=race             # Workout type
  search=morning                # Name search
  sort=start_date               # Sort field
  order=desc                    # Sort order
  page=1                        # Pagination
  per_page=50                   # Items per page
```

### 5.3 Response Format

```json
{
  "data": { ... },              // Response payload
  "meta": {                     // Pagination info (for lists)
    "page": 1,
    "per_page": 50,
    "total": 1234,
    "total_pages": 25
  }
}
```

Error responses:
```json
{
  "error": {
    "code": "RATE_LIMITED",
    "message": "Strava API rate limit exceeded",
    "details": {
      "retry_after": 900
    }
  }
}
```

---

## 6. Data Layer Abstraction

The frontend uses a unified interface that works in both Go backend mode and WASM browser mode.

### 6.1 Interface Definition

```typescript
// web/src/lib/db/types.ts

export interface DataSource {
  // Activities
  getActivities(filters: ActivityFilters): Promise<PaginatedResult<Activity>>;
  getActivity(id: string): Promise<Activity | null>;
  getActivityStreams(id: string): Promise<ActivityStreams | null>;

  // Dashboard
  getDashboardStats(): Promise<DashboardStats>;
  getWeeklyStats(): Promise<WeeklyStats>;
  getMonthlyStats(years: number[]): Promise<MonthlyStats[]>;
  getActivityGrid(year: number): Promise<ActivityGridData>;

  // Segments
  getSegments(filters: SegmentFilters): Promise<PaginatedResult<Segment>>;
  getSegment(id: string): Promise<Segment | null>;
  getSegmentEfforts(segmentId: string): Promise<SegmentEffort[]>;

  // Stats
  getEddingtonData(sportTypes: string[]): Promise<EddingtonData>;
  getBestEfforts(sportType: string): Promise<BestEffort[]>;
  getHeatmapRoutes(filters: HeatmapFilters): Promise<HeatmapRoute[]>;
  getTrainingLoad(range: DateRange): Promise<TrainingLoadData>;

  // Gear
  getGear(): Promise<Gear[]>;
  getGearStats(gearId: string): Promise<GearStats>;

  // Settings
  getSettings(): Promise<UserSettings>;
  updateSettings(settings: Partial<UserSettings>): Promise<void>;

  // Import (only available in certain modes)
  triggerImport?(): Promise<void>;
  getImportStatus?(): Promise<ImportStatus>;
}

export interface ActivityFilters {
  sportTypes?: string[];
  dateFrom?: Date;
  dateTo?: Date;
  country?: string;
  gearId?: string;
  device?: string;
  isCommute?: boolean;
  workoutType?: string;
  search?: string;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
  page?: number;
  perPage?: number;
}
```

### 6.2 Factory

```typescript
// web/src/lib/db/index.ts

import { DataSource } from './types';
import { ApiDataSource } from './api-client';
import { WasmDataSource } from './wasm-client';

let instance: DataSource | null = null;

export async function getDataSource(): Promise<DataSource> {
  if (instance) return instance;

  const mode = import.meta.env.VITE_DATA_MODE || 'api';

  if (mode === 'wasm') {
    const wasm = new WasmDataSource();
    await wasm.initialize();
    instance = wasm;
  } else {
    instance = new ApiDataSource();
  }

  return instance;
}

export function isWasmMode(): boolean {
  return import.meta.env.VITE_DATA_MODE === 'wasm';
}
```

---

## 7. Build & Deployment

### 7.1 Development Workflow

```just
# justfile

# Run both Go API and React dev server
dev:
    just dev-api &
    just dev-web

dev-api:
    go run ./cmd/quantlete serve --dev --port 8080

dev-web:
    cd web && yarn dev --port 5173

# Build production assets
build: build-web build-go

build-web:
    cd web && yarn build

build-go: build-web
    CGO_ENABLED=1 go build -o bin/quantlete ./cmd/quantlete

# Run tests
test: test-go test-web

test-go:
    go test ./...

test-web:
    cd web && yarn test:unit

test-e2e:
    cd web && yarn playwright test

# Lint
lint: lint-go lint-web

lint-go:
    golangci-lint run

lint-web:
    cd web && yarn lint && yarn typecheck

# Release (multi-platform)
release:
    goreleaser release --clean
```

### 7.2 GoReleaser Configuration

```yaml
# .goreleaser.yaml
version: 2

before:
  hooks:
    - go mod tidy
    - cd web && yarn install && yarn build

builds:
  - id: quantlete
    main: ./cmd/quantlete
    binary: quantlete
    env:
      - CGO_ENABLED=1
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64

    overrides:
      - goos: linux
        goarch: amd64
        env:
          - CC=x86_64-linux-gnu-gcc
      - goos: linux
        goarch: arm64
        env:
          - CC=aarch64-linux-gnu-gcc
      - goos: darwin
        goarch: amd64
        env:
          - CC=o64-clang
      - goos: darwin
        goarch: arm64
        env:
          - CC=oa64-clang
      - goos: windows
        goarch: amd64
        env:
          - CC=x86_64-w64-mingw32-gcc

archives:
  - format: tar.gz
    format_overrides:
      - goos: windows
        format: zip

dockers:
  - image_templates:
      - "ghcr.io/{{ .Env.GITHUB_REPOSITORY_OWNER }}/quantlete:{{ .Version }}"
      - "ghcr.io/{{ .Env.GITHUB_REPOSITORY_OWNER }}/quantlete:latest"
    dockerfile: Dockerfile
    build_flag_templates:
      - "--platform=linux/amd64"
```

### 7.3 Docker Deployment

```dockerfile
# Dockerfile
FROM golang:1.23-bookworm AS builder

RUN apt-get update && apt-get install -y \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o quantlete ./cmd/quantlete

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/quantlete /usr/local/bin/quantlete

VOLUME /data
ENV QUANTLETE_DATA_DIR=/data

EXPOSE 8080
CMD ["quantlete", "serve"]
```

---

## 8. WASM Browser-Only Mode

### 8.1 Architecture

In browser-only mode:
1. React app is served from static hosting (CDN, GitHub Pages)
2. User authenticates directly with Strava (implicit OAuth flow)
3. Browser fetches data from Strava API (CORS is supported)
4. Data is stored in SQLite-WASM with OPFS persistence
5. All processing happens client-side

### 8.2 Strava OAuth in Browser

```typescript
// web/src/lib/strava/oauth.ts

const STRAVA_AUTH_URL = 'https://www.strava.com/oauth/authorize';
const STRAVA_TOKEN_URL = 'https://www.strava.com/oauth/token';

export function initiateStravaAuth() {
  const params = new URLSearchParams({
    client_id: import.meta.env.VITE_STRAVA_CLIENT_ID,
    redirect_uri: window.location.origin + '/auth/callback',
    response_type: 'code',
    scope: 'read,activity:read_all',
  });

  window.location.href = `${STRAVA_AUTH_URL}?${params}`;
}

// Token exchange requires a backend or serverless function
// because client_secret cannot be exposed in browser
export async function exchangeCode(code: string): Promise<TokenResponse> {
  // Option 1: Use a serverless function (Cloudflare Worker, etc.)
  // Option 2: Use Strava's token endpoint with PKCE (if supported)
  const response = await fetch('/api/auth/exchange', {
    method: 'POST',
    body: JSON.stringify({ code }),
  });
  return response.json();
}
```

### 8.3 SQLite-WASM Setup

```typescript
// web/src/lib/db/wasm-client.ts

import initSqlJs, { Database } from 'sql.js';

export class WasmDataSource implements DataSource {
  private db: Database | null = null;

  async initialize(): Promise<void> {
    const SQL = await initSqlJs({
      locateFile: (file) => `https://sql.js.org/dist/${file}`,
    });

    // Try to load existing database from OPFS
    const opfsRoot = await navigator.storage.getDirectory();
    try {
      const fileHandle = await opfsRoot.getFileHandle('quantlete.db');
      const file = await fileHandle.getFile();
      const buffer = await file.arrayBuffer();
      this.db = new SQL.Database(new Uint8Array(buffer));
    } catch {
      // Create new database if none exists
      this.db = new SQL.Database();
    }

    // Run migrations if needed
    await this.runMigrations();
  }

  private async runMigrations(): Promise<void> {
    // Check schema version and apply migrations
    // Migrations are embedded in the JS bundle
  }

  async getActivities(filters: ActivityFilters): Promise<PaginatedResult<Activity>> {
    const { sql, params } = buildActivityQuery(filters);
    const result = this.db!.exec(sql, params);
    return {
      data: result[0]?.values.map(rowToActivity) ?? [],
      meta: { /* pagination */ }
    };
  }

  async persist(): Promise<void> {
    // Save database to OPFS
    const data = this.db!.export();
    const opfsRoot = await navigator.storage.getDirectory();
    const fileHandle = await opfsRoot.getFileHandle('quantlete.db', { create: true });
    const writable = await fileHandle.createWritable();
    await writable.write(data);
    await writable.close();
  }

  // ... implement other methods
}
```

### 8.4 Limitations in WASM Mode

| Feature | Go Binary | WASM Browser |
|---------|-----------|--------------|
| Webhooks | Yes | No (needs server) |
| Scheduled imports | Yes | No (must be tab open) |
| Background sync | Yes | Service Worker possible |
| Initial import speed | Fast | Slower (rate limits + single thread) |
| Large datasets | Good | Limited by browser memory |
| Challenge scraping | Yes | CORS issues |
| Weather enrichment | Yes | Possible (Open-Meteo has CORS) |

---

## 9. Security Considerations

### 9.1 Self-Hosted Mode

- OAuth tokens stored in SQLite database file
- Database file should have restricted permissions (0600)
- No built-in authentication (single-user assumption)
- If exposing to internet, place behind reverse proxy with auth

### 9.2 WASM Mode

- OAuth tokens stored in browser (localStorage or secure cookie)
- No server-side secrets exposure possible
- OPFS data is origin-bound (secure)
- User must trust the hosting domain

### 9.3 API Security

- Rate limit tracking to avoid Strava bans
- No credential storage in client-side code
- Webhook signature verification (Strava signs webhooks)

---

## 10. Performance Considerations

### 10.1 Database

- SQLite is adequate for single-user dashboard aggregations
- Consider materializing frequently-used aggregates
- Stream data (activity_streams) is the largest table - query selectively
- Use OPFS in WASM for better performance than IndexedDB

### 10.2 Frontend

- Virtual scrolling for activity tables (TanStack Virtual)
- Lazy load chart libraries (code splitting)
- Use TanStack Query for data caching and background refresh
- Lazy load map tiles
- Image lazy loading for photos gallery

### 10.3 Import

- Batch Strava API requests to stay within rate limits
- Background import with progress tracking
- Stream data import is optional (configurable)

---

*This architecture document accompanies SPECIFICATION.md and provides the technical foundation for implementation.*
