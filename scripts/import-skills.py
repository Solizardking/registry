#!/usr/bin/env python3
"""Import SKILL.md packages into AgentRegistry as ar.dev/v1alpha1 Skill resources.

Scans skill directories for SKILL.md frontmatter (name/description) and emits
multi-doc YAML for arctl apply.

Default sources (when --source-root is a monorepo like cheshire-terminal):
  - <root>/agents/skills/*
  - <root>/robinhood-agents/skills/*
  - <root>/skills/*            (top-level only; skips nested/deprecated)

Examples:
  python3 scripts/import-skills.py --source-root /path/to/cheshire-terminal
  python3 scripts/import-skills.py --source-root /path/to/cheshire-terminal \\
    --apply --registry-url https://cheshire-agentregistry.fly.dev
"""

from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
from pathlib import Path

DEFAULT_TAG = "stable"
BATCH_SIZE = 40
DEFAULT_REPO = "https://github.com/Solizardking/cheshire-terminal-main"
ROBINHOOD_REPO = "https://github.com/Solizardking/robinhood-agents"


def sanitize_name(name: str) -> str:
    name = name.lower().strip()
    name = re.sub(r"[^a-z0-9.-]+", "-", name)
    name = re.sub(r"-+", "-", name).strip("-")
    if not name:
        name = "unnamed-skill"
    if not re.match(r"^[a-z0-9]", name):
        name = "s-" + name
    return name[:63].rstrip("-.")


def yaml_escape(value) -> str:
    if value is None:
        return '""'
    text = str(value)
    if "\n" in text or text.strip() != text or any(c in text for c in ":\"'#{}[]&*!|>%@`"):
        return json.dumps(text, ensure_ascii=False)
    return text


def parse_frontmatter(skill_md: Path) -> tuple[str | None, str | None, str | None]:
    """Return (name, description, title) from SKILL.md YAML frontmatter."""
    try:
        text = skill_md.read_text(encoding="utf-8", errors="replace")
    except OSError:
        return None, None, None
    name = desc = title = None
    if not text.startswith("---"):
        return None, None, None
    end = text.find("\n---", 3)
    if end == -1:
        return None, None, None
    fm = text[3:end]
    for line in fm.splitlines():
        if line.startswith("name:"):
            name = line.split(":", 1)[1].strip().strip("\"'")
        elif line.startswith("description:"):
            desc = line.split(":", 1)[1].strip().strip("\"'")
        elif line.startswith("title:"):
            title = line.split(":", 1)[1].strip().strip("\"'")
    return name, desc, title


def skill_to_yaml(
    *,
    name: str,
    title: str,
    description: str,
    repo_url: str,
    subfolder: str,
    tag: str,
    labels: dict[str, str],
) -> str:
    lines = [
        "apiVersion: ar.dev/v1alpha1",
        "kind: Skill",
        "metadata:",
        f"  name: {sanitize_name(name)}",
        f"  tag: {tag}",
        "  labels:",
    ]
    for key, value in sorted(labels.items()):
        lines.append(f"    {key}: {yaml_escape(value)}")
    lines.extend(
        [
            "spec:",
            f"  title: {yaml_escape(title)}",
            f"  description: {yaml_escape(description)}",
            "  source:",
            "    repository:",
            f"      url: {yaml_escape(repo_url)}",
            f"      subfolder: {yaml_escape(subfolder)}",
        ]
    )
    return "\n".join(lines)


def collect_skills(source_root: Path, *, include_nested: bool) -> list[dict]:
    """Collect skill packages from known locations under source_root."""
    entries: list[dict] = []
    seen_names: set[str] = set()

    def add(
        skill_md: Path,
        *,
        catalog: str,
        repo_url: str,
        subfolder_prefix: str,
        monorepo_rel: str,
    ) -> None:
        name_fm, desc_fm, title_fm = parse_frontmatter(skill_md)
        folder = skill_md.parent.name
        raw_name = name_fm or folder
        name = sanitize_name(raw_name)
        # disambiguate collisions by prefixing catalog slug
        if name in seen_names:
            name = sanitize_name(f"{catalog}-{raw_name}")
        if name in seen_names:
            # path-based last resort
            rel = skill_md.parent.relative_to(source_root)
            name = sanitize_name(str(rel).replace("/", "-"))
        seen_names.add(name)

        title = title_fm or raw_name.replace("-", " ").title()
        desc = desc_fm or f"Skill package: {raw_name}"
        if len(desc) > 2000:
            desc = desc[:1997] + "..."

        rel_parent = skill_md.parent
        # subfolder relative to the git repo root convention
        try:
            sub = str(rel_parent.relative_to(source_root))
        except ValueError:
            sub = f"{subfolder_prefix}/{folder}"
        if monorepo_rel and not sub.startswith(monorepo_rel.split("/")[0]):
            # keep as computed
            pass

        entries.append(
            {
                "name": name,
                "title": title,
                "description": desc,
                "repo_url": repo_url,
                "subfolder": sub,
                "catalog": catalog,
                "path": str(skill_md),
            }
        )

    # 1) agents/skills — agent hub pack
    agents_skills = source_root / "agents" / "skills"
    if agents_skills.is_dir():
        for skill_md in sorted(agents_skills.glob("*/SKILL.md")):
            add(
                skill_md,
                catalog="agents-skills",
                repo_url=DEFAULT_REPO,
                subfolder_prefix="agents/skills",
                monorepo_rel="agents/skills",
            )

    # 2) robinhood-agents/skills
    rh_skills = source_root / "robinhood-agents" / "skills"
    if rh_skills.is_dir():
        for skill_md in sorted(rh_skills.glob("*/SKILL.md")):
            add(
                skill_md,
                catalog="robinhood-agents-skills",
                repo_url=ROBINHOOD_REPO,
                subfolder_prefix="skills",
                monorepo_rel="skills",
            )
            # fix subfolder to skills/<name> relative to robinhood-agents repo
            entries[-1]["subfolder"] = f"skills/{skill_md.parent.name}"
            entries[-1]["repo_url"] = ROBINHOOD_REPO

    # 3) skills/ top-level (and optionally nested)
    skills_root = source_root / "skills"
    if skills_root.is_dir():
        if include_nested:
            skill_files = sorted(skills_root.rglob("SKILL.md"))
        else:
            skill_files = sorted(skills_root.glob("*/SKILL.md"))
        for skill_md in skill_files:
            if any(p in skill_md.parts for p in ("node_modules", "target", "dist", ".git", "deprecated")):
                continue
            add(
                skill_md,
                catalog="cheshire-skills",
                repo_url=DEFAULT_REPO,
                subfolder_prefix="skills",
                monorepo_rel="skills",
            )

    return entries


def convert(entries: list[dict], out_dir: Path, *, tag: str) -> dict:
    docs: list[str] = []
    for e in entries:
        labels = {
            "app": "cheshire-terminal",
            "catalog": e["catalog"],
            "kind": "skill",
        }
        docs.append(
            skill_to_yaml(
                name=e["name"],
                title=e["title"],
                description=e["description"],
                repo_url=e["repo_url"],
                subfolder=e["subfolder"],
                tag=tag,
                labels=labels,
            )
        )

    out_dir.mkdir(parents=True, exist_ok=True)
    for old in out_dir.glob("skills-batch-*.yaml"):
        old.unlink()

    batch_files: list[Path] = []
    for i in range(0, len(docs), BATCH_SIZE):
        chunk = docs[i : i + BATCH_SIZE]
        out = out_dir / f"skills-batch-{(i // BATCH_SIZE) + 1:02d}.yaml"
        out.write_text("\n---\n".join(chunk) + "\n", encoding="utf-8")
        batch_files.append(out)

    manifest = {
        "total": len(docs),
        "tag": tag,
        "batches": [p.name for p in batch_files],
        "byCatalog": {},
        "names": [e["name"] for e in entries],
    }
    for e in entries:
        manifest["byCatalog"][e["catalog"]] = manifest["byCatalog"].get(e["catalog"], 0) + 1

    (out_dir / "import-manifest.json").write_text(
        json.dumps(manifest, indent=2) + "\n", encoding="utf-8"
    )
    return manifest


def apply_batches(out_dir: Path, registry_url: str, arctl: Path) -> int:
    batches = sorted(out_dir.glob("skills-batch-*.yaml"))
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
        "--source-root",
        type=Path,
        required=True,
        help="Monorepo root containing agents/skills, skills/, robinhood-agents/skills",
    )
    parser.add_argument(
        "--out",
        type=Path,
        default=repo_root / "examples" / "skills-catalog",
        help="Output directory for generated YAML batches",
    )
    parser.add_argument("--tag", default=DEFAULT_TAG)
    parser.add_argument(
        "--include-nested",
        action="store_true",
        help="Also import nested skills under skills/** (default: top-level only)",
    )
    parser.add_argument("--apply", action="store_true")
    parser.add_argument("--registry-url", default="http://localhost:12121")
    parser.add_argument(
        "--arctl",
        type=Path,
        default=repo_root / "bin" / "arctl",
    )
    args = parser.parse_args(argv)

    source = args.source_root.expanduser().resolve()
    if not source.is_dir():
        raise SystemExit(f"source root not found: {source}")

    entries = collect_skills(source, include_nested=args.include_nested)
    if not entries:
        raise SystemExit(f"no SKILL.md packages found under {source}")

    manifest = convert(entries, args.out, tag=args.tag)
    print(
        f"wrote {manifest['total']} skills in {len(manifest['batches'])} batches "
        f"to {args.out} (by catalog: {manifest['byCatalog']})"
    )

    if args.apply:
        if not args.arctl.exists():
            raise SystemExit(f"arctl not found at {args.arctl}")
        return apply_batches(args.out, args.registry_url, args.arctl)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
