#!/bin/bash
# Publish all agents to the local agentregistry server (localhost:12121).
# Uses the canonical source at /Users/8bit/agents/agents/src/ via the
# /v0/apply endpoint (multi-doc YAML apply).
#
# Preferred: ./publish-agents-public.py (Python, handles YAML edge cases)
# Legacy: this script (uses /v0/agents POST per agent)

set -euo pipefail

REGISTRY="${REGISTRY:-http://localhost:12121}"
SRC="${SRC:-/Users/8bit/agents/agents/src}"
SKIP_FILES="package.json tsconfig.json"

count=0
ok=0
err=0

echo "=== Publishing agents to agentregistry ($REGISTRY) ==="
echo "  Source: $SRC"
echo ""

# Clean up any leftover entries from previous failed attempts
for prev in $(curl -s "$REGISTRY/v0/agents" | jq -r '.items[]? | .metadata.name' 2>/dev/null | sort -u); do
  echo "  [cleanup] removing existing: $prev"
  curl -s -X DELETE "$REGISTRY/v0/agents/$prev/stable" > /dev/null 2>&1 || true
done

for f in "$SRC"/*.json; do
  base=$(basename "$f" .json)

  # Skip special files
  skip=0
  for s in $SKIP_FILES; do
    if [ "$base" = "$s" ]; then skip=1; break; fi
  done
  [ "$skip" = 1 ] && continue

  count=$((count + 1))

  # Extract fields from agent JSON (cheshire-terminal format)
  name=$(jq -r '.name // .identifier // "'"$base"'"' "$f")
  title=$(jq -r '.meta.title // .displayName // .name // "'"$name"'"' "$f")
  desc=$(jq -r '.meta.description // .description // ""' "$f")
  ver=$(jq -r '.version // "1.0.0"' "$f")
  modelProvider=$(jq -r '.config.model // "clawdrouter/auto"' "$f" | cut -d/ -f1)
  modelName=$(jq -r '.config.model // "auto"' "$f" | cut -d/ -f2-)
  [ -z "$modelName" ] && modelName="auto"
  repo=$(jq -r '.metadata.github // ""' "$f")
  framework=$(jq -r '.meta.category // "agent"' "$f")
  image=$(jq -r '.registry.program_id // ""' "$f")
  [ -z "$image" ] && image="ghcr.io/solizardking/clawd-agents/$name:$ver"

  # Build payload in agentregistry native format
  payload=$(jq -n \
    --arg name "$name" \
    --arg version "$ver" \
    --arg title "$title" \
    --arg description "$desc" \
    --arg modelProvider "$modelProvider" \
    --arg modelName "$modelName" \
    --arg image "$image" \
    --arg language "typescript" \
    --arg framework "$framework" \
    --arg repository "$repo" \
    '{
      name: $name,
      version: $version,
      title: $title,
      description: $description,
      modelProvider: $modelProvider,
      modelName: $modelName,
      image: $image,
      language: $language,
      framework: $framework,
      repository: {url: $repository}
    }')

  echo "[$count] $name v$ver ..."

  resp=$(curl -s -w "%{http_code}" -o /tmp/ar-pub-resp.json \
    -X POST "$REGISTRY/v0/agents" \
    -H 'Content-Type: application/json' \
    -d "$payload" 2>&1)
  code="${resp: -3}"

  if [ "$code" = "200" ] || [ "$code" = "201" ]; then
    echo "  ✓ Created/updated"
    ok=$((ok + 1))
  else
    msg=$(cat /tmp/ar-pub-resp.json 2>/dev/null | jq -r '.detail // .title // "unknown error"' 2>/dev/null)
    echo "  ✗ HTTP $code: $msg"
    err=$((err + 1))
  fi
done

echo ""
echo "=== Summary ==="
echo "  Total: $count"
echo "  OK:    $ok"
echo "  Error: $err"