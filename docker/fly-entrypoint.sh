#!/bin/sh
set -eu
# Fly postgres attach injects DATABASE_URL; arctl-server expects AGENT_REGISTRY_DATABASE_URL.
if [ -n "${DATABASE_URL:-}" ] && [ -z "${AGENT_REGISTRY_DATABASE_URL:-}" ]; then
  export AGENT_REGISTRY_DATABASE_URL="$DATABASE_URL"
fi
if [ -z "${AGENT_REGISTRY_SERVER_ADDRESS:-}" ]; then
  export AGENT_REGISTRY_SERVER_ADDRESS=":8080"
fi
if [ "$#" -gt 0 ]; then
  exec "$@"
fi
exec /app/bin/arctl-server
