#!/usr/bin/env python3
"""Import SkillHub (Solizardking/skills) catalog into AgentRegistry.

Uses catalog.json as the source of truth (slug/name/description/category) and
points each Skill at skills/<slug> in the SkillHub git repo.

Examples:
  python3 scripts/import-skillhub.py \\
    --source /Users/8bit/Downloads/skillhub-main

  python3 scripts/import-skillhub.py \\
    --source /Users/8bit/Downloads/skillhub-main \\
    --apply --registry-url https://cheshire-agentregistry.fly.dev
"""

from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
from pathlib import Path

DEFAULT_TAG = "skillhub"
# Canonical hub source (GitHub). package.json also publishes as @x402solana/skills.
DEFAULT_REPO = "https://github.com/Solizardking/skillhub-main"
BATCH_SIZE = 50


def sanitize_name(name: str) -> str:
    name = name.lower().strip()
    name = re.sub(r"[^a-z0-9.-]+", "-", name)
    name = re.sub(r"-+", "-", name).strip("-")
    if not name:
        name = "unnamed-skill"
    if not re.match(r"^[a-z0-9]", name):
        name = "s-" + name
    # DNS-1123 subdomain max 253, but keep short for practical names
    return name[:63].rstrip("-.")


def yaml_escape(value) -> str:
    if value is None:
        return '""'
    text = str(value)
    if "\n" in text or text.strip() != text or any(c in text for c in ":\"'#{}[]&*!|>%@`"):
        return json.dumps(text, ensure_ascii=False)
    return text


def category_label(category: str) -> str:
    return sanitize_name(category or "uncategorized")[:63]


def skill_to_yaml(
    *,
    name: str,
    title: str,
    description: str,
    slug: str,
    category: str,
    tag: str,
    repo_url: str,
) -> str:
    labels = {
        "app": "skillhub",
        "catalog": "skillhub",
        "kind": "skill",
        "category": category_label(category),
    }
    if "/" in slug:
        labels["namespace"] = sanitize_name(slug.split("/")[0])[:63]

    lines = [
        "apiVersion: ar.dev/v1alpha1",
        "kind: Skill",
        "metadata:",
        f"  name: {name}",
        f"  tag: {tag}",
        "  labels:",
    ]
    for key, value in sorted(labels.items()):
        lines.append(f"    {key}: {yaml_escape(value)}")

    desc = description or title
    if len(desc) > 2000:
        desc = desc[:1997] + "..."

    lines.extend(
        [
            "spec:",
            f"  title: {yaml_escape(title)}",
            f"  description: {yaml_escape(desc)}",
            "  source:",
            "    repository:",
            f"      url: {yaml_escape(repo_url)}",
            f"      subfolder: {yaml_escape('skills/' + slug)}",
        ]
    )
    return "\n".join(lines)


def load_entries(source: Path) -> list[dict]:
    catalog_path = source / "catalog.json"
    if not catalog_path.is_file():
        raise SystemExit(f"catalog.json not found at {catalog_path}")

    catalog = json.loads(catalog_path.read_text(encoding="utf-8"))
    if not isinstance(catalog, list):
        raise SystemExit("catalog.json must be a JSON array")

    entries: list[dict] = []
    seen: set[str] = set()

    for item in catalog:
        slug = item.get("slug") or item.get("name")
        if not slug:
            continue
        skill_md = source / "skills" / slug / "SKILL.md"
        if not skill_md.is_file():
            print(f"warn: missing SKILL.md for slug {slug}", file=sys.stderr)
            continue

        # Prefer path-based unique DNS name so nested skills don't collide.
        name = sanitize_name(str(slug).replace("/", "-"))
        if name in seen:
            name = sanitize_name(f"skillhub-{name}")
        seen.add(name)

        title = item.get("name") or Path(slug).name
        # Prefer human title casing for flat names
        if title == Path(slug).name:
            title = title.replace("-", " ")

        entries.append(
            {
                "name": name,
                "title": title,
                "description": item.get("description") or title,
                "slug": slug,
                "category": item.get("category") or "Utilities",
            }
        )

    return entries


def convert(entries: list[dict], out_dir: Path, *, tag: str, repo_url: str) -> dict:
    docs = [
        skill_to_yaml(
            name=e["name"],
            title=e["title"],
            description=e["description"],
            slug=e["slug"],
            category=e["category"],
            tag=tag,
            repo_url=repo_url,
        )
        for e in entries
    ]

    out_dir.mkdir(parents=True, exist_ok=True)
    for old in out_dir.glob("skills-batch-*.yaml"):
        old.unlink()

    batch_files: list[Path] = []
    for i in range(0, len(docs), BATCH_SIZE):
        chunk = docs[i : i + BATCH_SIZE]
        out = out_dir / f"skills-batch-{(i // BATCH_SIZE) + 1:02d}.yaml"
        out.write_text("\n---\n".join(chunk) + "\n", encoding="utf-8")
        batch_files.append(out)

    by_cat: dict[str, int] = {}
    for e in entries:
        by_cat[e["category"]] = by_cat.get(e["category"], 0) + 1

    manifest = {
        "total": len(docs),
        "tag": tag,
        "repo": repo_url,
        "batches": [p.name for p in batch_files],
        "byCategory": by_cat,
        "source": "catalog.json",
    }
    (out_dir / "import-manifest.json").write_text(
        json.dumps(manifest, indent=2) + "\n", encoding="utf-8"
    )
    return manifest


def apply_batches(out_dir: Path, registry_url: str, arctl: Path) -> int:
    batches = sorted(out_dir.glob("skills-batch-*.yaml"))
    if not batches:
        print("no batch files", file=sys.stderr)
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
            print(f"  batch failed exit={proc.returncode}", file=sys.stderr)
    return 1 if failed else 0


def main(argv: list[str] | None = None) -> int:
    repo_root = Path(__file__).resolve().parents[1]
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--source",
        type=Path,
        required=True,
        help="Path to skillhub-main (contains catalog.json + skills/)",
    )
    parser.add_argument(
        "--out",
        type=Path,
        default=repo_root / "examples" / "skillhub-catalog",
    )
    parser.add_argument("--tag", default=DEFAULT_TAG)
    parser.add_argument("--repo-url", default=DEFAULT_REPO)
    parser.add_argument("--apply", action="store_true")
    parser.add_argument("--registry-url", default="http://localhost:12121")
    parser.add_argument("--arctl", type=Path, default=repo_root / "bin" / "arctl")
    args = parser.parse_args(argv)

    source = args.source.expanduser().resolve()
    if not source.is_dir():
        raise SystemExit(f"source not found: {source}")

    entries = load_entries(source)
    if not entries:
        raise SystemExit("no skills loaded from catalog.json")

    manifest = convert(entries, args.out, tag=args.tag, repo_url=args.repo_url)
    print(
        f"wrote {manifest['total']} skills in {len(manifest['batches'])} batches "
        f"to {args.out} (tag={args['tag'] if False else args.tag})"
    )
    print("by category:")
    for cat, n in sorted(manifest["byCategory"].items(), key=lambda x: -x[1]):
        print(f"  {n:4d}  {cat}")

    if args.apply:
        if not args.arctl.exists():
            raise SystemExit(f"arctl not found: {args.arctl}")
        return apply_batches(args.out, args.registry_url, args.arctl)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
