#!/usr/bin/env bash
# Host SkillHub static site + ensure local agentregistry is serving skills.
#
# Usage:
#   ./scripts/host-skillhub-local.sh /path/to/skillhub-main
#   SKILLHUB_DIR=~/Downloads/skillhub-main ./scripts/host-skillhub-local.sh
#
# Local endpoints:
#   Site:     http://127.0.0.1:8787/
#   API:      http://127.0.0.1:8787/api/skills.json
#   Scanner:  http://127.0.0.1:8877/   (if scanner/public exists)
#   Registry: http://localhost:12121   (arctl-server)
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
HUB="${1:-${SKILLHUB_DIR:-}}"
if [[ -z "$HUB" ]]; then
  echo "Usage: $0 /path/to/skillhub-main" >&2
  exit 1
fi
HUB="$(cd "$HUB" && pwd)"
SITE_PORT="${SKILLHUB_SITE_PORT:-8787}"
SCAN_PORT="${SKILLHUB_SCAN_PORT:-8877}"

if [[ ! -f "$HUB/catalog.json" ]]; then
  echo "ERROR: $HUB/catalog.json missing" >&2
  exit 1
fi

echo "==> Building catalog in $HUB"
(cd "$HUB" && npm run build:catalog)

if [[ ! -f "$HUB/public/index.html" ]]; then
  echo "ERROR: build did not produce public/index.html" >&2
  exit 1
fi

if ! curl -sf "http://localhost:12121/v0/health" >/dev/null 2>&1; then
  echo "==> Local agentregistry not up on :12121"
  echo "    Start with: make build-server && ./bin/arctl-server  (or arctl daemon start)"
else
  echo "==> Registry healthy at http://localhost:12121"
  if [[ -x "$ROOT/scripts/import-skillhub.py" && -x "$ROOT/bin/arctl" ]]; then
    echo "==> Publishing skills into local registry (tag=skillhub)"
    python3 "$ROOT/scripts/import-skillhub.py" \
      --source "$HUB" \
      --repo-url https://github.com/Solizardking/skillhub-main \
      --apply --registry-url http://localhost:12121 \
      --arctl "$ROOT/bin/arctl" || true
  fi
fi

if lsof -ti ":$SITE_PORT" >/dev/null 2>&1; then
  echo "==> Port $SITE_PORT already in use — leaving existing process"
else
  echo "==> Serving SkillHub site on http://127.0.0.1:$SITE_PORT/"
  python3 -m http.server "$SITE_PORT" --bind 127.0.0.1 --directory "$HUB/public" &
  echo $! > /tmp/skillhub-local-site.pid
fi

if [[ -d "$HUB/scanner/public" ]]; then
  if lsof -ti ":$SCAN_PORT" >/dev/null 2>&1; then
    echo "==> Port $SCAN_PORT already in use — leaving existing scanner"
  else
    echo "==> Serving scanner on http://127.0.0.1:$SCAN_PORT/"
    python3 -m http.server "$SCAN_PORT" --bind 127.0.0.1 --directory "$HUB/scanner/public" &
    echo $! > /tmp/skillhub-scanner.pid
  fi
fi

echo
echo "Ready:"
echo "  Site:     http://127.0.0.1:$SITE_PORT/"
echo "  API:      http://127.0.0.1:$SITE_PORT/api/skills.json"
echo "  Registry: http://localhost:12121  (arctl get skills --tag skillhub)"
echo "  Prod:     https://skills.x402.wtf"
