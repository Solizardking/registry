#!/usr/bin/env python3
"""
Publish all agents from /Users/8bit/agents/agents/src/ to the live agentregistry.
Also rebuilds the static public API (catalog, registry, ACP, .well-known).

Usage: python3 publish-agents-public.py
  REGISTRY=http://localhost:12121 (default)
"""

import json
import os
import re
import subprocess
import sys
import tempfile
import urllib.request
import urllib.error

REGISTRY = os.environ.get("REGISTRY", "http://localhost:12121")
AGENTS_DIR = "/Users/8bit/agents/agents"
SRC_DIR = os.path.join(AGENTS_DIR, "src")
SKIP_BASENAMES = {"package", "tsconfig"}


def sanitize_yaml_key(s: str) -> str:
    """Sanitize a string to be a valid YAML key."""
    s = s.replace("@", "at-")
    s = s.replace("/", "-")
    s = s.replace(":", "--")
    s = re.sub(r"[^a-zA-Z0-9_.-]", "-", s)
    return s


def to_yaml_value(s: str, indent: int = 0) -> str:
    """Convert a string to a safe YAML value, handling multiline."""
    prefix = " " * indent
    if not s:
        return '""'
    if "\n" in s:
        # Pipe block scalar
        return "|\n" + "\n".join(f"{prefix}  {line}" for line in s.split("\n"))
    # Simple quoted string (escape double quotes and backslashes)
    escaped = s.replace("\\", "\\\\").replace('"', '\\"')
    return f'"{escaped}"'


def build_agent_yaml(agent: dict) -> str:
    """Build a YAML document for one agent."""
    name = agent.get("name") or agent.get("identifier") or "unknown"
    title = agent.get("meta", {}).get("title") or agent.get("displayName") or name
    desc = agent.get("meta", {}).get("description") or agent.get("description") or ""
    ver = agent.get("version") or "1.0.0"
    model_provider_full = agent.get("config", {}).get("model") or "clawdrouter/auto"
    repo_url = agent.get("metadata", {}).get("github") or "https://github.com/Solizardking/agents"
    category = agent.get("meta", {}).get("category") or "defi"
    tags = agent.get("meta", {}).get("tags") or []

    # Split model provider/name (format: "provider/name")
    if "/" in model_provider_full:
        model_provider, model_name = model_provider_full.split("/", 1)
    else:
        model_provider, model_name = model_provider_full, "auto"

    lines = []
    lines.append("---")
    lines.append('apiVersion: ar.dev/v1alpha1')
    lines.append('kind: Agent')
    lines.append('metadata:')
    lines.append(f'  name: "{name}"')
    lines.append('  labels:')
    lines.append('    app: "cheshire-terminal"')
    lines.append('    catalog: "cheshire-terminal-agents"')
    lines.append(f'    category: "{category}"')
    for tag in tags:
        if tag:
            key = sanitize_yaml_key(tag)
            lines.append(f'    tag.{key}: "true"')
    lines.append('    tag: "stable"')
    lines.append('spec:')
    lines.append(f'  title: "{title}"')
    # description - handle multiline
    if "\n" in desc:
        lines.append('  description: |')
        for line in desc.split("\n"):
            lines.append(f'    {line}')
    else:
        lines.append(f'  description: "{desc}"')
    lines.append(f'  version: "{ver}"')
    lines.append(f'  modelProvider: "{model_provider}"')
    lines.append(f'  modelName: "{model_name}"')
    lines.append('  source:')
    lines.append(f'    image: "{agent.get("registry", {}).get("program_id") or f"ghcr.io/solizardking/clawd-agents/{name}:{ver}"}"')
    lines.append('    repository:')
    lines.append(f'      url: "{repo_url}"')
    lines.append(f'      subfolder: "agents/{name}.json"')
    lines.append("")

    return "\n".join(lines)


def main():
    print("=== Step 1: Rebuild static public API (catalog + registry + ACP) ===")
    subprocess.run(
        ["node", "build-catalog.cjs"],
        cwd=AGENTS_DIR,
        check=True,
        capture_output=True,
        text=True,
    )
    print(f"✅ Wrote {AGENTS_DIR}/agents-catalog.json")
    print()

    print(f"=== Step 2: Publish all agents to agentregistry ({REGISTRY}) via /v0/apply ===")
    print()

    agents = []
    for fname in sorted(os.listdir(SRC_DIR)):
        if not fname.endswith(".json"):
            continue
        base = os.path.splitext(fname)[0]
        if base in SKIP_BASENAMES:
            continue

        with open(os.path.join(SRC_DIR, fname)) as f:
            try:
                agent = json.load(f)
            except json.JSONDecodeError as e:
                print(f"  ⚠ Skipping {fname}: invalid JSON ({e})")
                continue

        agents.append(agent)
        name = agent.get("name") or agent.get("identifier") or base
        ver = agent.get("version") or "1.0.0"
        print(f"  [{len(agents)}] {name} v{ver}")

    # Build multi-doc YAML
    yaml_parts = [build_agent_yaml(a) for a in agents]
    yaml_content = "\n".join(yaml_parts)

    if not yaml_content.strip():
        print("  ⚠ No agents to publish")
        return

    # Write temp file
    tmpdir = tempfile.mkdtemp()
    yaml_path = os.path.join(tmpdir, "agents.yaml")
    with open(yaml_path, "w") as f:
        f.write(yaml_content)

    print(f"\nApplying {len(agents)} agents to {REGISTRY}/v0/apply ...")

    # Read and POST
    with open(yaml_path, "rb") as f:
        data = f.read()

    req = urllib.request.Request(
        f"{REGISTRY}/v0/apply",
        data=data,
        headers={"Content-Type": "application/yaml"},
        method="POST",
    )
    try:
        with urllib.request.urlopen(req) as resp:
            body = resp.read().decode()
            print(f"  ✓ Apply succeeded (HTTP {resp.status})")
            result = json.loads(body)
            for r in result.get("results", []):
                status = r.get("status", "?")
                err = r.get("error", "")
                if err:
                    print(f"    - {status}: {err}")
                else:
                    print(f"    - {status}")
    except urllib.error.HTTPError as e:
        body = e.read().decode()
        print(f"  ✗ HTTP {e.code}: {body[:500]}")
    except Exception as e:
        print(f"  ✗ Error: {e}")
    finally:
        import shutil

        shutil.rmtree(tmpdir, ignore_errors=True)

    print()
    print("=== Summary ===")
    print(f"  Total:   {len(agents)}")
    print()
    print("=== Static API files ===")
    print(f"  Catalog:  {AGENTS_DIR}/public/api/agents/agents-catalog.json")
    print(f"  ACP:      {AGENTS_DIR}/public/api/agents/acp-registry.json")
    print(f"  Index:    {AGENTS_DIR}/public/api/agents/index.json")
    print(f"  .well-known/acp.json: {AGENTS_DIR}/public/.well-known/acp.json")
    print(f"  .well-known/ai-plugin.json: {AGENTS_DIR}/public/.well-known/ai-plugin.json")
    print()
    print("Done.")


if __name__ == "__main__":
    main()