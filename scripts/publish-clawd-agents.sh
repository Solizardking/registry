#!/bin/bash
# Bulk-publish all Clawd agents to the local agentregistry daemon.
# Idempotent: checks if each agent exists before publishing.
set -euo pipefail

CLAWD_SRC="${1:-/Users/8bit/Clawd-agents/src}"
BASE_URL="${2:-http://localhost:12121/v0}"

echo "=== Publishing Clawd agents to agentregistry ==="
echo "Source: $CLAWD_SRC"
echo "API: $BASE_URL"

python3 << 'PYEOF'
import json, requests, os, sys

base_url = "http://localhost:12121/v0"
src_dir = "/Users/8bit/Clawd-agents/src"

# Detect existing agents
existing = set()
r = requests.get(f"{base_url}/agents")
for item in r.json().get("agents", []):
    existing.add(item.get("agent", {}).get("name",""))

# Also check via cursor pagination for remaining
meta = r.json().get("metadata", {})
cursor = meta.get("nextCursor", "")
while cursor:
    r = requests.get(f"{base_url}/agents?cursor={cursor}")
    for item in r.json().get("agents", []):
        existing.add(item.get("agent", {}).get("name",""))
    meta = r.json().get("metadata", {})
    cursor = meta.get("nextCursor", "")

count = 0
for fname in sorted(os.listdir(src_dir)):
    if not fname.endswith(".json") or fname == "package.json":
        continue
    with open(os.path.join(src_dir, fname)) as f:
        d = json.load(f)
    
    name = d.get("identifier") or d.get("name") or fname.replace(".json","")
    if name in existing:
        continue
    
    version = d.get("version") or "1.0.0"
    meta = d.get("meta") or {}
    title = meta.get("title") or d.get("displayName","") or name
    description = meta.get("description") or d.get("description","")
    metadata = d.get("metadata") or {}
    repo_url = metadata.get("github") or d.get("homepage","")
    homepage = d.get("homepage","")
    config = d.get("config", {})
    model = config.get("model", "clawdrouter/auto")
    
    payload = {
        "name": name,
        "version": f"1.0.{count + 1}",
        "title": title,
        "description": description,
        "image": d.get("registry", {}).get("program_id", "") or f"ghcr.io/solizardking/clawd-agents/{name}:{version}",
        "language": "typescript",
        "framework": "agent",
        "modelProvider": model.split("/")[0] if "/" in model else model,
        "modelName": model,
    }
    if repo_url:
        payload["repository"] = {"url": repo_url}
    if homepage:
        payload["websiteUrl"] = homepage
    
    resp = requests.post(f"{base_url}/agents", json=payload, timeout=15)
    if resp.status_code in (200, 201, 204):
        count += 1
        existing.add(name)
        print(f"  [+] {name}")
    elif resp.status_code == 409:
        print(f"  [=] {name} (exists)")
        existing.add(name)
    else:
        print(f"  [FAIL] {name}: {resp.status_code} {resp.text[:120]}")

print(f"\nDone: {count} new agents published, {len(existing)} total in registry")
PYEOF</write_to_file>
</file>