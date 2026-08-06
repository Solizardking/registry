#!/usr/bin/env python3
"""Import Robinhood Agents forge catalog into AgentRegistry (ar.dev/v1alpha1).

Reads Cheshire-schema agent JSON files from a local robinhood-agents checkout
(`agents/*.json`) plus the dual-chain examples under `examples/`, emits multi-doc
YAML batches, and optionally applies them with arctl.

Uses tag ``robinhood`` by default so these rows coexist with Cheshire catalog
agents that share the same identifier under tag ``stable``.

Examples:
  python3 scripts/import-robinhood-agents.py \\
    --source /path/to/robinhood-agents

  python3 scripts/import-robinhood-agents.py \\
    --source /path/to/robinhood-agents \\
    --apply --registry-url http://localhost:12121
"""

from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
from pathlib import Path

DEFAULT_TAG = "robinhood"
DEFAULT_REPO = "https://github.com/Solizardking/agents"
DEFAULT_MODEL_PROVIDER = "clawdrouter"
DEFAULT_MODEL_NAME = "auto"
BATCH_SIZE = 25


def load_json(path: Path):
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:  # noqa: BLE001
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


def yaml_escape(value) -> str:
    if value is None:
        return '""'
    text = str(value)
    if "\n" in text or text.strip() != text or any(c in text for c in ":\"'#{}[]&*!|>%@`"):
        return json.dumps(text, ensure_ascii=False)
    return text


def labels_from(category, tags, *, character: bool = False) -> dict[str, str]:
    labs = {
        "app": "robinhood-agents",
        "catalog": "robinhood-agents",
        "forge": "cheshire-terminal",
    }
    if character:
        labs["kind"] = "character"
    if category:
        labs["category"] = sanitize_name(str(category))[:63]
    for tag in (tags or [])[:5]:
        labs["tag." + sanitize_name(str(tag))[:40]] = "true"
    return labs


def parse_model(cfg: dict) -> tuple[str, str]:
    model = (cfg or {}).get("model")
    if not model:
        return DEFAULT_MODEL_PROVIDER, DEFAULT_MODEL_NAME
    model = str(model).strip()
    if "/" in model:
        provider, name = model.split("/", 1)
        return provider, name
    return DEFAULT_MODEL_PROVIDER, model


def cheshire_agent_to_yaml(
    body: dict,
    *,
    source_rel: str,
    tag: str,
    repo_url: str,
) -> str | None:
    ident = body.get("identifier") or body.get("name")
    if not ident:
        return None
    meta = body.get("meta") or {}
    title = meta.get("title") or body.get("displayName") or body.get("name") or ident
    desc = (
        meta.get("description")
        or body.get("summary")
        or body.get("description")
        or title
    )
    if isinstance(desc, str) and len(desc) > 2000:
        desc = desc[:1997] + "..."

    cfg = body.get("config") or {}
    provider, model_name = parse_model(cfg)
    category = meta.get("category")
    tags = meta.get("tags") or []
    version = body.get("version") or body.get("schemaVersion") or "1.0.0"
    name = sanitize_name(str(ident))
    character = "character" in name or "character" in str(ident).lower()

    image = body.get("image") or cfg.get("image")
    if not image:
        image = f"ghcr.io/solizardking/robinhood-agents/{name}:{version}"

    labs = labels_from(category, tags, character=character)
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
            f"      subfolder: {yaml_escape(source_rel)}",
        ]
    )
    return "\n".join(lines)


def forge_example_to_yaml(
    body: dict,
    *,
    stem: str,
    source_rel: str,
    tag: str,
    repo_url: str,
) -> str:
    """Convert examples/robinhood-agent.json / solana-agent.json style docs."""
    platform = body.get("platform") or stem
    name_raw = body.get("name") or stem
    # DNS-safe name: prefer platform-prefixed slug
    name = sanitize_name(f"{platform}-{name_raw}")
    title = body.get("name") or stem
    desc = body.get("description") or title
    image = body.get("image") or f"ghcr.io/solizardking/robinhood-agents/{name}:1.0.0"
    # ipfs:// images are fine as catalog strings
    labs = labels_from(
        "forge-example",
        [platform, f"chain-{body.get('chainId', 'unknown')}"],
    )
    labs["platform"] = sanitize_name(str(platform))
    if body.get("chainId") is not None:
        labs["chain-id"] = str(body["chainId"])

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
            f"  modelProvider: {DEFAULT_MODEL_PROVIDER}",
            f"  modelName: {DEFAULT_MODEL_NAME}",
            "  source:",
            f"    image: {yaml_escape(image)}",
            "    repository:",
            f"      url: {yaml_escape(repo_url)}",
            f"      subfolder: {yaml_escape(source_rel)}",
        ]
    )
    # Metaplex fields only when clearly Solana-oriented example
    if str(platform).lower() in {"solana", "svm"}:
        lines.extend(
            [
                "  metaplex:",
                '    network: "solana-mainnet"',
                '    launchType: "bondingCurve"',
            ]
        )
    return "\n".join(lines)


def convert(source: Path, out_dir: Path, *, tag: str, repo_url: str) -> dict:
    agents_dir = source / "agents"
    examples_dir = source / "examples"
    docs: list[str] = []
    names: list[str] = []

    if agents_dir.is_dir():
        for path in sorted(agents_dir.glob("*.json")):
            body = load_json(path)
            if not body:
                continue
            doc = cheshire_agent_to_yaml(
                body,
                source_rel=f"agents/{path.name}",
                tag=tag,
                repo_url=repo_url,
            )
            if doc:
                docs.append(doc)
                names.append(body.get("identifier") or path.stem)

    if examples_dir.is_dir():
        for path in sorted(examples_dir.glob("*.json")):
            body = load_json(path)
            if not body:
                continue
            # Cheshire-schema examples vs forge platform examples
            if body.get("identifier") and body.get("meta"):
                doc = cheshire_agent_to_yaml(
                    body,
                    source_rel=f"examples/{path.name}",
                    tag=tag,
                    repo_url=repo_url,
                )
            else:
                doc = forge_example_to_yaml(
                    body,
                    stem=path.stem,
                    source_rel=f"examples/{path.name}",
                    tag=tag,
                    repo_url=repo_url,
                )
            if doc:
                docs.append(doc)
                names.append(body.get("identifier") or path.stem)

    if not docs:
        raise SystemExit(f"no agent JSON found under {agents_dir} or {examples_dir}")

    out_dir.mkdir(parents=True, exist_ok=True)
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
        "names": names,
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
        help="Path to robinhood-agents package root",
    )
    parser.add_argument(
        "--out",
        type=Path,
        default=repo_root / "examples" / "robinhood-agents-catalog",
        help="Output directory for generated YAML batches",
    )
    parser.add_argument("--tag", default=DEFAULT_TAG, help="Agent tag (default: robinhood)")
    parser.add_argument(
        "--repo-url",
        default=DEFAULT_REPO,
        help="Repository URL for Agent.spec.source.repository",
    )
    parser.add_argument("--apply", action="store_true", help="Apply batches via arctl")
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
        f"to {args.out} (tag={manifest['tag']})"
    )

    if args.apply:
        if not args.arctl.exists():
            raise SystemExit(f"arctl not found at {args.arctl}; build with make build-cli")
        return apply_batches(args.out, args.registry_url, args.arctl)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
