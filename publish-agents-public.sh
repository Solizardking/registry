#!/bin/bash
# Publish all agents from /Users/8bit/agents/agents/src/ to the live agentregistry.
# Also rebuilds the static public API (catalog, registry, ACP, .well-known).
#
# Usage: ./publish-agents-public.sh
#   REGISTRY=http://localhost:12121 (default)
#   SRC=/Users/8bit/agents/agents/src (default)

set -euo pipefail

REGISTRY="${REGISTRY:-http://localhost:12121}"
AGENTS_DIR="/Users/8bit/agents/agents"
SRC="${SRC:-${AGENTS_DIR}/src}"
SKIP_FILES="package tsconfig"

echo "=== Step 1: Rebuild static public API (catalog + registry + ACP) ==="
cd "$AGENTS_DIR"
node build-catalog.cjs
echo ""

echo "=== Step 2: Publish all agents to agentregistry ($REGISTRY) via /v0/apply ==="
echo ""

TMPDIR=$(mktemp -d)
YAML_FILE="$TMPDIR/agents.yaml"

count=0
skipped=0

for f in "$SRC"/*.json; do
  base=$(basename "$f" .json)

  # Skip special files
  skip=0
  for s in $SKIP_FILES; do
    if [ "$base" = "$s" ]; then skip=1; break; fi
  done
  [ "$skip" = 1 ] && { skipped=$((skipped + 1)); continue; }

  count=$((count + 1))

  # Extract fields from agent JSON (cheshire-terminal / Clawd format)
  name=$(jq -r '.name // .identifier // "'"$base"'"' "$f")
  title=$(jq -r '.meta.title // .displayName // .name // "'"$name"'"' "$f")
  desc=$(jq -r '.meta.description // .description // ""' "$f")
  # Convert multiline descriptions to YAML pipe-syntax (|+ for trailing newline)
  if echo "$desc" | grep -q -m1 . && [ "$(echo "$desc" | wc -l)" -gt 1 ]; then
    desc_yaml="|"$'\n'"$(echo "$desc" | sed 's/^/  /')"
  else
    desc_yaml="$desc"
  fi
  ver=$(jq -r '.version // "1.0.0"' "$f")
  author=$(jq -r '.author // ""' "$f")
  modelProvider=$(jq -r '.config.model // "clawdrouter/auto"' "$f")
  modelName=$(jq -r '.config.model // "auto"' "$f")
  homepage=$(jq -r '.homepage // ""' "$f")
  repo=$(jq -r '.metadata.github // ""' "$f")
  category=$(jq -r '.meta.category // "defi"' "$f")

  # Determine image/container reference
  image=$(jq -r '.registry.program_id // ""' "$f")
  [ -z "$image" ] && image="ghcr.io/solizardking/clawd-agents/$name:$ver"

  # Build tags from meta.tags (sanitize for YAML key compatibility)
  tags_json=$(jq -r '.meta.tags // []' "$f")
  tags_yaml=""
  while IFS= read -r tag; do
    if [ -n "$tag" ]; then
      # Replace @ and / chars which are invalid in YAML keys
      sanitized=$(echo "$tag" | sed 's/@/at-/g; s/\//-/g; s/:/--/g; s/[^a-zA-Z0-9_.-]/-/g')
      tags_yaml+="      tag.${sanitized}: \"true\""$'\n'
    fi
  done < <(echo "$tags_json" | jq -r '.[]')

  # Set repository URL
  if [ -z "$repo" ] || [ "$repo" = "null" ]; then
    repo="https://github.com/Solizardking/cheshire-terminal"
  fi

  # Append YAML document for this agent (using quoted scalars for safety)
  cat >> "$YAML_FILE" <<YAMLDOC
---
apiVersion: ar.dev/v1alpha1
kind: Agent
metadata:
  name: "$name"
  labels:
    app: "cheshire-terminal"
    catalog: "cheshire-terminal-agents"
    category: "$category"
$tags_yaml
  tag: "stable"
spec:
  title: "$title"
  description: "$desc"
  version: "$ver"
  modelProvider: "$modelProvider"
  modelName: "$modelName"
  source:
    image: "$image"
    repository:
      url: "$repo"
      subfolder: "agents/$name.json"

YAMLDOC

  echo "  [$count] $name v$ver"
done

echo ""
echo "Applying $count agents to $REGISTRY/v0/apply ..."

# Apply as multi-doc YAML
resp=$(curl -s -w "%{http_code}" -o "$TMPDIR/apply-resp.json" \
  -X POST "$REGISTRY/v0/apply" \
  -H 'Content-Type: application/yaml' \
  --data-binary "@$YAML_FILE" 2>&1)
code="${resp: -3}"

if [ "$code" = "200" ] || [ "$code" = "201" ]; then
  echo "  ✓ Apply succeeded (HTTP $code)"
  cat "$TMPDIR/apply-resp.json" | python3 -m json.tool 2>/dev/null | head -20
else
  msg=$(cat "$TMPDIR/apply-resp.json" 2>/dev/null)
  echo "  ✗ HTTP $code: $msg"
fi

# Cleanup
rm -rf "$TMPDIR"

echo ""
echo "=== Summary ==="
echo "  Total:   $count"
echo "  Skipped: $skipped"
echo ""
echo "=== Static API files ==="
echo "  Catalog:  $AGENTS_DIR/public/api/agents/agents-catalog.json"
echo "  ACP:      $AGENTS_DIR/public/api/agents/acp-registry.json"
echo "  Index:    $AGENTS_DIR/public/api/agents/index.json"
echo "  .well-known/acp.json: $AGENTS_DIR/public/.well-known/acp.json"
echo "  .well-known/ai-plugin.json: $AGENTS_DIR/public/.well-known/ai-plugin.json"
echo ""
echo "Done."