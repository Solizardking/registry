#!/bin/bash
# Script to publish all Clawd agents to the local agentregistry
set -euo pipefail

REGISTRY_URL="http://localhost:12121"
CLAWD_AGENTS_DIR="/Users/8bit/Clawd-agents/src"
GIT_REPO="https://github.com/Solizardking/agents"

echo "Publishing all Clawd agents to agentregistry at $REGISTRY_URL..."
echo ""

count=0
success=0
failed=0

for json_file in "$CLAWD_AGENTS_DIR"/*.json; do
  basename=$(basename "$json_file" .json)
  
  # Skip package.json
  if [ "$basename" = "package" ]; then
    continue
  fi

  count=$((count + 1))
  
  # Extract fields from the Clawd agent JSON
  name=$(jq -r '.name // .identifier // "'"$basename"'"' "$json_file")
  display_name=$(jq -r '.displayName // .name // "'"$name"'"' "$json_file")
  description=$(jq -r '.description // .meta.description // ""' "$json_file")
  version=$(jq -r '.version // "1.0.0"' "$json_file")
  author=$(jq -r '.author // ""' "$json_file")
  model=$(jq -r '.config.model // "clawdrouter/auto"' "$json_file")
  system_role=$(jq -r '.config.systemRole // ""' "$json_file")
  
  # Build the agent payload in agentregistry v1alpha1 format
  # TODO: Actually need to check what format the /v0 POST endpoint accepts
  
  # Try publishing via arctl first (git-based direct registration)
  echo "[$count] Publishing: $name ($version)..."
  
  if arctl agent publish "$basename" \
    --git "$GIT_REPO" \
    --version "$version" \
    --description "$description" \
    --overwrite 2>/dev/null; then
    echo "  ✓ Published: $name"
    success=$((success + 1))
  else
    echo "  ✗ Failed: $name"
    failed=$((failed + 1))
  fi
  
  echo ""
done

echo "=========================================="
echo "Summary: $success published, $failed failed out of $count total"