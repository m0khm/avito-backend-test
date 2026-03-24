#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "Waiting for ${BASE_URL} ..."
for _ in {1..60}; do
  if curl -fsS "${BASE_URL}/_info" >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

ADMIN_TOKEN="$(curl -fsS -X POST "${BASE_URL}/dummyLogin" -H 'Content-Type: application/json' -d '{"role":"admin"}' | python3 -c 'import json,sys; print(json.load(sys.stdin)["token"])')"

create_room() {
  local name="$1"
  local day="$2"
  local room_id
  room_id="$(curl -fsS -X POST "${BASE_URL}/rooms/create" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" \
    -H 'Content-Type: application/json' \
    -d "{\"name\":\"${name}\",\"capacity\":4}" | python3 -c 'import json,sys; print(json.load(sys.stdin)["room"]["id"])')"

  curl -fsS -X POST "${BASE_URL}/rooms/${room_id}/schedule/create" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" \
    -H 'Content-Type: application/json' \
    -d "{\"roomId\":\"${room_id}\",\"daysOfWeek\":[${day}],\"startTime\":\"09:00\",\"endTime\":\"18:00\"}" >/dev/null

  echo "Created room ${name} (${room_id})"
}

create_room "Alpha" 1
create_room "Beta" 3
create_room "Gamma" 5

echo "Seed completed"
