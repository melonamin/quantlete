# demo

Generate demo data for testing and demonstration.

## Usage

```bash
quantlete demo [OPTIONS]
```

## Options

| Flag | Description | Default |
|------|-------------|---------|
| `--activities` | Number of activities to generate | 100 |
| `--months` | Number of months to spread activities across | 12 |
| `--athlete` | Name of the demo athlete | "Demo User" |

## Description

The `demo` command generates realistic fake data for exploring the Quantlete dashboard without connecting to Strava. It creates:

- A demo athlete profile
- Activities with realistic metrics (distance, elevation, heart rate, power)
- Gear (bikes, shoes)
- Segment efforts
- Training load data

!> **Warning:** This command wipes any existing data before generating demo data.

## Examples

### Default Demo Data

```bash
quantlete demo
```

Creates 100 activities over 12 months.

### Extended History

```bash
quantlete demo --activities=200 --months=24
```

Creates 200 activities over 2 years.

### Custom Athlete Name

```bash
quantlete demo --athlete="Jane Doe"
```

### Quick Demo

```bash
quantlete demo --activities=20 --months=3
```

Creates a smaller dataset for quick testing.

## Use Cases

### Testing the Dashboard

```bash
quantlete demo
quantlete serve
# Open http://localhost:8081
```

### Trying Before Connecting Strava

Demo mode lets you explore all dashboard features before providing Strava credentials.

### Development

Generate test data during development:

```bash
quantlete demo --activities=50
```

## Notes

- Demo mode is indicated in the dashboard
- The demo athlete has randomized but realistic metrics
- Activities include various types: Run, Ride, Swim, etc.
- Training load and fitness calculations work with demo data
