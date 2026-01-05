# Notifications

Quantlete can send push notifications for various events like import completion, achievements, and maintenance reminders.

## Overview

Notifications are configured per-athlete through the Settings page. You can:

- Enable/disable notifications globally
- Configure multiple notification services
- Choose which events trigger notifications

## Supported Services

### Telegram

Send notifications to a Telegram chat or group.

**Configuration:**

| Field | Description |
|-------|-------------|
| `token` | Bot token from [@BotFather](https://t.me/BotFather) |
| `chat_id` | Chat ID (use [@userinfobot](https://t.me/userinfobot) to find yours) |

**Setup steps:**

1. Create a bot with [@BotFather](https://t.me/BotFather)
2. Copy the bot token (format: `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`)
3. Start a chat with your bot
4. Get your chat ID from [@userinfobot](https://t.me/userinfobot)
5. Enter both values in Quantlete settings

### Email (SMTP)

Send notifications via email.

**Configuration:**

| Field | Description |
|-------|-------------|
| `host` | SMTP server hostname |
| `port` | SMTP port (optional, defaults to 25) |
| `user` | SMTP username (optional) |
| `password` | SMTP password (optional) |
| `from` | Sender email address |
| `to` | Recipient email address |

**Example for Gmail:**

| Field | Value |
|-------|-------|
| `host` | smtp.gmail.com |
| `port` | 587 |
| `user` | your.email@gmail.com |
| `password` | App password (not your Google password) |
| `from` | your.email@gmail.com |
| `to` | your.email@gmail.com |

> **Note:** Gmail requires an [App Password](https://support.google.com/accounts/answer/185833) when 2FA is enabled.

### Generic Webhook

Send notifications to any HTTP endpoint.

**Configuration:**

| Field | Description |
|-------|-------------|
| `url` | Webhook URL (http, https, or generic scheme) |

The webhook receives a POST request with the notification message. Compatible with services like:

- Slack incoming webhooks
- Discord webhooks
- Custom notification endpoints
- Ntfy.sh
- Gotify

## Event Types

### Import Events

| Event | Description |
|-------|-------------|
| Import Complete | Sent when a Strava sync finishes |

### Achievement Notifications

Detected during import when new personal records are set:

| Event | Description |
|-------|-------------|
| Personal Records | Best effort PRs (5K, 10K, etc.) |
| Segment PRs | New segment personal records |
| Eddington Increase | Eddington number goes up |
| Power Records | New peak power records (5s, 1min, 5min, 20min, 1hr) |
| Goal Complete | Training goal reaches 100% |
| Gear Milestones | Gear distance milestones (every 5000 km) |

### Training Load Alerts

Based on fitness/fatigue calculations:

| Event | Condition | Description |
|-------|-----------|-------------|
| Fatigue Warning | TSB < -20 | High accumulated fatigue |
| Recovery Alert | TSB > +10 | Well-rested and ready to train |
| Overtraining Risk | ATL > CTL × 1.5 | Acute load significantly exceeds chronic |

See [Calculations](../calculations.md) for details on TSB, ATL, and CTL.

### Maintenance Reminders

| Event | Description |
|-------|-------------|
| Maintenance Due | Component maintenance thresholds reached |

Configure check frequency:
- **Weekly** - Check every Monday
- **Monthly** - Check on the 1st of each month

### Digest Summaries

| Event | Description |
|-------|-------------|
| Weekly Digest | Activity summary for the past 7 days |
| Monthly Digest | Activity summary for the previous month |

## Configuration via API

Notifications are stored in athlete settings. The configuration structure:

```json
{
  "enabled": true,
  "services": [
    {
      "id": "uuid-here",
      "type": "telegram",
      "name": "My Telegram",
      "enabled": true,
      "config": {
        "token": "123456789:ABC...",
        "chat_id": "-1001234567890"
      }
    }
  ],
  "events": {
    "importComplete": true,
    "maintenanceDue": true,
    "maintenanceSchedule": "weekly",
    "personalRecords": true,
    "segmentPRs": false,
    "eddingtonIncrease": true,
    "powerRecords": true,
    "goalComplete": true,
    "fatigueWarning": true,
    "recoveryAlert": true,
    "overtrainingRisk": true,
    "weeklyDigest": true,
    "monthlyDigest": false,
    "gearMilestones": true
  }
}
```

## Testing Notifications

Use the "Test" button next to each configured service to send a test message. This verifies:

- Service configuration is valid
- Credentials are correct
- Network connectivity works

## Troubleshooting

### Telegram not receiving messages

1. Ensure the bot token is correct (contains a colon)
2. Verify you've started a conversation with the bot
3. Check the chat ID is correct (groups use negative IDs)
4. For groups, ensure the bot is added as a member

### SMTP failures

1. Verify host and port are correct
2. Check username/password
3. For Gmail, use an App Password
4. Some providers require "less secure app access"

### Webhook not triggering

1. Verify the URL is accessible
2. Check if the endpoint expects specific headers
3. Review server logs for incoming requests
