#!/bin/bash
# B-B Deploy Script (Linux / macOS / WSL)
set -e

RED='\033[0;31m'
GRN='\033[0;32m'
YEL='\033[0;33m'
NC='\033[0m' # No Color

log()  { echo -e "${GRN}[$1]${NC} $2"; }
warn() { echo -e "${YEL}[$1]${NC} $2"; }
err()  { echo -e "${RED}[$1]${NC} $2"; }

echo -e "${GRN}=== B-B Deploy ===${NC}\n"

# ---- 1. Check prerequisites ----
log "1/6" "Checking environment..."

command -v docker >/dev/null 2>&1 || { err "ERR" "docker not found"; exit 1; }
docker compose version >/dev/null 2>&1 || { err "ERR" "docker compose not found"; exit 1; }

# ---- 2. Create .env if missing ----
if [ ! -f ".env" ]; then
  if [ -f ".env.example" ]; then
    log "2/6" "Creating .env from .env.example..."
    cp .env.example .env
    warn "TIP" "Edit .env to customize secrets, then re-run deploy.sh"
    exit 0
  fi
fi

# ---- 3. Generate SSL certificate ----
if [ ! -f "nginx/ssl/server.crt" ]; then
  log "3/6" "Generating self-signed SSL certificate..."
  mkdir -p nginx/ssl
  openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout nginx/ssl/server.key \
    -out nginx/ssl/server.crt \
    -subj "/C=CN/ST=Shanghai/L=Shanghai/O=B-B/OU=Dev/CN=localhost" 2>/dev/null
  log "3/6" "Certificate generated"
else
  log "3/6" "SSL certificate already exists, skipping"
fi

# ---- 4. Check Docker Hub ----
log "4/6" "Checking Docker Hub connectivity..."
if ! docker pull alpine:3.19 --quiet 2>/dev/null; then
  warn "NET" "Cannot reach Docker Hub (registry-1.docker.io)"
  echo ""
  echo "  If you are in China, configure a Docker mirror:"
  echo "    Linux:   edit /etc/docker/daemon.json"
  echo "    Windows: Docker Desktop Settings > Docker Engine"
  echo ""
  echo '  Add: { "registry-mirrors": ["https://docker.m.daocloud.io"] }'
  echo "  Then restart Docker and re-run this script."
  echo ""
  exit 1
fi

# ---- 5. Build & start ----
log "5/6" "Building images..."
docker compose build

log "5/6" "Starting services..."
docker compose up -d

# ---- 6. Health check ----
log "6/6" "Waiting for backend..."
for i in $(seq 1 30); do
  if curl -sk https://localhost/api/v1/categories/ >/dev/null 2>&1; then
    log "6/6" "Backend is ready!"
    break
  fi
  sleep 2
done

echo ""
echo -e "${GRN}=== Deploy Complete ===${NC}"
echo "  Frontend: https://localhost"
echo "  API:      https://localhost/api/v1"
echo "  MinIO:    http://localhost:9001"
echo ""
