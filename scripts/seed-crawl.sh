#!/usr/bin/env bash
# Triggers the initial dnd.su seed crawl against a locally running stack.
# Run this on the server itself, from the project root: ./scripts/seed-crawl.sh
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8000}"
EMAIL="${EMAIL:-admin@example.com}"
PASSWORD="${PASSWORD:-changeme123}"

echo "==> Registering user ${EMAIL} (ignored if it already exists)"
curl -s -X POST "${BASE_URL}/api/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${EMAIL}\",\"password\":\"${PASSWORD}\"}" \
  -o /dev/null -w "register: %{http_code}\n"

echo "==> Logging in"
TOKEN=$(curl -s -X POST "${BASE_URL}/api/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${EMAIL}\",\"password\":\"${PASSWORD}\"}" | jq -r .token)

if [ -z "${TOKEN}" ] || [ "${TOKEN}" = "null" ]; then
  echo "Login failed — check EMAIL/PASSWORD or that the backend is up." >&2
  exit 1
fi

echo "==> Starting seed crawl"
curl -s -X POST "${BASE_URL}/api/admin/crawl/seed" \
  -H "Authorization: Bearer ${TOKEN}" \
  -w "\nseed: %{http_code}\n"

echo "==> Polling status every 30s (Ctrl+C to stop watching, the crawl keeps running server-side)"
while true; do
  STATUS=$(curl -s "${BASE_URL}/api/admin/crawl/status" -H "Authorization: Bearer ${TOKEN}")
  echo "$(date '+%H:%M:%S') ${STATUS}"
  echo "${STATUS}" | grep -q '"running":false' && break
  sleep 30
done

echo "==> Done"
