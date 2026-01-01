# Quick Start

Get your Strava data into Quantlete in 5 minutes.

## Prerequisites

- Quantlete installed and running ([Installation](README.md))
- Strava API credentials ([Strava Setup](strava-setup.md))

## Step 1: Start the Server

```bash
# With environment variables
export QUANTLETE_STRAVA_CLIENT_ID=your_client_id
export QUANTLETE_STRAVA_CLIENT_SECRET=your_client_secret
quantlete serve

# Or with Docker
docker run -d -p 8081:8081 \
  -e QUANTLETE_STRAVA_CLIENT_ID=your_client_id \
  -e QUANTLETE_STRAVA_CLIENT_SECRET=your_client_secret \
  ghcr.io/melonamin/quantlete:latest
```

## Step 2: Connect Your Strava Account

1. Open http://localhost:8081 in your browser
2. Click **"Connect with Strava"**
3. Authorize Quantlete to read your activities
4. You'll be redirected back to Quantlete

## Step 3: Import Your Activities

After connecting, import your activity history:

```bash
# Import all activities
quantlete import

# Or from within Docker
docker exec quantlete quantlete import
```

The import shows progress as it works through:
- Activity list (basic info)
- Gear details
- Activity streams (GPS, heart rate, power)
- Segment efforts
- Photos

?> **Tip:** For large histories, you can interrupt the import with Ctrl+C and resume later with `quantlete import --resume`.

## Step 4: Explore Your Data

Open the dashboard at http://localhost:8081 to see:

- **Dashboard** - Overview stats and recent activities
- **Activities** - Searchable list of all activities
- **Calendar** - Monthly view of your training
- **Heatmap** - All your activities on a map
- **Training Load** - Fitness and fatigue tracking
- **Gear** - Equipment usage and maintenance

## Keeping Data in Sync

Quantlete can automatically import new activities via Strava webhooks. See [Configuration](../configuration/README.md) for webhook setup.

For manual syncing:
```bash
# Quick sync (new activities only)
quantlete import

# Full sync (refresh all data)
quantlete import --full
```

## Next Steps

- [Configuration options](../configuration/README.md)
- [CLI reference](../cli/README.md)
- [Deployment guide](../deployment/README.md)
