# Strava API Setup

Quantlete needs Strava API credentials to access your activities.

## Create a Strava API Application

1. Go to [Strava API Settings](https://www.strava.com/settings/api)
2. Click **"Create App"** (or use an existing app)
3. Fill in the application details:

| Field | Value |
|-------|-------|
| Application Name | Quantlete (or any name) |
| Category | Visualizer |
| Website | http://localhost:8081 |
| Authorization Callback Domain | localhost |

4. Click **"Create"**

## Get Your Credentials

After creating the app, you'll see:

- **Client ID** - A numeric ID
- **Client Secret** - A long alphanumeric string

Copy these values - you'll need them to configure Quantlete.

## Configure Quantlete

Set the credentials as environment variables:

```bash
export QUANTLETE_STRAVA_CLIENT_ID=12345
export QUANTLETE_STRAVA_CLIENT_SECRET=abcdef123456...
```

Or in a config file (`quantlete.yaml`):

```yaml
strava:
  client_id: "12345"
  client_secret: "abcdef123456..."
```

Or via Docker:

```bash
docker run -d \
  -e QUANTLETE_STRAVA_CLIENT_ID=12345 \
  -e QUANTLETE_STRAVA_CLIENT_SECRET=abcdef123456... \
  ghcr.io/melonamin/quantlete:latest
```

## Webhook Setup (Optional)

For automatic syncing of new activities, configure webhooks. Quantlete does not auto-register Strava webhooks, so you must create the subscription manually and store its ID.

1. Your Quantlete instance must be accessible from the internet
2. Set the callback domain in Strava to your public URL
3. Configure the webhook verify token:

```bash
export QUANTLETE_STRAVA_WEBHOOK_VERIFY_TOKEN=your_random_16char_token
```

4. Create the Strava webhook subscription:

```bash
curl -X POST https://www.strava.com/api/v3/push_subscriptions \
  -F client_id=12345 \
  -F client_secret=abcdef123456... \
  -F callback_url=https://your-domain.com/api/v1/webhooks/strava \
  -F verify_token=your_random_16char_token
```

5. List subscriptions and copy the `id`:

```bash
curl -G https://www.strava.com/api/v3/push_subscriptions \
  -d client_id=12345 \
  -d client_secret=abcdef123456...
```

6. Store the subscription ID so Quantlete can validate events:

```bash
export QUANTLETE_STRAVA_WEBHOOK_SUBSCRIPTION_ID=123456
```

The webhook endpoint is: `https://your-domain.com/api/v1/webhooks/strava`

?> **Note:** Webhooks require a publicly accessible URL. For local development, you can use a tunnel service like ngrok.

?> **Note:** Strava allows only one webhook subscription per application. If you need to replace it, delete the old subscription first:

```bash
curl -X DELETE "https://www.strava.com/api/v3/push_subscriptions/123456?client_id=12345&client_secret=abcdef123456..."
```

## Rate Limits

Strava enforces API rate limits:

- **15-minute limit**: 100 requests per 15 minutes
- **Daily limit**: 1,000 requests per day

Quantlete tracks these limits and will pause imports when approaching limits. Rate limit state is persisted, so you can stop and resume imports safely.

## Troubleshooting

**"Invalid client" error**
- Double-check your Client ID and Client Secret
- Ensure there are no extra spaces in the values

**"Redirect URI mismatch"**
- The Authorization Callback Domain in Strava must match your server's domain
- For local development, use `localhost`

**Rate limit exceeded**
- Wait for the limit to reset (15 minutes or next day)
- Check `data/rate_limits.json` for current state

## Next Steps

- [Quick Start](quick-start.md) - Import your activities
- [Configuration](../configuration/README.md) - All configuration options
