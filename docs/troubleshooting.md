# Troubleshooting

Common issues and their solutions.

## Authentication Issues

### "Not authenticated with Strava"

**Cause:** No valid Strava token stored.

**Solution:**
1. Start the server: `quantlete serve`
2. Open http://localhost:8081
3. Click "Connect with Strava"
4. Complete the OAuth flow

### "Invalid client" error

**Cause:** Incorrect Strava API credentials.

**Solution:**
1. Verify `QUANTLETE_STRAVA_CLIENT_ID` is correct
2. Verify `QUANTLETE_STRAVA_CLIENT_SECRET` is correct
3. Check for extra spaces or newlines in values

### "Redirect URI mismatch"

**Cause:** OAuth callback URL doesn't match Strava app settings.

**Solution:**
1. Go to https://www.strava.com/settings/api
2. Check "Authorization Callback Domain"
3. For local development, set to `localhost`
4. For production, set to your domain (without `https://` or path)

### Token expired / refresh failed

**Cause:** Token refresh failed, possibly due to revoked access.

**Solution:**
1. Re-authenticate via the web interface
2. Click "Connect with Strava" again
3. Re-authorize the application

## Import Issues

### "Rate limit exceeded"

**Cause:** Hit Strava API rate limits.

**Solution:**
- **15-minute limit:** Wait 15 minutes and retry
- **Daily limit:** Wait until midnight UTC

Rate limits:
- 100 requests per 15 minutes
- 1,000 requests per day

?> **Tip:** Use `--skip-streams` and `--skip-segments` for faster imports that use fewer API calls.

### Import stuck or very slow

**Cause:** Large activity history or slow network.

**Solution:**
1. Check progress with debug logging: `QUANTLETE_LOG_LEVEL=debug quantlete import`
2. Import in phases using `--skip-*` flags
3. Interrupt with Ctrl+C and resume with `--resume`

### Missing activity data

**Cause:** Strava privacy settings or import skipped.

**Solution:**
1. Check Strava privacy settings for the activity
2. Re-import with `quantlete import --full`
3. Ensure streams weren't skipped: `quantlete import` (without `--skip-streams`)

## Server Issues

### "Address already in use"

**Cause:** Another process is using the port.

**Solution:**
```bash
# Find what's using port 8081
lsof -i :8081

# Use a different port
quantlete serve --port 8082
```

### Database locked

**Cause:** Multiple processes accessing the database.

**Solution:**
1. Ensure only one instance is running
2. Stop the server before running import
3. Wait for any background processes to finish

### "Failed to open database"

**Cause:** Permission issues or invalid path.

**Solution:**
```bash
# Check data directory exists and is writable
ls -la ./data

# Create if missing
mkdir -p ./data

# Check permissions
chmod 755 ./data
```

## Dashboard Issues

### Blank page or loading forever

**Cause:** JavaScript error or failed API request.

**Solution:**
1. Open browser developer tools (F12)
2. Check Console tab for errors
3. Check Network tab for failed requests
4. Clear browser cache and reload

### Activities not showing

**Cause:** Import incomplete or filter applied.

**Solution:**
1. Check import completed: `quantlete import`
2. Clear any filters in the UI
3. Check browser console for API errors

### Charts not loading

**Cause:** No stream data or JavaScript error.

**Solution:**
1. Ensure streams were imported (check activity detail)
2. Re-import with streams: `quantlete import --full`
3. Check browser console for errors

## Docker Issues

### Container exits immediately

**Cause:** Configuration error or crash.

**Solution:**
```bash
# Check logs
docker logs quantlete

# Run interactively to see errors
docker run -it ghcr.io/melonamin/quantlete:latest
```

### Volume permissions

**Cause:** Container user can't write to volume.

**Solution:**
```bash
# Fix ownership
docker run --rm -v quantlete-data:/data alpine chown -R 1000:1000 /data
```

### Can't connect to container

**Cause:** Port mapping or network issues.

**Solution:**
```bash
# Check container is running
docker ps

# Check port mapping
docker port quantlete

# Check container logs
docker logs quantlete
```

## Debug Logging

Enable debug logging for more information:

```bash
# Environment variable
export QUANTLETE_LOG_LEVEL=debug
quantlete serve

# Or in config file
log:
  level: "debug"
```

Debug logs show:
- API requests and responses
- Database queries
- Token refresh events
- Detailed error messages

## Getting Help

If you can't resolve an issue:

1. Enable debug logging and reproduce the issue
2. Check for existing issues: https://github.com/melonamin/quantlete/issues
3. Open a new issue with:
   - Quantlete version (`quantlete version`)
   - Operating system
   - Debug logs (sanitize tokens/secrets)
   - Steps to reproduce
