# import

Import activities from Strava into the local database.

## Usage

```bash
quantlete import [OPTIONS]
```

## Options

| Flag | Description | Default |
|------|-------------|---------|
| `--full` | Perform a full sync (re-import all activities) | false |
| `--resume` | Resume from previous interrupted import | false |
| `--skip-streams` | Skip importing activity stream data (GPS, heartrate, etc.) | false |
| `--skip-segments` | Skip importing segment efforts and segment details | false |
| `--skip-best-efforts` | Skip importing Strava best efforts/PRs | false |
| `--skip-photos` | Skip importing activity photos | false |

## Prerequisites

Before running import, you must:

1. Have Strava API credentials configured
2. Authenticate via the web interface at http://localhost:8081

If not authenticated, the command will show instructions:

```
Not authenticated with Strava
Please start the server and authenticate via the web interface first:
  quantlete serve
  Then visit http://localhost:8081 and click "Connect with Strava"
```

## Description

The import command fetches your Strava data in phases:

1. **Activities** - Basic activity info (name, date, distance, etc.)
2. **Gear** - Bikes and shoes linked to activities
3. **Streams** - Detailed data (GPS track, heart rate, power, cadence)
4. **Activity Details** - Weather, perceived exertion, etc.
5. **Segments** - Segment efforts and segment definitions
6. **Photos** - Activity photos

Progress is displayed during import:

```
Import progress  phase=streams  progress="streams 45/120"  failed=0  eta="2m30s"
```

## Examples

### Standard Import

```bash
quantlete import
```

Imports new activities since the last sync.

### Full Re-sync

```bash
quantlete import --full
```

Re-imports all activities, updating any changed data.

### Resume Interrupted Import

```bash
quantlete import --resume
```

Continues from where a previous import was interrupted (Ctrl+C).

### Quick Import (Skip Heavy Data)

```bash
quantlete import --skip-streams --skip-segments --skip-photos
```

Faster import that skips large data types. Useful for quick updates.

### Import Only Activity List

```bash
quantlete import --skip-streams --skip-segments --skip-best-efforts --skip-photos
```

Imports only basic activity info and gear.

## Rate Limits

Strava enforces API rate limits:

- **15-minute limit**: 100 requests
- **Daily limit**: 1,000 requests

Quantlete tracks these limits and will pause automatically when approaching them. Rate limit state is persisted, so you can interrupt and resume imports safely.

If rate limited, wait for the limit to reset:
- 15-minute limit: Wait ~15 minutes
- Daily limit: Wait until midnight UTC

## Interrupt and Resume

You can safely interrupt an import with Ctrl+C:

```
^C
Interrupt received, stopping import...
Import canceled  phase=streams  activities=50  failed=0
```

Resume later with:

```bash
quantlete import --resume
```

The import will continue from where it left off.
