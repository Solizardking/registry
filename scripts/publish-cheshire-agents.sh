#!/bin/bash
# Publishes all cheshire-terminal agents as Prompts and Skills to AgentRegistry
set -euo pipefail

SRC_DIR="/Users/8bit/cheshire-terminal/agents/src"
REGISTRY_URL="${REGISTRY_URL:-https://agentregistry-1013652097839.us-central1.run.app}"
NAMESPACE="${NAMESPACE:-cheshire-terminal}"
COUNT=0
SKIPPED=0

for json_file in "$SRC_DIR"/*.json; do
  filename=$(basename "$json_file" .json)
  
  title=$(jq -r '.meta.title // empty' "$json_file")
  description=$(jq -r '.meta.description // empty' "$json_file")
  avatar=$(jq -r '.meta.avatar // ""' "$json_file")
  system_role=$(jq -r '.config.systemRole // empty' "$json_file")
  author=$(jq -r '.author // "cheshire-terminal"' "$json_file")
  
  if [ -z "$system_role" ]; then
    echo "⚠️  Skipping $filename — no systemRole"
    SKIPPED=$((SKIPPED + 1))
    continue
  fi
  
  # Build tags JSON array from the source file
  tags_json=$(jq -r '.meta.tags // [] | tojson' "$json_file" 2>/dev/null || echo "[]")
  
  echo "📄 [$((COUNT + 1))] Applying: $filename — $title..."
  
  # Build the payload safely using jq
  payload=$(jq -n \
    --arg kind "Prompt" \
    --arg name "$filename" \
    --arg ns "$NAMESPACE" \
    --arg title "$title" \
    --arg desc "$description" \
    --arg avatar "$avatar" \
    --arg author "$author" \
    --argjson tags "$tags_json" \
    --arg content "$system_role" \
    '{
      kind: $kind,
      apiVersion: "ar.dev/v1alpha1",
      metadata: { name: $name, namespace: $ns },
      spec: {
        displayName: $title,
        description: $desc,
        version: "1.0.0",
        metadata: { author: $author, avatar: $avatar, tags: $tags },
        content: $content
      }
    }' 2>/dev/null)
  
  if [ -z "$payload" ]; then
    echo "   ⚠️  $filename — failed to build Prompt payload"
    let SKIPPED=SKIPPED+1
    continue
  fi
  
  response=$(curl -s -X POST "$REGISTRY_URL/v0/apply?namespace=$NAMESPACE" \
    -H "Content-Type: application/json" \
    -d "$payload" 2>&1)
  
  if echo "$response" | jq -e '.results[0].status == "created" or .results[0].status == "updated"' > /dev/null 2>&1; then
    echo "   ✅ Prompt $filename — $(echo "$response" | jq -r '.results[0].status')"
    COUNT=$((COUNT + 1))
  else
    err=$(echo "$response" | jq -r '.results[0].error // "unknown"' 2>/dev/null)
    echo "   ⚠️  $filename — $err"
    SKIPPED=$((SKIPPED + 1))
    continue
  fi
  
  # Also publish as Skill (separate request since v0/apply doesn't batch well)
  skill_payload=$(jq -n \
    --arg kind "Skill" \
    --arg name "$filename" \
    --arg ns "$NAMESPACE" \
    --arg title "$title" \
    --arg desc "$description" \
    --arg avatar "$avatar" \
    --argjson tags "$tags_json" \
    '{
      kind: $kind,
      apiVersion: "ar.dev/v1alpha1",
      metadata: { name: $name, namespace: $ns },
      spec: {
        displayName: $title,
        description: $desc,
        version: "1.0.0",
        icon: $avatar,
        tags: $tags
      }
    }' 2>/dev/null)
  
  if [ -n "$skill_payload" ]; then
    skill_response=$(curl -s -X POST "$REGISTRY_URL/v0/apply?namespace=$NAMESPACE" \
      -H "Content-Type: application/json" \
      -d "$skill_payload" 2>&1)
    if echo "$skill_response" | jq -e '.results[0].status == "created" or .results[0].status == "updated"' > /dev/null 2>&1; then
      skill_status=$(echo "$skill_response" | jq -r '.results[0].status')
    fi
  fi
done

echo ""
echo "========================"
echo "Published: $COUNT agents"
echo "Skipped:   $SKIPPED"
echo "========================"