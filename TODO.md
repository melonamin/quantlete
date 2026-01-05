# Statistics for Strava - Remaining Work

This document tracks remaining implementation work. Completed phases (0-11) have been archived.

---

## Overview

| Phase | Focus                          | Status |
| ----- | ------------------------------ | ------ |
| 0-11  | Core Implementation            | ✓      |
| 12    | Polish                         | ~90%   |
| 13    | Advanced Analytics             | Planned |
| 14    | Explorer Tiles                 | Planned |

---

## Phase 12: Polish (Remaining)

### 12.7 Notifications ✓
- [x] Integrate Shoutrrr for notifications
- [x] Create notification settings UI
- [x] Send notifications for import completion, maintenance due

### 12.10 Final Polish
- [ ] Accessibility audit (ARIA labels, keyboard navigation)
- [ ] Performance audit (Lighthouse, bundle size, lazy loading)
- [ ] Error handling review
- [ ] Mobile responsiveness testing

---

## Phase 13: Advanced Analytics & Gamification

### 13.1 Gear Price UI + ROI ⭐ Quick Win

Add UI to input gear purchase prices and display cost/km.

**Database**: None (columns exist: `purchase_price`, `purchase_currency`)

**Backend**:
- [ ] Add `UpdateGearPrice` query to `schema/queries/gear.sql`
- [ ] Add `UpdatePrice()` to `internal/storage/gear.go`
- [ ] Add `UpdateGearPrice` service method to `internal/services/gear.go`

**Frontend**:
- [ ] Add `useUpdateGearPrice` mutation hook
- [ ] Create `PriceEditModal` component in `web/src/pages/gear.tsx`
- [ ] Add cost/km calculation and display to `GearCard`
- [ ] Add "Edit Price" button for Strava gear
- [ ] Create `GearROISummary` component (total investment, avg cost/km)

### 13.2 Smart Coach Insights ⭐ Quick Win

Rule-based text alerts from training data (TSB, streaks, CTL trends).

**Backend**:
- [ ] Create `internal/services/insights.go` with rule engine
- [ ] Implement insight rules:
  - High Fatigue (TSB < -20)
  - CTL Drop (>10% in 7 days)
  - Streak Detection (consecutive active days)
  - Well Recovered (TSB > 10)
  - Overtraining Risk (ATL > CTL × 1.5)

**Frontend**:
- [ ] Add `useInsights` hook
- [ ] Create `SmartCoach` dashboard widget
- [ ] Add to widget grid (default visible)

### 13.3 Zone Trend Analysis ⭐ Quick Win

Stacked area chart showing % training time in HR zones over 52 weeks.

**Database**:
- [ ] Create migration `004_zone_distribution.sql`
- [ ] Add `activity_zone_distribution` table

**Backend**:
- [ ] Add `GetWeeklyZoneDistribution` query to `schema/queries/zones.sql`
- [ ] Add `SaveActivityZoneDistribution()` to storage
- [ ] Add `GetZoneTrend` service method
- [ ] Modify importer to compute zone distribution from HR streams

**Frontend**:
- [ ] Create `ZoneTrendChart` component (stacked area)
- [ ] Create `ZoneTrend` dashboard widget
- [ ] Add to widget grid (hidden by default)

### 13.4 Aerobic Decoupling (Cardiac Drift)

Calculate Pw:HR or Pa:HR ratio comparing first vs second half of steady-state efforts.

**Algorithm** (add to `algorithms/go/algorithms.go`):
- [ ] `AerobicDecoupling(heartrate, output []float64) float64`
- [ ] Steady-state detection (filter warmup/cooldown, min 30min)

**Database**:
- [ ] Add `activity_analytics` table (activity_id, decoupling_pct, efficiency_index)

**Backend**:
- [ ] Create `internal/storage/analytics.go` repository
- [ ] Create `internal/services/analytics.go` service
- [ ] Add API endpoint `GET /api/v1/activities/{id}/decoupling`

**Frontend**:
- [ ] Create `DecouplingBadge` component (green <5%, yellow 5-10%, red >10%)
- [ ] Add to activity detail page

### 13.5 Race Time Predictor

Use best efforts to predict race times via Riegel's formula.

**Algorithm** (add to `algorithms/go/algorithms.go`):
- [ ] `PredictRaceTime(baseTime, baseDist, targetDist, fatigueFactor) int`
- [ ] `PredictAllRaces()` for standard distances (5k → marathon)

**Backend**:
- [ ] Create `internal/services/race_predictor.go`
- [ ] Add API endpoint `GET /api/v1/stats/race-predictions`

**Frontend**:
- [ ] Create `RacePredictionsWidget` dashboard component
- [ ] Create `/race-predictions` page with full predictions table
- [ ] Show confidence based on how close base effort is to target

### 13.6 Efficiency Index (EF)

Track NP/AvgHR (cycling) or NGP/AvgHR (running) over time.

**Algorithm** (add to `algorithms/go/algorithms.go`):
- [ ] `EfficiencyIndex(normalizedOutput, avgHR float64) float64`
- [ ] `NormalizedGradedPace(paces, grades []float64) float64` (running)

**Backend**:
- [ ] Add EF calculation to `internal/services/analytics.go`
- [ ] Store in `activity_analytics` table
- [ ] Add API endpoints for activity EF and EF history

**Frontend**:
- [ ] Create `EfficiencyHistoryChart` component
- [ ] Create `EfficiencyBadge` for activity detail
- [ ] Create `/efficiency` page with trend chart

---

## Phase 14: Explorer Tiles (Tile Hunting)

Full-featured tile tracking with counting, max cluster detection, and map overlay.

### 14.1 Database Schema
- [ ] Create migration `005_explorer_tiles.sql`
- [ ] Add `explorer_tiles` table (athlete_id, tile_x, tile_y, first_visit_date, first_activity_id)
- [ ] Add `explorer_clusters` cache table (max_cluster_size, cluster_x, cluster_y)

### 14.2 Tile Algorithms
Create `internal/geo/tiles.go`:
- [ ] `LatLngToTile(lat, lng) (x, y int)` - Slippy map tiles at zoom 14
- [ ] `TileToBounds(x, y) (minLat, minLng, maxLat, maxLng)`
- [ ] `ExtractTilesFromLatLng(latlng [][2]float64) []TileCoord`
- [ ] `FindMaxCluster(tiles) (size, x, y int)` - DP algorithm
- [ ] `FindBorderTiles(tiles, cluster) []TileCoord`

### 14.3 Storage Layer
Create `internal/storage/explorer_tiles.go`:
- [ ] Add queries to `schema/queries/explorer.sql`
- [ ] `UpsertTiles()`, `GetAllTiles()`, `GetTilesByDateRange()`
- [ ] `CountTiles()`, `CountTilesByYear()`, `CountTilesByMonth()`
- [ ] `GetCluster()`, `UpdateCluster()`

### 14.4 Import Integration
- [ ] Add tile extraction to activity import in `internal/importer/importer.go`
- [ ] Add `SaveExplorerTiles()` to ImportStorage interface
- [ ] Handle activity deletion (cascade delete tiles)

### 14.5 Service Layer
Create `internal/services/explorer.go`:
- [ ] `GetExplorerData()` - stats + tiles + border tiles
- [ ] `RecalculateTiles()` - for historical backfill

### 14.6 Frontend - TypeScript Utilities
Create `web/src/lib/geo/tiles.ts`:
- [ ] Port tile calculation functions for WASM mode parity
- [ ] `latLngToTile()`, `tileToBounds()`, `extractTilesFromLatLng()`

### 14.7 Frontend - Explorer Page
- [ ] Add `useExplorerData` hook
- [ ] Create `ExplorerMap` component with Leaflet Rectangle overlays
- [ ] Create `/explorer` page with stats bar and map
- [ ] Display: total tiles, this year/month, max cluster (N×N)
- [ ] Color coding: cluster tiles (green), other visited (orange), border (gray dashed)
- [ ] Click tile → popup with first visit date and link to activity

### 14.8 Backfill & Polish
- [ ] Create recalculation endpoint for existing activities
- [ ] Add explorer stats to dashboard (optional widget)
- [ ] Performance: use Canvas renderer for large tile sets

---

## Testing Milestones

### Unit Tests
- [ ] Go: Algorithm functions (decoupling, race predictor, EF, tiles)
- [ ] Go: Storage repositories (analytics, explorer)
- [ ] React: Tile calculation utilities

### Integration Tests
- [ ] Go: Explorer tile import flow
- [ ] Go: Analytics calculation endpoints

### E2E Tests
- [ ] Explorer page map interaction
- [ ] Gear price editing flow
- [ ] Race predictions page

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

## Critical Files Reference

**Algorithms**: `algorithms/go/algorithms.go`
**Services pattern**: `internal/services/stats.go`
**API handlers**: `internal/api/handlers/stats.go`
**WASM bindings**: `cmd/wasm/algorithms.go`
**Dashboard widgets**: `web/src/components/dashboard/`
**Map components**: `web/src/components/maps/`
**Data hooks**: `web/src/lib/data/hooks.ts`
**Importer**: `internal/importer/importer.go`

---

## Archived Phases (Completed)

<details>
<summary>Phase 0-11: Core Implementation ✓</summary>

- Phase 0: Project Setup
- Phase 1: Core Backend
- Phase 2: Database & Storage
- Phase 3: Frontend Foundation
- Phase 4: Activities Feature
- Phase 5: Dashboard
- Phase 6: Charts & Visualizations
- Phase 7: Maps & Heatmap
- Phase 8: Advanced Features (Segments, Gear, Calendar, Photos, Challenges)
- Phase 9: Analytics (Eddington, Best Efforts, Training Load, Rewind)
- Phase 10: Feature Parity
- Phase 11: WASM Mode
</details>

<details>
<summary>Phase 12: Polish (Completed Items)</summary>

PWA support with service worker, unit system (metric/imperial), settings page with zones/FTP/weight, SVG badges with themes, Strava webhooks, scheduler with sync settings, security hardening (CSRF/CSP/headers), demo mode with realistic data generation, Shoutrrr notifications with settings UI.
</details>
