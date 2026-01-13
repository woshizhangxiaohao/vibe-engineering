#!/bin/sh
set -e

echo "=== Starting Vibe Backend Server ==="
echo "PORT: ${PORT:-8080}"
echo "ENV: ${ENV:-production}"
echo "DATABASE_URL: ${DATABASE_URL:0:30}..."
echo "REDIS_URL: ${REDIS_URL:0:20}..."
echo "====================================="

exec /server "$@"
