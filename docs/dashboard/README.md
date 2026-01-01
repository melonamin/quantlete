# Dashboard Overview

The Quantlete dashboard provides a comprehensive view of your Strava activities and training data.

## Pages

### Home Dashboard

The main dashboard (`/`) shows:

- Summary statistics (total distance, time, elevation)
- Recent activities
- Weekly/monthly comparisons
- Quick stats widgets

### [Activities](activities.md)

Browse and search all your activities:

- Activity list with filtering
- Activity detail view with maps and charts
- Calendar view
- Heatmap of all locations

### [Analytics](analytics.md)

Training analysis tools:

- Training load (fitness/fatigue/form)
- Power analysis and zones
- Best efforts and PRs
- Eddington number
- Monthly statistics

### [Gear](gear.md)

Equipment tracking:

- Bikes and shoes
- Usage statistics
- Maintenance logging
- Component tracking

### [Settings](settings.md)

Configuration and data management:

- Display preferences (units, date format)
- Goals and targets
- Data export
- Athlete profile

## Navigation

The sidebar provides quick access to all pages. On mobile, use the hamburger menu.

## Data Freshness

Activity data syncs automatically when:

- You open the dashboard (background sync)
- Webhooks are configured (real-time)
- Scheduled sync runs (every hour)

For manual sync, use the CLI:

```bash
quantlete import
```

## Demo Mode

If running with demo data (`quantlete demo`), the dashboard shows realistic sample data. A banner indicates demo mode is active.
