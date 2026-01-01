# Statistics for Strava - Remaining Work

This document tracks remaining implementation work. Completed phases (0-9) have been archived.

---

## Overview

| Phase | Focus                   | Status |
| ----- | ----------------------- | ------ |
| 0-9   | Core Implementation     | ✓      |
| 11    | WASM Mode               | ✓      |
| 12    | Polish                  | ~85%   |

---

## Phase 12: Polish

### 12.7 Notifications

- [ ] Integrate Shoutrrr for notifications
- [ ] Create notification settings UI
- [ ] Send notifications for:
  - [ ] Import completion
  - [ ] Maintenance due
  - [ ] New features (optional)

### 12.10 Final Polish

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

## Archived Phases (Completed)

<details>
<summary>Phase 0: Project Setup ✓</summary>

Repository structure, tooling, CI/CD pipeline, Go/React project setup, linting, testing setup.
</details>

<details>
<summary>Phase 1: Core Backend ✓</summary>

Configuration system, HTTP server with chi, Strava OAuth flow, Strava API client, basic activity fetching.
</details>

<details>
<summary>Phase 2: Database & Storage ✓</summary>

SQLite connection, schema migrations, activity repository, importer service, athlete/token storage, pagination utilities.
</details>

<details>
<summary>Phase 3: Frontend Foundation ✓</summary>

React app shell, layout components, shadcn/ui setup, data layer, state management, TanStack Query, utility functions.
</details>

<details>
<summary>Phase 4: Activities Feature ✓</summary>

Activities API endpoints, activities list page with filters, activity detail page, import functionality, activity streams.
</details>

<details>
<summary>Phase 5: Dashboard ✓</summary>

Dashboard infrastructure, widget system with drag-and-drop, core widgets (stats, recent activities, weekly stats, sport breakdown, training goals).
</details>

<details>
<summary>Phase 6: Charts & Visualizations ✓</summary>

ECharts infrastructure, line/bar/pie/donut charts, calendar heatmap, activity stream charts, dashboard chart widgets.
</details>

<details>
<summary>Phase 7: Maps & Heatmap ✓</summary>

Leaflet map infrastructure, activity maps with elevation coloring, heatmap page with filters, segment maps, virtual world maps (Zwift/Rouvy/MyWhoosh).
</details>

<details>
<summary>Phase 8: Advanced Features ✓</summary>

Segments with efforts, gear with custom gear support, gear maintenance with hashtag tracking, calendar with day details, photos with lightbox, challenges with consistency widget.
</details>

<details>
<summary>Phase 9: Analytics ✓</summary>

Eddington number with history, best efforts with progression, training load (TSS/CTL/ATL/TSB), Strava Rewind year-in-review with comparison.
</details>

<details>
<summary>Phase 10: Feature Parity (Completed Items)</summary>

Dashboard widgets: Yearly Stats, Distance Breakdown, Zwift Stats, Recent Challenges, Intro Text, Challenge Consistency Grid, Kudos Leaders.

Pages: Monthly Stats, Badge Display, Export.

Heatmap: Click-to-Explore, Country FlyTo, Workout Type Filter.

UX: Chart Detail Modals, Multi-Field Search, Accordion Tables, Virtualized Tables.

Beyond Reference: AI-Powered Insights, Weather Overlay, Goal Recommendations, Export & Backup.
</details>

<details>
<summary>Phase 11: WASM Mode ✓</summary>

SQLite-WASM with sql.js, OPFS/IndexedDB persistence, Strava direct integration with OAuth, browser importer with progress tracking, mode detection & switching, DataProvider abstraction.

WASM Provider implements: Activities, Dashboard stats, Eddington, Training load, Best efforts, Import, Weather, Segments, Gear, Maintenance, Photos, Challenges.

ETA estimation (ported from Go), pause/resume with state persistence, power stats with automatic computation during stream import and backfill for existing data.
</details>

<details>
<summary>Phase 12: Polish (Completed Items)</summary>

PWA support with service worker, unit system (metric/imperial), settings page with zones/FTP/weight, SVG badges with themes, Strava webhooks, scheduler with sync settings, security hardening (CSRF/CSP/headers), demo mode with realistic data generation.
</details>
