#!/bin/bash
# B-B Deploy Script
# Usage: ./deploy.sh
set -e

echo "=== B-B Deploy ==="

# Check prerequisites
echo "[Check] docker..."
command -v docker >/dev/null 2>&1 || { echo "Error: docker not found"; exit 1; }
echo "[Check] docker compose..."
docker compose version >/dev/null 2>&1 || { echo "Error: docker compose not found"; exit 1; }

# Ensure .env exists
if [ ! -f "./.env" ]; then
  if [ -f "./.env.example" ]; then
    echo "[Env] Creating .env from .env.example..."
    cp .env.example .env
    echo "[Env] Edit .env to customize secrets, then re-run deploy.sh"
    exit 0
  else
    echo "[Env] Warning: no .env or .env.example found"
  fi
fi

# Generate self-signed SSL certificate if not exists
if [ ! -f "./nginx/ssl/server.crt" ]; then
  echo "[SSL] Generating self-signed certificate..."
  mkdir -p ./nginx/ssl
  openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout ./nginx/ssl/server.key \
    -out ./nginx/ssl/server.crt \
    -subj "/C=CN/ST=Shanghai/L=Shanghai/O=B-B/OU=Dev/CN=localhost"
  echo "[SSL] Certificate generated"
fi

# Build and start
echo "[Docker] Building images..."
docker compose build

echo "[Docker] Starting services..."
docker compose up -d

# Wait for services to be healthy
echo "[Health] Waiting for backend..."
for i in $(seq 1 30); do
  if curl -s -k https://localhost/api/v1/categories/ > /dev/null 2>&1; then
    echo "[Health] Backend is ready"
    break
  fi
  sleep 2
done

echo "=== Deploy Complete ==="
echo "Frontend: https://localhost"
echo "API:      https://localhost/api/v1"
echo "MinIO:    http://localhost:9001"
