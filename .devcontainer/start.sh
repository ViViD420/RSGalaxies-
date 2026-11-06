#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

if ! curl -fsS http://127.0.0.1:8080/healthz >/dev/null 2>&1; then
  (cd server && nohup go run ./cmd/worldd > /tmp/rs-galaxies-world.log 2>&1 < /dev/null &)
fi

mkdir -p client/build/web
godot --headless --path client --export-release "Web" client/build/web/index.html
(cd client/build/web && nohup python3 -m http.server 8000 --bind 0.0.0.0 > /tmp/rs-galaxies-web.log 2>&1 < /dev/null &)

echo "RS Galaxies browser prototype: open forwarded port 8000."
echo "World API: forwarded port 8080."
