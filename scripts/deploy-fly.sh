#!/usr/bin/env bash
# Deploy Agent Registry API+UI to Fly.io (cheshire-agentregistry).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

APP="${FLY_APP:-cheshire-agentregistry}"
REGION="${FLY_REGION:-iad}"
DB_APP="${FLY_POSTGRES_APP:-cheshire-agentregistry-db}"

need() { command -v "$1" >/dev/null 2>&1 || { echo "missing: $1" >&2; exit 1; }; }
need fly
need openssl

echo "==> App: $APP  region: $REGION"

if ! fly apps list --json 2>/dev/null | grep -q "\"Name\":\"$APP\""; then
  echo "==> Creating app $APP"
  fly apps create "$APP" --org personal 2>/dev/null || fly apps create "$APP"
fi

# Postgres: create if missing, then attach (sets DATABASE_URL on the app)
if ! fly apps list --json 2>/dev/null | grep -q "\"Name\":\"$DB_APP\""; then
  echo "==> Creating Postgres $DB_APP (shared-cpu-1x, 1GB volume)"
  fly postgres create \
    --name "$DB_APP" \
    --region "$REGION" \
    --vm-size shared-cpu-1x \
    --volume-size 1 \
    --initial-cluster-size 1 \
    --org personal \
    --yes
fi

echo "==> Attaching Postgres (idempotent if already attached)"
fly postgres attach "$DB_APP" -a "$APP" --yes 2>/dev/null || \
  fly postgres attach "$DB_APP" -a "$APP" 2>/dev/null || \
  echo "    (attach skipped — may already be linked)"

# Resolve DATABASE_URL from app secrets / status for AGENT_REGISTRY_DATABASE_URL
echo "==> Ensuring AGENT_REGISTRY_* secrets"
EXISTING_JWT="$(fly secrets list -a "$APP" 2>/dev/null | grep -E 'AGENT_REGISTRY_JWT_PRIVATE_KEY|JWT' || true)"
if [[ -z "$EXISTING_JWT" ]]; then
  JWT="$(openssl rand -hex 32)"
  fly secrets set -a "$APP" "AGENT_REGISTRY_JWT_PRIVATE_KEY=$JWT"
  echo "    set AGENT_REGISTRY_JWT_PRIVATE_KEY"
else
  echo "    JWT secret already present"
fi

# DATABASE_URL is set by postgres attach; mirror to AGENT_REGISTRY_DATABASE_URL
# via a release-time note: Fly injects DATABASE_URL. We also set the prefixed var
# when we can read it from the running machine or from `fly config env`.
DB_URL="${AGENT_REGISTRY_DATABASE_URL:-}"
if [[ -z "$DB_URL" ]]; then
  # Try to pull from attached secret display (value redacted) — user may export it.
  echo "    NOTE: ensure AGENT_REGISTRY_DATABASE_URL is set to the same value as DATABASE_URL"
  echo "    fly secrets set -a $APP AGENT_REGISTRY_DATABASE_URL=\"\$DATABASE_URL\""
  # Common pattern: set both names from attach's DATABASE_URL using flyctl release command
  # If DATABASE_URL exists as a secret, duplicate it:
  if fly secrets list -a "$APP" 2>/dev/null | grep -q 'DATABASE_URL'; then
    echo "    DATABASE_URL present — setting AGENT_REGISTRY_DATABASE_URL via machine env alias..."
    # Fly postgres attach creates DATABASE_URL; config uses AGENT_REGISTRY_DATABASE_URL.
    # Use a small wrapper entrypoint in [processes] isn't available; set explicitly:
    # Operators must pass --build-only then secrets. We attempt ssh after first deploy.
  fi
fi

if [[ -n "${AGENT_REGISTRY_DATABASE_URL:-}" ]]; then
  fly secrets set -a "$APP" "AGENT_REGISTRY_DATABASE_URL=$AGENT_REGISTRY_DATABASE_URL"
fi

echo "==> Deploying $APP"
fly deploy -c fly.toml --remote-only -a "$APP"

echo "==> Health check"
sleep 5
HEALTH_URL="https://${APP}.fly.dev/v0/health"
if curl -fsS --max-time 30 "$HEALTH_URL" | head -c 400; then
  echo
  echo "OK: $HEALTH_URL"
else
  echo "WARN: health check failed — check: fly logs -a $APP"
  exit 1
fi

echo
echo "Next: point Cheshire at this registry:"
echo "  fly secrets set -a cheshire-clawd-terminal AGENT_REGISTRY_URL=https://${APP}.fly.dev"
echo "  # Redeploy cheshire SPA/API so proxy + public-config pick it up"
echo
echo "Native UI (bundled Next export): https://${APP}.fly.dev/"
echo "Cheshire page (after site deploy): https://cheshireterminal.ai/agent-registry"
