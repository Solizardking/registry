#!/usr/bin/env python3
"""Unit tests: agent source/repo links target github.com/solizardking/agents.

Drives the real import/publish helpers and scans committed agent example YAMLs.
Skill catalogs and skill import defaults are intentionally out of scope.
"""

from __future__ import annotations

import importlib.util
import re
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SCRIPTS = ROOT / "scripts"

# Org/repo match is case-insensitive (plan + OBJECTIVE).
AGENTS_URL_RE = re.compile(
    r"https?://github\.com/solizardking/agents(?:\.git)?/?$",
    re.IGNORECASE,
)
# Legacy agent-source hosts that must no longer appear in agent-link sites.
FORBIDDEN_AGENT_HOST_RE = re.compile(
    r"github\.com/solizardking/"
    r"(cheshire-terminal(?:-main)?|robinhood-agents|Clawd-agents|clawd-agents)"
    r"(?:[\"'\s]|$)",
    re.IGNORECASE,
)
# Repository URL lines inside Agent YAML (quoted or bare).
REPO_URL_LINE_RE = re.compile(
    r"^\s*url:\s*[\"']?(https://github\.com/[^\"'\s]+)[\"']?\s*$",
    re.MULTILINE | re.IGNORECASE,
)

# In-scope agent paths (examples + agent publish/import + agent docs).
AGENT_LINK_GLOBS = [
    "examples/cheshire-terminal-catalog/agents-batch-*.yaml",
    "examples/robinhood-agents-catalog/agents-batch-*.yaml",
    "examples/robinhood-agents-catalog/README.md",
    "examples/clawd-perps-agent.yaml",
    "examples/clawd-perps-agent-project/agent.yaml",
    "docs/clawd-perps-integration.md",
    "publish-agents-public.sh",
    "publish-agents-public.py",
    "publish-clawd-agents.sh",
    "scripts/import-cheshire-agents.py",
    "scripts/import-robinhood-agents.py",
]


def load_module(name: str, path: Path):
    spec = importlib.util.spec_from_file_location(name, path)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"cannot load {path}")
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


def agent_link_files() -> list[Path]:
    files: list[Path] = []
    for pattern in AGENT_LINK_GLOBS:
        if any(ch in pattern for ch in "*?["):
            files.extend(sorted(ROOT.glob(pattern)))
        else:
            p = ROOT / pattern
            if p.is_file():
                files.append(p)
    return files


class TestAgentSourceURLs(unittest.TestCase):
    def test_import_cheshire_default_repo(self):
        mod = load_module("import_cheshire_agents", SCRIPTS / "import-cheshire-agents.py")
        self.assertTrue(
            AGENTS_URL_RE.match(mod.DEFAULT_REPO),
            f"import-cheshire DEFAULT_REPO={mod.DEFAULT_REPO!r}",
        )
        # Drive the real YAML emitter with the shipped default.
        doc = mod.agent_to_yaml_doc(
            "sample-agent",
            {"meta": {"title": "Sample", "description": "d"}, "config": {}},
            {},
            tag=mod.DEFAULT_TAG,
            repo_url=mod.DEFAULT_REPO,
        )
        self.assertIn(mod.DEFAULT_REPO, doc)
        self.assertIsNone(FORBIDDEN_AGENT_HOST_RE.search(doc))

    def test_import_robinhood_default_repo(self):
        mod = load_module("import_robinhood_agents", SCRIPTS / "import-robinhood-agents.py")
        self.assertTrue(
            AGENTS_URL_RE.match(mod.DEFAULT_REPO),
            f"import-robinhood DEFAULT_REPO={mod.DEFAULT_REPO!r}",
        )
        doc = mod.cheshire_agent_to_yaml(
            {
                "identifier": "rh-sample",
                "meta": {"title": "RH Sample", "description": "d"},
                "config": {},
            },
            source_rel="agents/rh-sample.json",
            tag=mod.DEFAULT_TAG,
            repo_url=mod.DEFAULT_REPO,
        )
        self.assertIsNotNone(doc)
        self.assertIn(mod.DEFAULT_REPO, doc)
        # Only GitHub web agent-source hosts are in scope; GHCR image refs may differ.
        self.assertIsNone(
            FORBIDDEN_AGENT_HOST_RE.search(doc),
            f"legacy github agent host leaked into YAML:\n{doc}",
        )

    def test_publish_agents_public_fallback_emits_agents_url(self):
        mod = load_module("publish_agents_public", ROOT / "publish-agents-public.py")
        # No metadata.github → must use the shipped agents fallback.
        yaml_doc = mod.build_agent_yaml(
            {
                "name": "pub-sample",
                "meta": {"title": "Pub", "description": "d", "category": "defi"},
                "version": "1.0.0",
                "config": {"model": "clawdrouter/auto"},
            }
        )
        urls = REPO_URL_LINE_RE.findall(yaml_doc)
        self.assertTrue(urls, "build_agent_yaml must emit a repository url")
        for u in urls:
            self.assertTrue(AGENTS_URL_RE.match(u), f"unexpected url {u!r} in {yaml_doc}")
        self.assertIsNone(FORBIDDEN_AGENT_HOST_RE.search(yaml_doc))

    def test_publish_clawd_git_repo_constant(self):
        text = (ROOT / "publish-clawd-agents.sh").read_text(encoding="utf-8")
        m = re.search(r'^GIT_REPO="([^"]+)"', text, re.MULTILINE)
        self.assertIsNotNone(m, "GIT_REPO assignment missing in publish-clawd-agents.sh")
        self.assertTrue(AGENTS_URL_RE.match(m.group(1)), f"GIT_REPO={m.group(1)!r}")

    def test_agent_example_yamls_use_agents_repo(self):
        files = sorted(
            set(
                list(ROOT.glob("examples/cheshire-terminal-catalog/agents-batch-*.yaml"))
                + list(ROOT.glob("examples/robinhood-agents-catalog/agents-batch-*.yaml"))
                + [ROOT / "examples/clawd-perps-agent.yaml"]
                + [ROOT / "examples/clawd-perps-agent-project/agent.yaml"]
            )
        )
        self.assertGreater(len(files), 5, "expected agent example YAML inventory")
        total_urls = 0
        for path in files:
            text = path.read_text(encoding="utf-8")
            urls = REPO_URL_LINE_RE.findall(text)
            # Agent batch docs all set repository.url
            self.assertTrue(urls, f"no repository url lines in {path.relative_to(ROOT)}")
            for u in urls:
                # Skip non-github local/http service urls if any (clawd MCP)
                if "github.com" not in u.lower():
                    continue
                total_urls += 1
                self.assertTrue(
                    AGENTS_URL_RE.match(u),
                    f"{path.relative_to(ROOT)}: want agents repo, got {u!r}",
                )
            bad = FORBIDDEN_AGENT_HOST_RE.search(text)
            self.assertIsNone(
                bad,
                f"{path.relative_to(ROOT)} still references legacy agent host: {bad.group(0) if bad else ''}",
            )
        self.assertGreater(total_urls, 50, f"expected many agent repo urls, got {total_urls}")

    def test_agent_link_sites_have_no_legacy_agent_hosts(self):
        files = agent_link_files()
        self.assertGreater(len(files), 10)
        offenders: list[str] = []
        for path in files:
            text = path.read_text(encoding="utf-8")
            for m in FORBIDDEN_AGENT_HOST_RE.finditer(text):
                offenders.append(f"{path.relative_to(ROOT)}: {m.group(0)}")
        self.assertEqual(offenders, [], "legacy agent-source hosts remain:\n" + "\n".join(offenders))

    def test_agents_url_present_in_defaults_and_samples(self):
        hits: list[str] = []
        for path in agent_link_files():
            text = path.read_text(encoding="utf-8")
            if re.search(r"github\.com/solizardking/agents", text, re.IGNORECASE):
                hits.append(str(path.relative_to(ROOT)))
        # Must cover catalogs, import scripts, publish helpers, docs
        joined = "\n".join(hits)
        for required in (
            "examples/cheshire-terminal-catalog/",
            "examples/robinhood-agents-catalog/",
            "scripts/import-cheshire-agents.py",
            "scripts/import-robinhood-agents.py",
            "publish-clawd-agents.sh",
            "publish-agents-public",
            "docs/clawd-perps-integration.md",
        ):
            self.assertTrue(
                any(required in h for h in hits),
                f"missing agents URL coverage for {required!r}; hits:\n{joined}",
            )


if __name__ == "__main__":
    unittest.main()
