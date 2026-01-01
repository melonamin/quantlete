# Quantlete

Self-hosted analytics dashboard for your Strava activities.

## What is Quantlete?

Quantlete is a personal analytics platform that imports your Strava data and provides detailed insights into your training. All data stays on your own infrastructure - no cloud services, no subscriptions, complete privacy.

## Features

- **Activity Analytics** - Detailed stats, elevation profiles, power/HR zones
- **Training Load** - Track fitness, fatigue, and form over time
- **Calendar & Heatmap** - Visualize your activities on maps and calendars
- **Gear Tracking** - Monitor equipment usage and maintenance
- **Personal Records** - Best efforts by distance and time
- **Eddington Number** - Track your cycling Eddington number
- **Year in Review** - Strava Rewind-style annual summaries
- **Data Export** - Download your stats in CSV/JSON formats

## Quick Links

- [Installation](getting-started/README.md) - Get Quantlete running
- [Quick Start](getting-started/quick-start.md) - Import your first activities
- [Configuration](configuration/README.md) - Customize your setup
- [CLI Reference](cli/README.md) - Command line usage

## Requirements

- Strava account with activities
- Strava API credentials ([setup guide](getting-started/strava-setup.md))
- One of:
  - Docker
  - Linux/macOS/Windows system for binary
  - Go 1.23+ and Node.js 22+ (building from source)

## How It Works

1. **Connect Strava** - Authorize Quantlete to read your activities
2. **Import Data** - Pull your activity history from Strava
3. **Explore** - View your stats in the web dashboard
4. **Stay Synced** - New activities import automatically via webhooks

## Privacy

Quantlete is designed for self-hosting. Your data never leaves your infrastructure:

- SQLite database stored locally
- No external analytics or tracking
- No cloud dependencies (except Strava API)
- Open source - audit the code yourself
