# Statistics for Strava - Feature Specification

This document provides a comprehensive specification of the "Statistics for Strava" application, focused on features and user-facing functionality for potential re-implementation.

---

## 1. Application Overview

**Name:** Statistics for Strava
**Type:** Self-hosted web dashboard (Progressive Web App)
**Purpose:** Comprehensive analytics and statistics dashboard for Strava athletic activities
**Target Users:** Athletes who want deeper insights into their Strava data than the official app provides

### Key Characteristics
- Self-hosted (Docker-based deployment)
- Single-user focused
- Offline-capable (PWA)
- Privacy-focused (data stays on user's infrastructure)
- Multi-language support (en_US, fr_FR, it_IT, nl_BE, de_DE, pt_BR, pt_PT, sv_SE, zh_CN)
- Metric and Imperial unit systems

---

## 2. Data Sources & Integration

### 2.1 Strava API Integration
- OAuth 2.0 authentication flow
- Activity data import (with rate limiting awareness: 15-min and daily limits)
- Webhook support for real-time activity import
- Stream data import (heart rate, power, cadence, temperature, elevation)
- Segment and segment effort import
- Gear data import
- Challenge/trophy import
- Activity photos import

### 2.2 External Data Enrichment
- **Weather data:** Open-Meteo API integration for historical weather at activity location
- **Virtual worlds:** Zwift, Rouvy, MyWhoosh map integration

### 2.3 Strava API Endpoints Used

| Endpoint | Purpose |
|----------|---------|
| `POST /oauth/token` | OAuth token refresh |
| `GET /api/v3/athlete` | Athlete profile data |
| `GET /api/v3/athlete/activities` | List all activities (paginated) |
| `GET /api/v3/activities/{id}` | Single activity details |
| `GET /api/v3/activities/{id}/zones` | Activity heart rate/power zones |
| `GET /api/v3/activities/{id}/streams` | Activity time-series data |
| `GET /api/v3/activities/{id}/photos` | Activity photos |
| `GET /api/v3/gear/{id}` | Gear details |
| `GET /api/v3/segments/{id}` | Segment details |
| `GET /api/v3/push_subscriptions` | Webhook subscription management |
| `POST /api/v3/push_subscriptions` | Create webhook subscription |
| `DELETE /api/v3/push_subscriptions/{id}` | Delete webhook subscription |
| `GET /athletes/{id}` | Public profile (for challenges scraping) |

### 2.4 Strava API Data Structures

#### Activity Object (from Strava API)

```json
{
  "id": 12345678901,
  "sport_type": "Ride",
  "name": "Morning Ride",
  "description": "Great weather today",
  "start_date_local": "2025-01-15T08:30:00Z",
  "distance": 45230.5,
  "total_elevation_gain": 523.0,
  "moving_time": 5400,
  "average_speed": 8.38,
  "max_speed": 15.2,
  "average_watts": 185,
  "max_watts": 650,
  "average_heartrate": 142.5,
  "max_heartrate": 178,
  "average_cadence": 82.3,
  "calories": 1250,
  "kudos_count": 15,
  "total_photo_count": 3,
  "commute": false,
  "workout_type": 10,
  "device_name": "Garmin Edge 530",
  "gear_id": "b12345678",
  "start_latlng": [51.5074, -0.1278],
  "map": {
    "summary_polyline": "encoded_polyline_string..."
  },
  "segment_efforts": [...],
  "splits_metric": [...],
  "splits_standard": [...],
  "laps": [...]
}
```

**Key Field Mappings:**
| Strava API Field | Unit | Notes |
|------------------|------|-------|
| `distance` | meters | Divide by 1000 for km |
| `total_elevation_gain` | meters | |
| `moving_time` | seconds | |
| `average_speed` | m/s | Convert to km/h (* 3.6) |
| `max_speed` | m/s | Convert to km/h (* 3.6) |
| `average_watts` | watts | Optional (requires power meter) |
| `average_heartrate` | bpm | Optional (requires HR monitor) |
| `average_cadence` | rpm/spm | Optional |
| `workout_type` | integer | See workout type mapping below |

**Workout Type Values:**
| Value | Type |
|-------|------|
| 0 | Default |
| 1 | Race |
| 2 | Long Run |
| 3 | Workout |
| 10 | Default (Ride) |
| 11 | Race (Ride) |
| 12 | Workout (Ride) |

#### Stream Data (from `/activities/{id}/streams`)

Available stream types:
| Stream Key | Description |
|------------|-------------|
| `time` | Seconds from start |
| `distance` | Meters from start |
| `latlng` | GPS coordinates array |
| `altitude` | Elevation in meters |
| `velocity_smooth` | Smoothed speed (m/s) |
| `heartrate` | Heart rate (bpm) |
| `cadence` | Cadence (rpm/spm) |
| `watts` | Power (watts) |
| `temp` | Temperature (°C) |
| `moving` | Boolean moving indicator |
| `grade_smooth` | Smoothed grade (%) |

#### Segment Effort Object (nested in Activity)

```json
{
  "segment": {
    "id": 123456,
    "name": "Segment Name",
    "distance": 2500.0,
    "average_grade": 4.5,
    "maximum_grade": 12.0,
    "climb_category": 2,
    "start_latlng": [lat, lng],
    "end_latlng": [lat, lng],
    "starred": true
  },
  "elapsed_time": 420,
  "moving_time": 415,
  "start_date_local": "2025-01-15T09:15:00Z",
  "pr_rank": 1,
  "average_watts": 220,
  "average_heartrate": 165
}
```

#### Gear Object (from `/gear/{id}`)

```json
{
  "id": "b12345678",
  "name": "Canyon Aeroad",
  "distance": 15234567,
  "retired": false,
  "primary": true
}
```

**Note:** Gear distance is in meters.

#### Split Object (nested in Activity)

```json
{
  "distance": 1000.0,
  "elapsed_time": 240,
  "moving_time": 235,
  "average_speed": 4.17,
  "average_heartrate": 145.2,
  "pace_zone": 3,
  "split": 1
}
```

#### Lap Object (nested in Activity)

```json
{
  "id": 123456789,
  "name": "Lap 1",
  "elapsed_time": 600,
  "moving_time": 595,
  "distance": 2500.0,
  "average_speed": 4.17,
  "max_speed": 5.5,
  "average_watts": 190,
  "average_heartrate": 148,
  "max_heartrate": 165,
  "average_cadence": 85,
  "lap_index": 1,
  "split": 1
}
```

### 2.5 Rate Limiting

Strava enforces two rate limits:
- **15-minute limit:** ~100 requests per 15 minutes
- **Daily limit:** ~1000 requests per day

Rate limit headers in responses:
- `x-ratelimit-limit`: Total limits (format: "15min,daily")
- `x-ratelimit-usage`: Current usage (format: "15min,daily")

---

## 3. Supported Sport Types

### Cycling
- Ride, MountainBikeRide, GravelRide, EBikeRide, EMountainBikeRide, VirtualRide, Velomobile

### Running
- Run, TrailRun, VirtualRun

### Walking
- Walk, Hike

### Water Sports
- Canoeing, Kayaking, Kitesurf, Rowing, StandUpPaddling, Surfing, Swim, Windsurf

### Winter Sports
- BackcountrySki, AlpineSki, NordicSki, IceSkate, Snowboard, Snowshoe

### Skating
- InlineSkate, RollerSki, Skateboard

### Racquet & Paddle Sports
- Badminton, Pickleball, Racquetball, Squash, TableTennis, Tennis

### Fitness
- Crossfit, WeightTraining, Workout, StairStepper, VirtualRow, Elliptical, HighIntensityIntervalTraining

### Mind & Body
- Pilates, Yoga

### Outdoor Sports
- Golf, RockClimbing, Sail, Soccer

### Adaptive Sports
- Handcycle, Wheelchair

---

## 4. Activity Data Model

### Core Activity Properties
| Field | Description |
|-------|-------------|
| Activity ID | Unique identifier (from Strava) |
| Start Date/Time | When the activity started |
| Sport Type | Type of activity (see Section 3) |
| World Type | Real World, Zwift, Rouvy, or MyWhoosh |
| Name | Activity title |
| Description | Optional activity description |
| Distance | Distance covered (km/mi) |
| Elevation | Total elevation gain (m/ft) |
| Starting Coordinate | GPS latitude/longitude |
| Calories | Calories burned |
| Average Power | Average power output (watts) |
| Max Power | Maximum power output (watts) |
| Average Speed | Average speed (km/h or mph) |
| Max Speed | Maximum speed achieved |
| Average Heart Rate | Average BPM |
| Max Heart Rate | Maximum BPM |
| Average Cadence | Average cadence (RPM/SPM) |
| Max Cadence | Maximum cadence |
| Moving Time | Active time (excluding stops) |
| Kudo Count | Social appreciation count |
| Device Name | Recording device |
| Photos | Activity photos |
| Polyline | Encoded route geometry |
| Route Geography | Countries passed through |
| Weather | Temperature, conditions at activity time |
| Gear | Associated equipment |
| Is Commute | Flag for commute activities |
| Workout Type | Race, Workout, Long Run, etc. |

### Derived/Calculated Data
| Field | Description |
|-------|-------------|
| Normalized Power | Weighted average power accounting for variability |
| Best Power Outputs | Peak power for various durations (5s, 10s, 30s, 1m, 5m, 8m, 20m, 1h) |
| Pace | Speed expressed as time per distance (sec/km, sec/100m) |
| Relative Power | Watts per kilogram (requires weight history) |
| Activity Intensity | Based on FTP and power data |

### Stream Data (Granular Time-Series)
- Heart rate stream
- Power stream
- Cadence stream
- Temperature stream
- Elevation stream
- GPS coordinates stream

---

## 5. Main Features

### 5.1 Dashboard

Configurable widget-based dashboard with the following widgets:

#### Most Recent Activities
- Displays N most recent activities (configurable)
- Shows activity name, date, distance, time, elevation
- Links to Strava activity

#### Intro Text
- Summary of total workout history
- Total activities, distance, time across all time

#### Training Goals
- Weekly, monthly, yearly, and lifetime goals
- Goal types: distance, elevation, moving time
- Configurable per sport type
- Visual progress indicators

#### Weekly Stats
- Current week statistics per sport type
- Distance, moving time, elevation
- Configurable metrics display order

#### Peak Power Outputs
- All-time best power outputs
- Duration intervals: 5s, 10s, 30s, 1m, 5m, 8m, 20m, 1h
- Displayed as bar chart

#### Heart Rate Zones
- Time distribution across 5 heart rate zones
- Configurable zones (relative % or absolute BPM)
- Zone ranges customizable by date and sport type

#### Activity Grid
- GitHub-style contribution graph
- Shows activity intensity per day
- Covers past 12 months

#### Monthly Stats
- Month-over-month comparison chart
- Distance, moving time, elevation per month
- Multi-year overlay for trend analysis
- Configurable years to display

#### Training Load
- Fitness and fatigue tracking over time
- Based on activity intensity calculations

#### Weekday Stats
- Activity distribution by day of week
- Pie/donut chart visualization

#### Day Time Stats
- Activity distribution by time of day
- Shows preferred workout times

#### Distance Breakdown
- Distance buckets per sport type
- Shows how many activities fall into distance ranges

#### Yearly Stats
- Year-over-year comparison
- Distance, time, elevation totals
- Multi-year chart overlay

#### Zwift Stats
- Zwift-specific statistics (if applicable)
- Virtual world activity summaries

#### Gear Stats
- Hours/distance per piece of gear
- Donut chart visualization
- Option to include/exclude retired gear

#### Eddington Widget
- Current Eddington numbers for configured sports
- Compact dashboard display

#### Challenge Consistency
- Monthly challenge completion tracking
- Configurable challenges:
  - Distance goals (total or single activity)
  - Elevation goals
  - Moving time goals
  - Number of activities
  - Calories burned
- Per sport type filtering

#### Most Recent Challenges
- Recently completed Strava challenges
- Challenge badges with completion dates

#### FTP History
- Functional Threshold Power over time
- Line chart showing FTP evolution

#### Athlete Weight History
- Body weight tracking over time
- Line chart visualization

### 5.2 Activities Page

Comprehensive activity listing with:

#### Filtering Options
- Sport type (multi-select)
- Date range (from/to)
- Country
- Gear
- Device
- Commute (yes/no)
- Workout type

#### Search
- Full-text search on activity names

#### Sortable Columns
- Date
- Distance
- Elevation
- Moving time
- Speed
- Heart rate
- Calories
- Power (average and best intervals)

#### Displayed Data Per Activity
- Date
- Activity name with Strava link
- Distance
- Elevation
- Moving time
- Average speed
- Average heart rate
- Calories
- Average power
- Best power outputs (5s, 10s, 30s, 1m, 5m, 8m, 20m, 1h)

#### Totals Row
- Sum of distance, elevation, time, calories for filtered results

### 5.3 Individual Activity View

Sport-type specific detail pages showing:

#### Common Elements
- Activity name and date
- Distance, elevation, time
- Speed (average/max)
- Heart rate (average/max)
- Cadence (average/max)
- Power data (if available)
- Weather conditions
- Gear used
- Route map (Leaflet-based)
- Activity photos

#### Sport-Specific Views
- **Running:** Pace display, split times, best efforts for standard distances
- **Swimming:** Pace per 100m, stroke data
- **Cycling:** Power zones, normalized power, relative power

#### Charts
- Elevation profile
- Heart rate over distance/time
- Power distribution
- Combined stream profile (HR, power, cadence, elevation)

### 5.4 Monthly Calendar View

Interactive calendar showing:
- Month navigation (previous/next)
- Activities displayed on their dates
- Color-coded by sport type
- Monthly statistics summary:
  - Total distance
  - Total elevation
  - Total time
  - Number of challenges completed
  - Total calories
  - Number of workouts

### 5.5 Segments & Efforts

#### Segment List
- Searchable segment database
- Filtering by:
  - Sport type
  - Country
  - Starred/favorite segments
  - KOM (King of Mountain) status

#### Segment Data
- Segment name
- Distance
- Maximum gradient
- Climb category (if applicable)
- Number of times completed
- Last effort date
- Best effort time

#### Segment Detail View
- Segment map
- All personal efforts history
- Best time progression

### 5.6 Best Efforts

Personal records for standard distances:

#### Running Distances
- 400m, 1/2 mile, 1km, 1 mile, 2 mile, 5km, 10km, 15km, 10 mile, 20km, Half Marathon, 30km, Marathon, 50km, 100km

#### Presentation
- Grouped by activity type
- Chart showing records over time
- Table with date, time, and activity link
- Click-through to effort history per distance

### 5.7 Eddington Number

Mathematical achievement tracking:
- Definition: Maximum number E where athlete covered at least E km/mi on at least E days
- Configurable per sport type groups
- Separate tabs for different sport groupings

#### Display Elements
- Current Eddington number
- Distribution chart (days vs. distance)
- History chart (Eddington progression over time)
- "Days needed" table showing efforts required for next N numbers

### 5.8 Heatmap

Geographic visualization of all activities:

#### Features
- Full-screen interactive map
- All activity routes overlaid
- Customizable polyline color
- Multiple tile layer support (street, satellite)
- Optional grayscale styling

#### Filtering
- Sport type
- Date range
- Commute filter
- Workout type

#### Statistics
- Number of routes displayed
- Countries covered (with percentage of world)

### 5.9 Strava Rewind (Year in Review)

Annual/all-time statistics showcase with comparison capability:

#### Available Metrics
- **Total activities count** by month
- **Distance per month** chart
- **Elevation per month** chart
- **Moving time per sport type** breakdown
- **Active days vs rest days** comparison
- **Activity start times** by hour
- **Personal records per month**
- **Activity locations** world map
- **Streaks** (consecutive active days, rest days)
- **Carbon saved** (estimated CO2 reduction from cycling commutes)
- **Socials** (kudos received)
- **Biggest activities** (longest, most elevation, etc.)
- **Random photo** from the year

#### Comparison Feature
- Compare any two years or year vs. all-time
- Side-by-side metric display

### 5.10 Challenges

Strava challenge tracking:

#### Display
- Grouped by completion month
- Challenge badge images
- Challenge names
- Links to Strava challenge pages

#### Import
- Automatic import of visible challenges from public profile
- Manual import option via trophy case HTML export

### 5.11 Photos Gallery

Activity photo collection:

#### Features
- Masonry-style photo wall
- Lazy loading for performance
- Lightbox slideshow mode

#### Filtering
- Sport type (multi-select)
- Country

#### Photo Metadata
- Activity name overlay
- Activity date
- Link to source activity

### 5.12 Gear Management

#### Strava Gear
- Imported automatically from Strava
- Bikes, shoes tracked
- Distance, elevation, time per gear piece
- Active vs. retired status

#### Custom Gear
- User-defined gear not trackable in Strava
- Examples: skateboards, kayaks, snowboards
- Linked via hashtags in activity titles
- Same statistics as Strava gear

#### Gear Statistics Page
- Per-gear metrics:
  - Number of workouts
  - Total and average distance
  - Total elevation
  - Total moving time
  - Average speed
  - Total calories
  - Purchase price (optional)
  - Relative cost per hour
  - Relative cost per activity

#### Charts
- Distance per month per gear
- Distance over time per gear

### 5.13 Gear Maintenance

Component-based maintenance tracking:

#### Features
- Define components (chain, cassette, brake pads, etc.)
- Attach components to specific gear
- Set maintenance intervals:
  - Distance-based (every X km/mi)
  - Time-based (every X hours used)
  - Calendar-based (every X days)
- Track maintenance via hashtags in activity titles
- Visual progress indicators
- Component images support

#### Maintenance History
- Log of completed maintenance tasks
- Date and trigger activity

### 5.14 User Badges

Embeddable SVG badges for external use:

#### Badge Types
- **User Badge:** Overall statistics summary
- **PB Badges:** Personal best records per sport type
- **Zwift Badge:** Zwift level and racing score

#### Usage
- Dynamically generated SVG images
- Embed code provided for websites/forums

### 5.15 AI Workout Assistant

AI-powered analysis and recommendations:

#### Features
- Chat interface (CLI and optional web UI)
- Activity analysis and performance assessment
- Training recommendations
- Comparison of workout periods
- Pre-defined command shortcuts

#### Supported Providers
- Anthropic (Claude)
- OpenAI
- Azure OpenAI
- Google Gemini
- Ollama (self-hosted)
- DeepSeek
- Mistral

---

## 6. Configuration System

### 6.1 General Settings
- App URL (for PWA manifest)
- App subtitle (for multi-instance identification)
- Profile picture URL
- Locale selection
- Unit system (metric/imperial)
- Time format (12h/24h)
- Date format patterns

### 6.2 Athlete Configuration
- Birthday (for heart rate calculations)
- Max heart rate formula selection:
  - Fox (220 - age)
  - Tanaka
  - Astrand
  - Gellish
  - Nes
  - Arena
  - Fixed value with date ranges
- Weight history (date-keyed)
- FTP history (date-keyed, separate for cycling/running)
- Heart rate zones:
  - Mode: relative (%) or absolute (BPM)
  - 5 configurable zones
  - Date range overrides
  - Sport-type specific overrides

### 6.3 Import Settings
- Activities per import batch (rate limit management)
- Sport types to import (whitelist)
- Activity visibility filter (everyone, followers_only, only_me)
- Skip activities before date
- Skip specific activity IDs
- Opt-in to segment detail import

### 6.4 Dashboard Configuration
- Widget selection and ordering
- Widget width (33%, 50%, 66%, 100%)
- Per-widget configuration options

### 6.5 Appearance Settings
- Heatmap polyline color
- Heatmap tile layer URL(s)
- Heatmap grayscale toggle
- Photo sport type hiding
- Photo default filters
- Sport type display ordering

### 6.6 Eddington Configuration
- Multiple Eddington definitions
- Sport type groupings per definition
- NavBar visibility toggle
- Dashboard widget visibility toggle

### 6.7 Notification Settings
- Shoutrrr integration for multi-service notifications
- Supported services: ntfy, Slack, Discord, etc.

### 6.8 Daemon/Scheduler
- Scheduled actions:
  - Import data and rebuild app
  - Gear maintenance notifications
  - App update notifications
- Cron expression configuration

---

## 7. Technical Characteristics

### 7.1 Architecture Pattern
- Static site generation (HTML/JSON files)
- Pre-built dashboard for fast loading
- Client-side rendering with JavaScript enhancements

### 7.2 Data Storage
- SQLite database
- Local file storage for images
- JSON exports for frontend consumption

### 7.3 Frontend Technologies
- Tailwind CSS for styling
- Leaflet.js for maps
- ECharts for charts
- DataTables for sortable/filterable tables
- Flowbite for UI components
- LightGallery for photo viewing

### 7.4 PWA Features
- Installable on devices
- Offline-capable
- App manifest generation

### 7.5 Localization
- Translation system with locale files
- Date/time formatting per locale
- Number formatting per locale
- Measurement unit conversion

---

## 8. User Workflows

### 8.1 Initial Setup
1. Create Strava API application
2. Deploy Docker container
3. Configure environment variables
4. Access web UI
5. Authorize with Strava OAuth
6. Wait for initial data import
7. Access dashboard

### 8.2 Ongoing Usage
1. Complete activities on Strava
2. Webhook triggers automatic import (or scheduled/manual)
3. View updated statistics on dashboard
4. Explore detailed views as needed

### 8.3 Maintenance Tracking
1. Configure gear and components in YAML
2. Complete activities using the gear
3. Add maintenance hashtag to activity title when maintenance performed
4. System tracks and resets counters

---

## 9. Data Export Formats

### Activity Data Table (JSON)
- All filterable/sortable activity data
- Used for client-side table rendering

### Segment Data Table (JSON)
- All segment data with efforts
- Client-side searchable

### Chart Data (JSON)
- ECharts-compatible options objects
- Pre-calculated for performance

### Badge Files (SVG)
- Dynamically generated
- Embeddable graphics

### GPX Export
- Standard GPS exchange format
- Per-activity download capability

---

## 10. Visualization Types

### 10.1 ECharts-Based Charts

All interactive charts are rendered using Apache ECharts library with custom configurations.

#### Line Charts
| Usage Location | Description |
|----------------|-------------|
| Monthly Stats | Distance/elevation/time trends per month with multi-year overlay |
| Yearly Stats | Year-over-year distance comparison |
| FTP History | Functional Threshold Power progression over time |
| Weight History | Athlete body weight changes over time |
| Training Load | Fitness and fatigue curves over time |
| Eddington History | Eddington number progression over time |
| Combined Stream Profile | Heart rate, power, cadence, elevation over distance/time |
| Power Output | Power curve (best efforts at different durations) |
| Weekly Stats | Weekly metrics trend |
| Personal Records | PRs achieved per month (Rewind) |
| Activity Start Times | Distribution of workout start hours (Rewind) |
| Distance Over Time Per Gear | Cumulative distance per gear piece |

#### Bar Charts
| Usage Location | Description |
|----------------|-------------|
| Eddington Chart | Days at each distance bucket |
| Best Effort Chart | Personal records by distance |
| Heart Rate Distribution | Time spent in each HR zone (stacked) |
| Power Distribution | Time spent in each power zone (stacked) |
| Velocity Distribution | Time spent at different speeds |
| Training Load Bars | Daily training stress scores |
| Distance Per Month Per Gear | Monthly distance by gear (stacked) |
| Activity Count Per Month | Number of activities per month (Rewind) |
| Distance Per Month | Monthly distance totals (Rewind) |
| Elevation Per Month | Monthly elevation totals (Rewind) |

#### Pie/Donut Charts
| Usage Location | Description |
|----------------|-------------|
| Daytime Stats | Activity distribution by time of day (morning/afternoon/evening/night) |
| Weekday Stats | Activity distribution by day of week |
| Time in Heart Rate Zones | Percentage in each HR zone |
| Moving Time Per Gear | Hours spent on each piece of equipment |
| Moving Time Per Sport Type | Time breakdown by sport (Rewind) |
| Rest Days vs Active Days | Active/rest day ratio (Rewind) |

#### Calendar Heatmap
| Usage Location | Description |
|----------------|-------------|
| Activity Grid | GitHub-style contribution graph showing daily activity intensity |
| Daily Activities (Rewind) | Year calendar with activity markers |

#### Geo/World Map
| Usage Location | Description |
|----------------|-------------|
| Activity Locations (Rewind) | World map with activity hotspots using effectScatter |

### 10.2 Leaflet.js Maps

Interactive geographic maps for route visualization.

| Map Type | Description |
|----------|-------------|
| Heatmap Page | Full-screen map with all activity polylines overlaid |
| Activity Route Map | Individual activity route with elevation coloring |
| Segment Map | Segment route visualization |
| Zwift World Maps | Custom tile layers for Zwift virtual worlds (Watopia, London, etc.) |
| Rouvy World Maps | Virtual route mapping |
| MyWhoosh World Maps | Virtual route mapping |

**Map Features:**
- Multiple tile layer options (OpenStreetMap, satellite, custom)
- Polyline rendering with configurable color
- Grayscale filter option
- Zoom and pan controls
- Route start/end markers
- Country boundary overlays

### 10.3 Data Tables

Interactive sortable/filterable tables using DataTables.js with Clusterize.js for virtual scrolling.

| Table | Features |
|-------|----------|
| Activities List | Sort by 15+ columns, filter by sport/date/country/gear/device, search by name |
| Segments List | Sort by name/distance/gradient/ride count, filter by sport/country/starred/KOM |
| Best Efforts Table | Per-distance records with date and link to activity |
| Gear Statistics | Per-gear metrics with expandable retired gear section |
| Gear Maintenance | Component status with maintenance history accordion |
| Challenge List | Grouped by month with challenge counts |

**Table Features:**
- Column sorting (ascending/descending)
- Multi-filter dropdowns
- Date range pickers
- Full-text search
- Live totals row (updates on filter)
- Virtual scrolling for large datasets
- Sticky headers
- Responsive horizontal scroll

### 10.4 Progress Indicators

#### Linear Progress Bars
| Usage | Description |
|-------|-------------|
| Training Goals | Horizontal bar showing % progress toward distance/elevation/time goals |
| Gear Maintenance | Component wear indicator showing % toward maintenance threshold |

**Progress Bar States:**
- Empty (0%)
- Partial fill with percentage label
- Complete (100%)
- Overdue (maintenance past due - visual warning)

### 10.5 Photo Gallery

Flexbox-based responsive photo wall with lightbox.

**Features:**
- Fluid grid layout (photos grow to fill rows)
- Lazy loading (placeholder until scroll into view)
- Hover overlay with activity name and date
- LightGallery.js slideshow mode
- Filter by sport type and country
- Photo count display

### 10.6 Monthly Calendar

CSS Grid-based calendar visualization.

**Features:**
- 7-column grid (Mon-Sun)
- Activities listed in date cells
- Color-coded by activity type
- Navigation arrows (previous/next month)
- Header with monthly totals
- Modal display

### 10.7 SVG Badges

Dynamically generated SVG images for embedding.

| Badge Type | Content |
|------------|---------|
| Strava Badge | Athlete name, total activities, total distance, total time, total elevation |
| PB Badge | Personal best times for standard distances per sport type |
| Zwift Badge | Zwift level, racing score, distance ridden |

**Badge Characteristics:**
- Fixed dimensions for consistent embedding
- Strava orange color scheme
- Icon integration
- Responsive text sizing
- URL-embeddable format

### 10.8 Color Schemes

#### Activity Type Colors
| Activity Type | Color |
|---------------|-------|
| Ride | Emerald (#10b981) |
| Run | Orange (#f97316) |
| Walk | Yellow (#facc15) |
| Swim | Blue (#2563eb) |
| Winter Sports | Red (#dc2626) |
| Other | Slate (#475569) |

#### Intensity Colors (Activity Grid)
- Level 0: Gray (no activity)
- Level 1-4: Progressive orange intensity

#### Heart Rate Zone Colors
- Zone 1 (Recovery): Light blue
- Zone 2 (Aerobic): Green
- Zone 3 (Tempo): Yellow
- Zone 4 (Threshold): Orange
- Zone 5 (Anaerobic): Red

---

## 11. Navigation, Drill-Downs & Filtering

### 11.1 Dashboard Widget Drill-Downs

Each dashboard widget provides navigation to detailed views:

| Widget | Drill-Down Target | Link Text |
|--------|-------------------|-----------|
| Most Recent Activities | Activities page | "View all" |
| Gear Stats | Gear statistics page | "View details" |
| Eddington | Eddington detail page | "View details" |
| Most Recent Challenges | Challenges page | "View all" |
| Training Load | Training load detail page | "View details" |
| Peak Power Outputs | Power output analysis page | "View details" |
| Heart Rate Zones | (inline display, no drill-down) | - |
| Activity Grid | (inline display, no drill-down) | - |

### 11.2 Activities Page Filtering

#### Filter Categories

| Filter | Type | Values |
|--------|------|--------|
| Sport Type | Multi-select checkboxes | All imported sport types |
| Date Range | Date picker (from/to) | Any date range |
| Country | Radio buttons with flags | Countries with recorded activities |
| Gear | Radio buttons | All gear + "Unspecified" |
| Device | Radio buttons | All devices + "Unspecified" |
| Commute | Radio buttons | "Commutes only" / "Excludes commutes" |
| Workout Type | Radio buttons | Default, Race, Workout, Long Run |

#### Search
- Full-text search on activity name (case-insensitive)
- Real-time filtering as user types

#### Sortable Columns

| Column | Sort Key | Direction |
|--------|----------|-----------|
| Date | start-date | Asc/Desc |
| Distance | distance | Asc/Desc |
| Elevation | elevation | Asc/Desc |
| Moving Time | moving-time | Asc/Desc |
| Average Speed | speed | Asc/Desc |
| Heart Rate | heart-rate | Asc/Desc |
| Calories | calories | Asc/Desc |
| Average Power | power | Asc/Desc |
| Best 5s Power | power-5s | Asc/Desc |
| Best 10s Power | power-10s | Asc/Desc |
| Best 30s Power | power-30s | Asc/Desc |
| Best 1m Power | power-60s | Asc/Desc |
| Best 5m Power | power-300s | Asc/Desc |
| Best 8m Power | power-480s | Asc/Desc |
| Best 20m Power | power-1200s | Asc/Desc |
| Best 1h Power | power-3600s | Asc/Desc |

#### Live Totals
Footer row displays sum of filtered results:
- Total distance
- Total elevation
- Total moving time (hours)
- Total calories

### 11.3 Segments Page Filtering

| Filter | Type | Values |
|--------|------|--------|
| Sport Type | Radio buttons | Ride, Run, etc. |
| Country | Radio buttons with flags | Countries with segments |
| Starred | Checkbox | Show only starred segments |
| KOM Status | Checkbox | Show only KOM/climb segments |

#### Sortable Columns
- Segment name (alphabetical)
- Distance
- Maximum gradient
- Ride count (times completed)
- Last effort date

### 11.4 Heatmap Page Filtering

| Filter | Type | Values |
|--------|------|--------|
| Sport Type | Radio buttons | All sport types |
| Date Range | Date picker (from/to) | Any date range |
| Commute | Radio buttons | "Commutes only" / "Excludes commutes" |
| Workout Type | Radio buttons | Default, Race, Workout, Long Run |

#### Live Updates
- Route count updates in real-time
- Map redraws filtered polylines
- "Clear" link to reset all filters

### 11.5 Photos Page Filtering

| Filter | Type | Values |
|--------|------|--------|
| Sport Type | Multi-select checkboxes | Sport types with photos |
| Country | Radio buttons with flags | Countries where photos were taken |

#### Default Filters
- Configurable default filters via settings
- Persisted across sessions

### 11.6 Cross-Page Deep Linking

Navigation patterns that pass filter state between pages:

| Source | Action | Target | Pre-applied Filter |
|--------|--------|--------|-------------------|
| Gear Statistics Table | Click gear name | Activities page | gear={gearId} |
| Monthly Calendar | Click month | Month detail modal | month={month} |
| Best Efforts Table | Click distance | Distance efforts modal | distance={distance} |
| Segment List | Click segment | Segment detail modal | segment={segmentId} |
| Dashboard Activity | Click activity | Activity detail page | activity={activityId} |

### 11.7 Modal/Overlay Drill-Downs

Certain views open in modal overlays rather than page navigation:

| Trigger | Modal Content |
|---------|---------------|
| Click distance in Best Efforts | All efforts for that distance with progression chart |
| Click month in calendar | Full month view with daily activities |
| Click segment in list | Segment detail with all personal efforts |
| Click "info" icon on pages | Help/documentation overlay |
| Click photo | Lightbox with slideshow controls |

### 11.8 External Links

Links that navigate to Strava:

| Element | Link Target |
|---------|-------------|
| Activity name | `strava.com/activities/{id}` |
| Segment name | `strava.com/segments/{id}` |
| Challenge badge | `strava.com/challenges/{slug}` |
| Gear name (in activity) | `strava.com/gear/{id}` |

### 11.9 URL State Persistence

Filters can be applied via URL parameters:

```
/activities?sportType=Ride&gear=b12345
/heatmap?sportType=Run&isCommute=false
/segments?isFavourite=true&sportType=Ride
```

This enables:
- Bookmarking filtered views
- Sharing specific filtered views
- Deep linking from external sources
- Browser back/forward navigation

### 11.10 Tab Navigation

Multi-tab views within pages:

| Page | Tab Options |
|------|-------------|
| Best Efforts | Activity type tabs (Run, Ride, etc.) |
| Gear Stats Widget | "All" + per-activity-type tabs |
| Training Goals | Period tabs (Weekly, Monthly, Yearly, Lifetime) |
| Eddington | Sport type grouping tabs |
| Rewind | Year comparison tabs |

### 11.11 Breadcrumb Navigation

Hierarchical breadcrumb trail on all pages:
- Dashboard (home)
- Current page
- Sub-page (if applicable)

Examples:
- `Dashboard > Activities`
- `Dashboard > Gear > Maintenance`
- `Dashboard > Segments > [Segment Name]`

---

## 12. Accessibility Features

- Semantic HTML structure
- ARIA labels on interactive elements
- Keyboard navigation support
- Screen reader compatible tables
- High contrast color usage

---

## 11. Performance Considerations

- Lazy loading for images
- Virtual scrolling for large tables (Clusterize)
- Pre-rendered static HTML
- Client-side caching
- Optimized chart rendering

---

## 12. Security Considerations

- OAuth 2.0 for Strava authentication
- Self-hosted (data doesn't leave user infrastructure)
- AI features disabled by default with warning for public instances
- No user authentication layer (single-user assumption)
- Rate limiting awareness to prevent API bans

---

*This specification reflects the feature set as of version 4.3.0 (December 2025).*
