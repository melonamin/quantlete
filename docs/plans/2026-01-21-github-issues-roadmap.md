# Plan: GitHub Issues Roadmap

Address all 12 open GitHub issues in priority order: bugs first, then dashboard charts, activity analysis, and finally year review features. This plan provides a complete roadmap for shipping fixes and features across the Quantlete codebase.

## Validation Commands

- `just test`
- `just lint`
- `just build`
- `cd web && yarn typecheck`

### Task 1: Fix Training Load zeros (#9)

The Training Load page shows CTL=0, ATL=0, TSB=0 despite activities with power/HR data. This is a critical bug affecting core analytics functionality. The issue likely stems from missing FTP configuration, stream data not being processed, or date range filtering problems in `internal/storage/training_load.go`.

- [x] Add debug logging to `computeAndUpsertActivity()` to trace why activities are skipped
- [x] Verify FTP/threshold values are being read from athlete settings
- [x] Check that `activity_training_load` table is being populated during import
- [x] Add UI warning when FTP/threshold not configured
- [x] Add activity count breakdown showing: total activities, activities with power data, activities with computed TSS
- [x] Write tests for TSS calculation pipeline
- [x] Verify fix by checking Training Load page shows non-zero values

### Task 2: Fix Welcome widget text overlap (#19)

The Welcome widget's "Lifetime Stats" section overlaps with navigation elements and the "Master Athlete" badge is cut off. This is a CSS/layout issue in `web/src/components/dashboard/intro-text.tsx` related to flex container height and overflow handling.

- [x] Reproduce issue and identify exact cause (widget height vs content)
- [x] Check `WidgetWrapper` overflow handling
- [x] Add appropriate `minHeight` to Welcome widget configuration
- [x] Add `overflow-hidden` or `overflow-y-auto` to content container
- [x] Test at different widget sizes (1x1, 2x1, 2x2)
- [x] Verify badge text is fully visible

### Task 3: Fix Sport Distribution legend overlap (#14)

When the Sport Distribution donut chart has 10+ sport types, legend labels overlap and become unreadable. The fix involves switching to a bottom legend for charts with many categories.

- [x] Add `maxInlineLabels` threshold constant (default: 6) to chart constants
- [x] Update `DonutChart` component to conditionally show inline labels vs bottom legend
- [x] Add scrollable legend support for many categories using ECharts `legend.type: 'scroll'`
- [x] Test with various category counts (3, 6, 10, 15 sport types)
- [x] Update Sport Distribution widget to use the improved DonutChart

### Task 4: Improve chart color palettes (#13)

Multiple charts have poor color differentiation. This task establishes a consistent color system that will be used by all subsequent chart work. Defines semantic colors for sports, weekdays, and time-of-day distributions.

- [x] Create comprehensive `sportColors` map in `web/src/components/charts/chart-constants.ts` (20+ sport types)
- [x] Create `weekdayColors` array with weekend/weekday distinction
- [x] Create `daytimeColors` map (Night/Morning/Afternoon/Evening)
- [x] Update `DonutChart` to accept `colorMap` prop for semantic coloring
- [x] Update Sport Distribution widget to use `sportColors`
- [x] Update Weekday widget to use `weekdayColors`
- [x] Update Time of Day widget to use `daytimeColors`
- [x] Add slice grouping for charts with >8 categories (group small values into "Other")

### Task 5: Enhance Activity Calendar (#10)

The Activity Calendar widget needs intensity-based coloring, rolling 365-day view option, and a legend. Currently shows calendar year only with monochrome gradient and no legend to interpret colors.

- [x] Add `CalendarRange` type (`'year' | 'rolling365'`) and toggle UI to widget
- [x] Modify `useCalendarData` hook to support rolling 365 days calculation
- [x] Add intensity metric calculation to calendar SQL query (using `suffer_score` or TSS)
- [x] Update `calendarPalettes` in chart-constants with multi-color intensity gradient (gray/green/amber/orange/red)
- [x] Create `CalendarLegend` component showing color scale with ranges
- [x] Update `ActivityCalendarChart` to use multi-color gradient based on metric
- [x] Add intensity metric tab to widget alongside Count/Distance/Time/Calories
- [x] Run `just generate` after SQL changes

### Task 6: Add Weekly Trends widget (#11)

Add a new dashboard widget showing rolling 12-week trends with sport type filter and metric toggle. The current Monthly Activity widget only shows one bar, which is insufficient for trend analysis.

- [x] Add SQL query `GetWeeklyTrends` in `schema/queries/stats.sql` grouping by ISO week
- [x] Add `WeeklyTrends` method to stats service in `internal/services/stats.go`
- [x] Add API endpoint and handler for weekly trends
- [x] Run `just generate` to create adapters
- [x] Create `WeeklyTrendsChart` component (line + area chart)
- [x] Create `WeeklyTrends` widget with sport type filter UI
- [x] Add metric toggle (distance/time/elevation)
- [x] Register widget in dashboard config with default size 2x2
- [x] Add React Query hook `useWeeklyTrends`

### Task 7: Add Monthly cross-year comparison (#12)

Add a widget showing monthly stats overlaid across multiple years to identify seasonal patterns. Users need to compare January 2024 vs January 2025 vs January 2026 on the same chart.

- [ ] Add SQL query `GetMonthlyComparison` in `schema/queries/stats.sql`
- [ ] Add service method in `internal/services/stats.go`
- [ ] Add API endpoint with `years` and `sport_type` parameters
- [ ] Run `just generate` to create adapters
- [ ] Create `MonthlyComparisonChart` component (multi-series line chart)
- [ ] Create `MonthlyComparison` widget
- [ ] Add year toggle via legend interactions (click to show/hide years)
- [ ] Add sport type filter
- [ ] Add metric toggle (distance/time/elevation)
- [ ] Define `yearColors` palette for up to 10 years

### Task 8: Add Activity detail analysis (#17)

The Activity detail page is missing splits table, HR zone distribution, and pace distribution histogram. These are key analysis features that require stream data processing.

- [ ] Create `getSplits()` function to calculate per-km splits from distance/time streams
- [ ] Create `getHRZoneDistribution()` function using heartrate stream and athlete zones
- [ ] Create `getPaceDistribution()` function to bucket pace data into histogram
- [ ] Add `/api/v1/activities/:id/analysis` endpoint returning splits, zones, distributions
- [ ] Run `just generate` after adding endpoint
- [ ] Create `ActivitySplits` component with pace bars visualization
- [ ] Create `ActivityZoneDistribution` component (horizontal bar chart)
- [ ] Create `ActivityPaceDistribution` histogram component
- [ ] Integrate components into activity detail page as collapsible sections
- [ ] Handle missing stream data gracefully with appropriate empty states

### Task 9: Add distance/duration filters (#15)

The Activities page lacks ability to filter by distance or time duration. Users need filters like "show all runs longer than 10km" or "find activities over 2 hours".

- [ ] Add `min_distance_m`, `max_distance_m`, `min_duration_s`, `max_duration_s` to API types
- [ ] Update SQL query in `schema/queries/activities.sql` with distance/duration WHERE clauses
- [ ] Update `ActivityService.List` input struct with new filter fields and adapter tags
- [ ] Run `just generate` to regenerate adapters
- [ ] Add distance input fields to `ActivityFiltersPanel` (with km/mi unit label)
- [ ] Add duration input fields (hours:minutes format)
- [ ] Add unit conversion based on user preference (km vs mi)
- [ ] Add input validation (min < max, reasonable bounds)
- [ ] Update `useActivities` hook to pass new filter parameters

### Task 10: Group Eddington by sport (#16)

The Eddington number currently calculates across all activity types combined, producing meaningless results for multi-sport athletes. Need sport-specific Eddington calculations.

- [ ] Add `sport_group` parameter to Eddington API endpoint
- [ ] Create `/api/v1/eddington/compare` endpoint returning all sport groups at once
- [ ] Define sport groupings (Running: Run/VirtualRun/TrailRun, Cycling: Ride/VirtualRide/etc.)
- [ ] Run `just generate` after API changes
- [ ] Update `useEddington` hook to accept sport group parameter
- [ ] Update Eddington page to show sport selector prominently
- [ ] Update Eddington dashboard widget to show multiple sports
- [ ] Add sport comparison table showing Eddington per sport group
- [ ] Update Eddington definitions editor to support sport filtering

### Task 11: Improve Rewind/Wrapped view (#18)

The Rewind view has UX issues: months shown as numbers instead of names, unnecessary multi-coloring on bar charts, and missing heatmap view.

- [ ] Fix `monthLabels` to use month names ("Jan", "Feb") instead of numbers ("01", "02")
- [ ] Add `uniformColor` prop to `BarChart` component for single-color bars
- [ ] Update Wrapped page to use single color for monthly bar charts
- [ ] Add year heatmap section using `ActivityCalendarChart` component
- [ ] Fetch calendar data for selected year (may need new hook)
- [ ] Reorganize page sections: Summary → Heatmap → Monthly → PRs → Donuts → Map → Biggest

### Task 12: WASM weather lookup (#3)

Weather data is not available in WASM mode because external API calls cannot work in browser due to CORS. This requires an architectural decision on the approach.

- [ ] Document the three options: proxy through backend, client-friendly API, accept limitation
- [ ] Decide on approach (recommend: accept as permanent WASM limitation with clear UI message)
- [ ] If accepting limitation: add clear "Weather unavailable in browser mode" message in UI
- [ ] If implementing proxy: design API endpoint that proxies Open-Meteo requests
- [ ] Update `getActivityWeather()` in go-provider.ts based on chosen approach
- [ ] Add feature flag or mode detection to show/hide weather section appropriately
