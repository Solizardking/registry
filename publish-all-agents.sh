#!/bin/bash
# Publish all Clawd agents to the local agentregistry server.
# Each agent JSON file in /Users/8bit/Clawd-agents/src/ is transformed
# and POSTed to the registry's /v0/agents endpoint.

set -euo pipefail

REGISTRY="http://localhost:12121"
SRC="/Users/8bit/Clawd-agents/src"
SKIP_FILES="package.json"

count=0
ok=0
err=0

echo "=== Publishing Clawd agents to agentregistry ($REGISTRY) ==="
echo ""

# Clean up any leftover docker containers from previous failed attempts
for prev in $(curl -s "$REGISTRY/v0/agents" | jq -r '.agents[]? | .name' 2>/dev/null); do
  ver=$(curl -s "$REGISTRY/v0/agents/$prev/versions/latest" | jq -r '.version // empty' 2>/dev/null)
  if [ -n "$ver" ]; then
    echo "  [cleanup] removing existing: $prev v$ver"
    curl -s -X DELETE "$REGISTRY/v0/agents/$prev/versions/$ver" > /dev/null 2>&1 || true
  fi
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

  # Extract fields from Clawd agent JSON
  name=$(jq -r '.name // .identifier // "'"$base"'"' "$f")
  title=$(jq -r '.meta.title // .displayName // .name // "'"$name"'"' "$f")
  desc=$(jq -r '.description // .meta.description // ""' "$f")
  ver=$(jq -r '.version // "1.0.0"' "$f")
  author=$(jq -r '.author // ""' "$f")
  modelProvider=$(jq -r '.config.model // "clawdrouter/auto"' "$f")
  modelName=$(jq -r '.config.model // "auto"' "$f")
  homepage=$(jq -r '.homepage // ""' "$f")
  repo=$(jq -r '.metadata.github // ""' "$f")

  # Map to fields the registry expects
  framework=$(jq -r '.metadata.category // "agent"' "$f")
  image=$(jq -r '.registry.program_id // ""' "$f")
  [ -z "$image" ] && image="ghcr.io/solizardking/clawd-agents/$name:$ver"

  # Build payload
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
    --arg websiteUrl "$homepage" \
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
      repository: {url: $repository},
      websiteUrl: $websiteUrl
    }')

  echo "[$count] $name v$ver ..."

  resp=$(curl -s -w "%{http_code}" -o /tmp/ar-pub-resp.json \
    -X POST "$REGISTRY/v0/agents" \
    -H 'Content-Type: application/json' \
    -d "$payload" 2>&1)
  code="${resp: -3}"

  if [ "$code" = "200" ]; then
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