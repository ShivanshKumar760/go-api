#!/usr/bin/env bash
# Runs ON THE DROPLET. The GitHub workflow calls it over SSH, and you can run it by hand too.
set -euo pipefail
cd /opt/blogapi

echo "building images and starting containers..."
docker compose up -d --build --remove-orphans

echo "waiting for the API to become healthy..."
for i in $(seq 1 30); do
  id=$(docker compose ps -q api)
  status=$(docker inspect --format '{{.State.Health.Status}}' "$id" 2>/dev/null || echo starting)
  if [ "$status" = "healthy" ]; then
    echo "API is healthy"
    docker image prune -f > /dev/null   # remove old, unused image layers
    exit 0
  fi
  sleep 2
done

echo "API did not become healthy. Last log lines:"
docker compose logs --tail=50 api
exit 1
