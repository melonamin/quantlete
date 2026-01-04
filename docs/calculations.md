# Calculations and Data Sources

Quantlete computes analytics locally using your Strava activity data. This page documents the inputs and formulas behind key metrics.

## Data Sources

- Activity summary fields from Strava (distance, duration, elevation, sport type, etc.)
- Streams (when imported): `watts`, `velocity_smooth`, `heartrate`
- Athlete settings you set in the dashboard: cycling FTP, running threshold speed, HR zones
- Computed values are stored in SQLite; no third-party analytics are used

## Training Stress Score (TSS)

TSS is computed per activity from streams and your thresholds.

### Cycling (power-based)

Required: power stream (`watts`) and cycling FTP.

- Normalized Power (NP) uses a 30-second rolling average and 4th-power mean.
- Intensity Factor (IF) = NP / FTP
- TSS = (duration_seconds * NP * IF) / (FTP * 3600) * 100

### Running (speed-based)

Required: `velocity_smooth` stream and running threshold speed (FTP running, m/s).

- Uses the same NP/IF/TSS formula with speed data.

### Running (HR-based fallback)

Required: `heartrate` stream and HR zones.

- Uses the 4th boundary from your HR zone definition as the threshold HR (zone 4 lower bound when zones are defined as bounds).
- If zones are defined as percent of HR max, the boundary is multiplied by HR max.
- Uses the same NP/IF/TSS formula with heart rate data.
- This is an approximation to keep training load consistent when power/speed data is missing.

Notes:
- Duration is moving time, not elapsed time.
- Normalized Power assumes 1 Hz streams. If fewer than 30 samples are available, it falls back to a simple average.
- If the required stream or threshold is missing, no TSS is computed for that activity.

## Training Load (CTL/ATL/TSB)

Daily TSS is the sum of all activity TSS values per local day (`start_date_local`). CTL and ATL are computed as exponential moving averages:

- CTL tau = 42 days
- ATL tau = 7 days
- new = old + (tss - old) * (1 / tau)
- TSB = CTL - ATL

Rest days are included as TSS = 0. The series starts at 0 and builds over time.

## Power Zones

Power zones are defined as percentages of your cycling FTP:

| Zone | % of FTP | Name | Purpose |
|------|----------|------|---------|
| Z1 | 0–55% | Recovery | Active recovery, warm-up |
| Z2 | 55–75% | Endurance | Aerobic base building |
| Z3 | 75–90% | Tempo | Sustained aerobic efforts |
| Z4 | 90–105% | Threshold | Lactate threshold training |
| Z5 | 105%+ | VO2max/Anaerobic | High-intensity intervals |

Time-in-zone is calculated from 1 Hz power streams. Each sample is assigned to a zone based on the ratio `watts / FTP`. FTP must be configured in Settings → Athlete Profile.

## Heart Rate Zones

HR zones support two configuration methods:

### Percent of HR Max (default)

Zone boundaries are defined as percentages of your maximum heart rate:

| Zone | Default % | Example (HR max = 185) |
|------|-----------|------------------------|
| Z1 | 0–60% | 0–111 bpm |
| Z2 | 60–70% | 111–130 bpm |
| Z3 | 70–80% | 130–148 bpm |
| Z4 | 80–90% | 148–167 bpm |
| Z5 | 90–100% | 167–185 bpm |

Formula: `zone_boundary_bpm = hr_max × percentage`

### Absolute BPM

Alternatively, you can define zone boundaries as absolute heart rate values (e.g., 120, 140, 160, 175 bpm). This method does not require HR max.

### Configuration

- HR zones can be configured per sport (All, Run, Ride) in Settings → Athlete Profile
- Each zone definition has an effective date, allowing you to track changes over time
- Time-in-zone analysis uses the zone definition applicable to each activity's date

## Power Curve (Peak Power)

Peak power is computed from 1 Hz power streams using a rolling max average:

- Durations: 5, 10, 30, 60, 300, 480, 1200, 3600 seconds
- Negative power samples are treated as 0
- If a power stream is missing, the power curve is not available

## Eddington Number

Eddington numbers are computed from daily total distances (stored in kilometers) for the selected activity types:

- E is the largest number where you have E days with distance >= E.
- Defaults include all activities, rides, and runs (configurable in settings).

## Best Efforts and PRs

- Distance/time best efforts are imported from Strava activity details when available.
- Power best efforts come from Quantlete's power curve computation.

## Missing Data Behavior

- If you skip streams (`quantlete import --skip-streams`), TSS, training load, and power curve metrics cannot be computed until streams are imported.
- If FTP or HR zones are not configured, TSS will be missing for activities that depend on those thresholds.
