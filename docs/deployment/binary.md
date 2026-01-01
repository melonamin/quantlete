# Binary Deployment

Run Quantlete as a standalone binary without Docker.

## Download

Download the latest release for your platform:

```bash
# Linux (amd64)
curl -LO https://github.com/melonamin/quantlete/releases/latest/download/quantlete-linux-amd64
chmod +x quantlete-linux-amd64
sudo mv quantlete-linux-amd64 /usr/local/bin/quantlete

# Linux (arm64)
curl -LO https://github.com/melonamin/quantlete/releases/latest/download/quantlete-linux-arm64
chmod +x quantlete-linux-arm64
sudo mv quantlete-linux-arm64 /usr/local/bin/quantlete
```

## Configuration

Create a config file or use environment variables.

### Using Environment Variables

```bash
export QUANTLETE_STRAVA_CLIENT_ID=12345
export QUANTLETE_STRAVA_CLIENT_SECRET=your_secret
export QUANTLETE_STORAGE_DATA_DIR=/var/lib/quantlete
```

### Using Config File

Create `/etc/quantlete/quantlete.yaml`:

```yaml
strava:
  client_id: "12345"
  client_secret: "your_secret"

storage:
  data_dir: "/var/lib/quantlete"

log:
  level: "info"
```

## Systemd Service

Create a systemd service for automatic startup:

### Create Service User

```bash
sudo useradd -r -s /bin/false quantlete
sudo mkdir -p /var/lib/quantlete
sudo chown quantlete:quantlete /var/lib/quantlete
```

### Create Service File

Create `/etc/systemd/system/quantlete.service`:

```ini
[Unit]
Description=Quantlete Analytics Server
After=network.target

[Service]
Type=simple
User=quantlete
Group=quantlete
ExecStart=/usr/local/bin/quantlete serve
Restart=on-failure
RestartSec=5

# Environment
Environment=QUANTLETE_STORAGE_DATA_DIR=/var/lib/quantlete
EnvironmentFile=-/etc/quantlete/env

# Security
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/quantlete

[Install]
WantedBy=multi-user.target
```

### Create Environment File

Create `/etc/quantlete/env`:

```bash
QUANTLETE_STRAVA_CLIENT_ID=12345
QUANTLETE_STRAVA_CLIENT_SECRET=your_secret_here
```

Secure the file:

```bash
sudo chmod 600 /etc/quantlete/env
sudo chown quantlete:quantlete /etc/quantlete/env
```

### Enable Service

```bash
sudo systemctl daemon-reload
sudo systemctl enable quantlete
sudo systemctl start quantlete
```

### Check Status

```bash
sudo systemctl status quantlete
sudo journalctl -u quantlete -f
```

## Reverse Proxy

### Nginx

```nginx
server {
    listen 80;
    server_name quantlete.example.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name quantlete.example.com;

    ssl_certificate /etc/letsencrypt/live/quantlete.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/quantlete.example.com/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Caddy

```
quantlete.example.com {
    reverse_proxy localhost:8081
}
```

## Running Import

```bash
# As the service user
sudo -u quantlete quantlete import

# Or with the environment file
sudo -u quantlete env $(cat /etc/quantlete/env | xargs) quantlete import
```

## Updating

```bash
# Stop service
sudo systemctl stop quantlete

# Download new version
curl -LO https://github.com/melonamin/quantlete/releases/latest/download/quantlete-linux-amd64
sudo mv quantlete-linux-amd64 /usr/local/bin/quantlete
sudo chmod +x /usr/local/bin/quantlete

# Start service
sudo systemctl start quantlete
```

## Backup

```bash
# Stop service for consistent backup
sudo systemctl stop quantlete

# Copy database
sudo cp /var/lib/quantlete/quantlete.db /backup/quantlete-$(date +%Y%m%d).db

# Start service
sudo systemctl start quantlete
```

## Logs

```bash
# View service logs
sudo journalctl -u quantlete

# Follow logs
sudo journalctl -u quantlete -f

# Last 100 lines
sudo journalctl -u quantlete -n 100
```
