# Docker Deployment

Docker is the recommended way to run Quantlete.

## Quick Start

```bash
docker run -d \
  --name quantlete \
  -p 8081:8081 \
  -v quantlete-data:/data \
  -e QUANTLETE_STRAVA_CLIENT_ID=your_client_id \
  -e QUANTLETE_STRAVA_CLIENT_SECRET=your_client_secret \
  ghcr.io/melonamin/quantlete:latest
```

Open http://localhost:8081

## Docker Compose

For production deployments, use Docker Compose:

```yaml
# docker-compose.yml
services:
  quantlete:
    image: ghcr.io/melonamin/quantlete:latest
    container_name: quantlete
    restart: unless-stopped
    ports:
      - "8081:8081"
    volumes:
      - quantlete-data:/data
    environment:
      - QUANTLETE_STRAVA_CLIENT_ID=${STRAVA_CLIENT_ID}
      - QUANTLETE_STRAVA_CLIENT_SECRET=${STRAVA_CLIENT_SECRET}
      - QUANTLETE_LOG_LEVEL=info

volumes:
  quantlete-data:
```

Create a `.env` file:

```bash
STRAVA_CLIENT_ID=12345
STRAVA_CLIENT_SECRET=your_secret_here
```

Start:

```bash
docker compose up -d
```

## With Reverse Proxy

### Traefik

```yaml
services:
  quantlete:
    image: ghcr.io/melonamin/quantlete:latest
    restart: unless-stopped
    volumes:
      - quantlete-data:/data
    environment:
      - QUANTLETE_STRAVA_CLIENT_ID=${STRAVA_CLIENT_ID}
      - QUANTLETE_STRAVA_CLIENT_SECRET=${STRAVA_CLIENT_SECRET}
      - QUANTLETE_STRAVA_REDIRECT_URI=https://quantlete.example.com/api/v1/auth/strava/callback
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.quantlete.rule=Host(`quantlete.example.com`)"
      - "traefik.http.routers.quantlete.tls.certresolver=letsencrypt"
      - "traefik.http.services.quantlete.loadbalancer.server.port=8081"

volumes:
  quantlete-data:
```

### Caddy

```yaml
services:
  quantlete:
    image: ghcr.io/melonamin/quantlete:latest
    restart: unless-stopped
    volumes:
      - quantlete-data:/data
    environment:
      - QUANTLETE_STRAVA_CLIENT_ID=${STRAVA_CLIENT_ID}
      - QUANTLETE_STRAVA_CLIENT_SECRET=${STRAVA_CLIENT_SECRET}
      - QUANTLETE_STRAVA_REDIRECT_URI=https://quantlete.example.com/api/v1/auth/strava/callback
    networks:
      - caddy

  caddy:
    image: caddy:2
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
      - caddy-data:/data
    networks:
      - caddy

networks:
  caddy:

volumes:
  quantlete-data:
  caddy-data:
```

Caddyfile:

```
quantlete.example.com {
    reverse_proxy quantlete:8081
}
```

## Running Import

Execute the import command inside the container:

```bash
docker exec quantlete quantlete import
```

Or with Docker Compose:

```bash
docker compose exec quantlete quantlete import
```

## Updating

```bash
# Pull latest image
docker pull ghcr.io/melonamin/quantlete:latest

# Restart container
docker compose down
docker compose up -d
```

## Backup

Back up the data volume:

```bash
# Stop container
docker compose stop

# Copy database
docker run --rm \
  -v quantlete-data:/data \
  -v $(pwd):/backup \
  alpine cp /data/quantlete.db /backup/quantlete-backup.db

# Restart
docker compose start
```

## Logs

View logs:

```bash
docker logs quantlete

# Follow logs
docker logs -f quantlete

# With Docker Compose
docker compose logs -f quantlete
```
