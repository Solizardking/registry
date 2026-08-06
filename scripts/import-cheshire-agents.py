#!/usr/bin/env python3
"""Import Cheshire Terminal agents catalog into AgentRegistry (ar.dev/v1alpha1).

Reads agents from a local cheshire-terminal/agents checkout (agents-catalog.json
plus per-agent JSON under src/, defi-agents/src/, public/api/agents/catalog/),
emits multi-doc YAML batches, and optionally applies them with arctl.

Examples:
  # Generate YAML only (default output under examples/cheshire-terminal-catalog/)
  python3 scripts/import-cheshire-agents.py \\
    --source /path/to/cheshire-terminal/agents

  # Generate + apply against a live registry
  python3 scripts/import-cheshire-agents.py \\
    --source /path/to/cheshire-terminal/agents \\
    --apply --registry-url http://localhost:12121
"""

from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
from pathlib import Path


DEFAULT_TAG = "stable"
DEFAULT_REPO = "https://github.com/Solizardking/agents"
DEFAULT_MODEL = "clawdrouter/auto"
BATCH_SIZE = 25


def load_json(path: Path):
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:  # noqa: BLE001 — surface parse failures, keep going
        print(f"warn: cannot parse {path}: {exc}", file=sys.stderr)
        return None


def sanitize_name(name: str) -> str:
    name = name.lower().strip()
    name = re.sub(r"[^a-z0-9.-]+", "-", name)
    name = re.sub(r"-+", "-", name).strip("-")
    if not name:
        name = "unnamed-agent"
    if not re.match(r"^[a-z0-9]", name):
        name = "a-" + name
    return name[:63].rstrip("-.")


def parse_model(model: str) -> tuple[str, str]:
    model = (model or DEFAULT_MODEL).strip()
    if "/" in model:
        provider, name = model.split("/", 1)
        return provider, name
    return "clawdrouter", model


def yaml_escape(value) -> str:
    if value is None:
        return '""'
    text = str(value)
    if "\n" in text or text.strip() != text or any(c in text for c in ":\"'#{}[]&*!|>%@`"):
        return json.dumps(text, ensure_ascii=False)
    return text


def find_agent_json(root: Path, ident: str) -> tuple[Path | None, dict | None]:
    candidates = [
        root / "src" / f"{ident}.json",
        root / "defi-agents" / "src" / f"{ident}.json",
        root / "plugin.delivery" / "src" / f"{ident}.json",
        root / "public" / "api" / "agents" / "catalog" / f"{ident}.json",
        root / "public" / "api" / "agents" / "registry" / f"{ident}.json",
    ]
    for path in candidates:
        if path.exists():
            return path, load_json(path)
    for path in root.rglob(f"{ident}.json"):
        if any(part in {"locales", "node_modules", "target", "dist"} for part in path.parts):
            continue
        return path, load_json(path)
    return None, None


def labels_from(category, tags) -> dict[str, str]:
    labs = {
        "app": "cheshire-terminal",
        "catalog": "cheshire-terminal-agents",
    }
    if category:
        labs["category"] = sanitize_name(str(category))[:63]
    for tag in (tags or [])[:5]:
        labs["tag." + sanitize_name(str(tag))[:40]] = "true"
    return labs


def agent_to_yaml_doc(
    ident: str,
    body: dict,
    catalog_entry: dict,
    *,
    tag: str,
    repo_url: str,
) -> str:
    meta = body.get("meta") or {}
    title = (
        meta.get("title")
        or catalog_entry.get("title")
        or body.get("displayName")
        or body.get("name")
        or ident
    )
    desc = (
        meta.get("description")
        or catalog_entry.get("description")
        or body.get("description")
        or body.get("summary")
        or title
    )
    if isinstance(desc, str) and len(desc) > 2000:
        desc = desc[:1997] + "..."

    cfg = body.get("config") or {}
    provider, model_name = parse_model(cfg.get("model") or DEFAULT_MODEL)
    category = catalog_entry.get("category") or meta.get("category")
    tags = catalog_entry.get("tags") or meta.get("tags") or []
    version = body.get("version") or body.get("schemaVersion") or "1.0.0"
    name = sanitize_name(ident)

    image = body.get("image") or body.get("containerImage") or cfg.get("image")
    if not image:
        image = f"ghcr.io/solizardking/clawd-agents/{name}:{version}"

    labs = labels_from(category, tags)
    lines = [
        "apiVersion: ar.dev/v1alpha1",
        "kind: Agent",
        "metadata:",
        f"  name: {name}",
        f"  tag: {tag}",
        "  labels:",
    ]
    for key, value in sorted(labs.items()):
        lines.append(f"    {key}: {yaml_escape(value)}")
    lines.extend(
        [
            "spec:",
            f"  title: {yaml_escape(title)}",
            f"  description: {yaml_escape(desc)}",
            f"  modelProvider: {yaml_escape(provider)}",
            f"  modelName: {yaml_escape(model_name)}",
            "  source:",
            f"    image: {yaml_escape(image)}",
            "    repository:",
            f"      url: {yaml_escape(repo_url)}",
        ]
    )
    src_rel = body.get("_source_rel")
    if src_rel and src_rel.endswith(".json"):
        parent = str(Path(src_rel).parent)
        if parent and parent != ".":
            lines.append(f"      subfolder: {yaml_escape(parent)}")
    return "\n".join(lines)


def convert(source: Path, out_dir: Path, *, tag: str, repo_url: str) -> dict:
    catalog_path = source / "agents-catalog.json"
    catalog = load_json(catalog_path)
    if not catalog or "agents" not in catalog:
        raise SystemExit(f"missing or invalid catalog at {catalog_path}")

    entries = {a["identifier"]: a for a in catalog["agents"]}
    docs: list[str] = []
    resolved: list[str] = []
    catalog_only: list[str] = []

    for ident in sorted(entries):
        path, body = find_agent_json(source, ident)
        if body is None:
            body = {
                "identifier": ident,
                "meta": {
                    "title": entries[ident].get("title"),
                    "description": entries[ident].get("description"),
                    "tags": entries[ident].get("tags"),
                    "category": entries[ident].get("category"),
                },
                "config": {"model": DEFAULT_MODEL},
                "_source_rel": "agents-catalog.json",
            }
            catalog_only.append(ident)
        else:
            rel = str(path.relative_to(source)) if path else None
            body = dict(body)
            body["_source_rel"] = f"agents/{rel}" if rel else None
            resolved.append(ident)
        docs.append(
            agent_to_yaml_doc(ident, body, entries[ident], tag=tag, repo_url=repo_url)
        )

    out_dir.mkdir(parents=True, exist_ok=True)
    # Clear prior batches so re-runs don't leave stale files.
    for old in out_dir.glob("agents-batch-*.yaml"):
        old.unlink()

    batch_files: list[Path] = []
    for i in range(0, len(docs), BATCH_SIZE):
        chunk = docs[i : i + BATCH_SIZE]
        out = out_dir / f"agents-batch-{(i // BATCH_SIZE) + 1:02d}.yaml"
        out.write_text("\n---\n".join(chunk) + "\n", encoding="utf-8")
        batch_files.append(out)

    manifest = {
        "total": len(docs),
        "tag": tag,
        "resolvedFromJson": len(resolved),
        "catalogOnly": catalog_only,
        "batches": [p.name for p in batch_files],
        "source": str(source),
    }
    (out_dir / "import-manifest.json").write_text(
        json.dumps(manifest, indent=2) + "\n", encoding="utf-8"
    )
    return manifest


def apply_batches(out_dir: Path, registry_url: str, arctl: Path) -> int:
    batches = sorted(out_dir.glob("agents-batch-*.yaml"))
    if not batches:
        print("no batch files to apply", file=sys.stderr)
        return 1
    failed = 0
    for path in batches:
        print(f"--- apply {path.name} ---")
        proc = subprocess.run(
            [str(arctl), "apply", "-f", str(path), "--registry-url", registry_url],
            text=True,
        )
        if proc.returncode != 0:
            failed += 1
    return 1 if failed else 0


def main(argv: list[str] | None = None) -> int:
    repo_root = Path(__file__).resolve().parents[1]
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--source",
        type=Path,
        required=True,
        help="Path to cheshire-terminal/agents directory",
    )
    parser.add_argument(
        "--out",
        type=Path,
        default=repo_root / "examples" / "cheshire-terminal-catalog",
        help="Output directory for generated YAML batches",
    )
    parser.add_argument("--tag", default=DEFAULT_TAG, help="Agent tag (default: stable)")
    parser.add_argument(
        "--repo-url",
        default=DEFAULT_REPO,
        help="Repository URL written into Agent.spec.source.repository",
    )
    parser.add_argument(
        "--apply",
        action="store_true",
        help="Run arctl apply on generated batches against --registry-url",
    )
    parser.add_argument(
        "--registry-url",
        default="http://localhost:12121",
        help="Registry URL for --apply",
    )
    parser.add_argument(
        "--arctl",
        type=Path,
        default=repo_root / "bin" / "arctl",
        help="Path to arctl binary",
    )
    args = parser.parse_args(argv)

    source = args.source.expanduser().resolve()
    if not source.is_dir():
        raise SystemExit(f"source directory not found: {source}")

    manifest = convert(source, args.out, tag=args.tag, repo_url=args.repo_url)
    print(
        f"wrote {manifest['total']} agents in {len(manifest['batches'])} batches "
        f"to {args.out} (resolved={manifest['resolvedFromJson']}, "
        f"catalogOnly={len(manifest['catalogOnly'])})"
    )

    if args.apply:
        if not args.arctl.exists():
            raise SystemExit(f"arctl not found at {args.arctl}; build with make build-cli")
        return apply_batches(args.out, args.registry_url, args.arctl)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
