# Server Deployment Guide

## Prerequisites on Server
1. Git installed
2. Docker and Docker Compose installed
3. Swap space configured (if RAM < 2GB)

## Initial Setup

### 1. Clone Repository
```bash
cd /home/chenaho
git clone https://github.com/YOUR_USERNAME/line-liff-event-manager.git
cd line-liff-event-manager
```

### 2. Create Environment File
```bash
# Copy the example
cp .env.example .env

# Edit with your actual values
nano .env
```

Set these values:
```env
JWT_SECRET=<generate using: openssl rand -base64 32>
ADMIN_LIST=U1234567890abcdef,U0987654321fedcba
```

### 3. Add Firebase Credentials
Copy your `firebase-key.json` to the project root:
```bash
# On your local machine, use scp
scp /path/to/firebase-key.json root@165.22.161.59:/home/chenaho/line-liff-event-manager/
```

Or create it manually:
```bash
nano firebase-key.json
# Paste the JSON content
```

### 4. DNS Configuration
Ensure `eventmanager.chenaho.com` points to your server IP (`165.22.161.59`):
```
A Record: eventmanager.chenaho.com -> 165.22.161.59
```

## Deploy

### First Time Deployment
```bash
docker compose -f docker-compose.prod.yml up -d --build
```

### Subsequent Deployments
```bash
# Pull latest code
git pull origin main

# Rebuild and restart
docker compose -f docker-compose.prod.yml down
docker compose -f docker-compose.prod.yml up -d --build
```

## Verify Deployment

### Check Services
```bash
# Check if containers are running
docker compose -f docker-compose.prod.yml ps

# View logs
docker compose -f docker-compose.prod.yml logs -f
```

### Test API
```bash
# Test HTTP (Caddy will redirect to HTTPS)
curl http://eventmanager.chenaho.com

# Test HTTPS (once certificate is issued)
curl https://eventmanager.chenaho.com/api/health
```

### Check Caddy Certificate
```bash
# View Caddy logs
docker compose -f docker-compose.prod.yml logs caddy

# Check certificate status
docker exec line-liff-event-manager-caddy-1 caddy list-certificates
```

## Troubleshooting

### Build Fails with OOM
Add swap space:
```bash
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
```

### Check Container Logs
```bash
# Backend logs
docker compose -f docker-compose.prod.yml logs app-backend

# Caddy logs
docker compose -f docker-compose.prod.yml logs caddy
```

### Restart Services
```bash
docker compose -f docker-compose.prod.yml restart
```

### Clean Rebuild
```bash
docker compose -f docker-compose.prod.yml down -v
docker system prune -a
docker compose -f docker-compose.prod.yml up -d --build
```

## File Checklist
- [ ] `.env` with JWT_SECRET and ADMIN_LIST
- [ ] `firebase-key.json` in project root
- [ ] DNS A record pointing to server IP
- [ ] Swap configured (if needed)
