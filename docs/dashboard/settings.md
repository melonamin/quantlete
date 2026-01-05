# Settings

Configure your Quantlete experience.

## Display Preferences

### Units

Choose your measurement system:
- **Metric** - Kilometers, meters, kg
- **Imperial** - Miles, feet, lbs

### Date Format

Select date display format:
- YYYY-MM-DD
- DD/MM/YYYY
- MM/DD/YYYY

### Week Start

Choose first day of week for calendar:
- Sunday
- Monday

## Goals

Set training targets:

### Weekly Goals

- Distance target
- Time target
- Elevation target
- Activity count target

### Annual Goals

- Total distance
- Total elevation
- Specific activity goals

Goals appear on the dashboard with progress indicators.

## Athlete Profile

Configure your athlete data for accurate calculations:

See [Calculations](../calculations.md) for how FTP and HR zones are used.

### FTP (Functional Threshold Power)

Set your cycling FTP for:
- Power zone calculations
- Training load (TSS)
- Performance analysis

### Heart Rate Zones

Configure heart rate zones for:
- Zone-based analysis
- Effort calculations

### Weight

Set weight for:
- Power-to-weight ratios
- Calorie calculations

## Data Export

Export your data in various formats:

### CSV Export

Download activity data as CSV:
- All activities
- Filtered selection
- Include/exclude specific fields

### JSON Export

Full data export in JSON format:
- Complete activity data
- Streams and segments
- Gear and maintenance

### Backup

Create a complete backup:
- Database file
- All settings
- Maintenance logs

## Strava Connection

Manage your Strava integration:

### Connection Status

View current connection state:
- Connected athlete
- Token expiration
- Last sync time

### Reconnect

If authentication expires:
1. Click "Reconnect to Strava"
2. Authorize in Strava
3. Return to Quantlete

### Disconnect

Remove Strava connection:
- Revokes access token
- Keeps imported data
- Stops automatic syncs

## Notifications

Configure push notifications for events and achievements. See [Notifications Configuration](../configuration/notifications.md) for details.

### Notification Services

Add one or more notification services:
- **Telegram** - Send to a Telegram chat
- **Email (SMTP)** - Send via email
- **Generic Webhook** - Send to any HTTP endpoint

### Event Types

Choose which events trigger notifications:

- **Import Complete** - When sync finishes
- **Achievements** - PRs, Eddington increases, goal completions
- **Training Load** - Fatigue warnings, recovery alerts
- **Maintenance** - Component service reminders
- **Digests** - Weekly/monthly activity summaries

### Testing

Use the "Test" button to verify each service is configured correctly.

## Advanced

### Debug Mode

Enable debug logging:
- Detailed API logs
- Performance metrics
- Troubleshooting info

### Database

View database statistics:
- Activity count
- Database size
- Last migration version
