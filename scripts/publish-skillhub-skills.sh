#!/usr/bin/env bash
# Publish Skill Hub skills into a running agentregistry.
#
# Reads skillhub catalog.json (all nested slugs, e.g. google/cloud/gke-basics)
# and registers each skill. Works with:
#   • Modern registry:  arctl apply / POST /v0/apply  (ar.dev/v1alpha1 Skill)
#   • Legacy registry:  POST /v0/skills               (v0.3.x daemon)
#
# Usage:
#   ./scripts/publish-skillhub-skills.sh
#   SKILLHUB_DIR=/path/to/skillhub-main ./scripts/publish-skillhub-skills.sh
#   ./scripts/publish-skillhub-skills.sh --limit 10
#   ./scripts/publish-skillhub-skills.sh --dry-run
#   ./scripts/publish-skillhub-skills.sh --prefix solana --prefix pump
#   ./scripts/publish-skillhub-skills.sh --category "Solana / Blockchain"
#
# Env:
#   SKILLHUB_DIR   Path to skillhub checkout (default: auto-detect)
#   REGISTRY_URL   agentregistry base URL (default: http://localhost:12121)
#   ARCTL          arctl binary (default: bin/arctl, then PATH)
#   REPO_URL       Git URL stored on each skill
#                  (default: https://github.com/Solizardking/skills)
#   API_MODE       auto | apply | legacy  (default: auto)
#   ONCHAIN=1      Attach Skill Hub on-chain proofs (Arweave packages + verification URL)
#                  Prefer: SKILLHUB_DIR=... npm run publish:agentregistry:onchain  (from skillhub)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

REGISTRY_URL="${REGISTRY_URL:-${ARCTL_API_BASE_URL:-http://localhost:12121}}"
REPO_URL="${REPO_URL:-https://github.com/Solizardking/skills}"
API_MODE="${API_MODE:-auto}"
DRY_RUN=0
LIMIT=0
PREFIXES=()
CATEGORY=""
OUT_DIR=""

usage() {
  sed -n '2,28p' "$0" | sed 's/^# \{0,1\}//'
  exit 0
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help) usage ;;
    --dry-run) DRY_RUN=1; shift ;;
    --limit) LIMIT="${2:?}"; shift 2 ;;
    --prefix) PREFIXES+=("$2"); shift 2 ;;
    --category) CATEGORY="${2:?}"; shift 2 ;;
    --skillhub-dir) SKILLHUB_DIR="${2:?}"; shift 2 ;;
    --registry-url) REGISTRY_URL="${2:?}"; shift 2 ;;
    --api-mode) API_MODE="${2:?}"; shift 2 ;;
    --out-dir) OUT_DIR="${2:?}"; shift 2 ;;
    *) echo "Unknown arg: $1" >&2; usage ;;
  esac
done

# ── resolve skillhub dir ────────────────────────────────────────────────────
resolve_skillhub() {
  local candidates=(
    "${SKILLHUB_DIR:-}"
    "$HOME/Downloads/skillhub-main"
    "$HOME/skillhub-main"
    "$REPO_ROOT/../skillhub-main"
    "/Users/8bit/Downloads/skillhub-main"
    "/Users/8bit/Clawd-agents/skills"
  )
  local c
  for c in "${candidates[@]}"; do
    [[ -z "$c" ]] && continue
    if [[ -f "$c/catalog.json" ]]; then
      echo "$c"
      return 0
    fi
  done
  return 1
}

if ! SKILLHUB_DIR="$(resolve_skillhub)"; then
  cat >&2 <<EOF
ERROR: Could not find skillhub (need catalog.json).

Set SKILLHUB_DIR to your skillhub checkout, e.g.:
  SKILLHUB_DIR=/Users/8bit/Downloads/skillhub-main $0
EOF
  exit 1
fi

CATALOG="$SKILLHUB_DIR/catalog.json"
[[ -f "$CATALOG" ]] || { echo "ERROR: missing $CATALOG" >&2; exit 1; }

# ── resolve arctl ───────────────────────────────────────────────────────────
resolve_arctl() {
  if [[ -n "${ARCTL:-}" && -x "$ARCTL" ]]; then echo "$ARCTL"; return 0; fi
  if [[ -x "$REPO_ROOT/bin/arctl" ]]; then echo "$REPO_ROOT/bin/arctl"; return 0; fi
  if command -v arctl >/dev/null 2>&1; then command -v arctl; return 0; fi
  return 1
}

ARCTL_BIN=""
if ARCTL_BIN="$(resolve_arctl)"; then
  :
else
  ARCTL_BIN=""
fi

# ── wait for registry / detect API mode ─────────────────────────────────────
wait_registry() {
  local i
  for i in $(seq 1 40); do
    if curl -fsS "$REGISTRY_URL/v0/skills?limit=1" >/dev/null 2>&1; then
      return 0
    fi
    if curl -fsS "$REGISTRY_URL/v0/version" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  return 1
}

detect_api_mode() {
  # Prefer modern apply if the endpoint exists and accepts Skill YAML.
  local code
  code=$(curl -s -o /dev/null -w '%{http_code}' -X POST "$REGISTRY_URL/v0/apply" \
    -H "Content-Type: application/yaml" \
    --data-binary $'apiVersion: ar.dev/v1alpha1\nkind: Skill\nmetadata:\n  name: __probe__\nspec:\n  description: probe\n' 2>/dev/null || echo 000)
  case "$code" in
    200|201|202|207|400|422) echo "apply"; return 0 ;;  # 400/422 = endpoint lives, payload rejected
  esac
  code=$(curl -s -o /dev/null -w '%{http_code}' "$REGISTRY_URL/v0/skills?limit=1" 2>/dev/null || echo 000)
  if [[ "$code" == "200" ]]; then
    echo "legacy"
    return 0
  fi
  echo "legacy"
}

echo "=== Publish Skill Hub → agentregistry ==="
echo "  skillhub:  $SKILLHUB_DIR"
echo "  catalog:   $CATALOG"
echo "  registry:  $REGISTRY_URL"
echo "  arctl:     ${ARCTL_BIN:-none}"
echo "  repo url:  $REPO_URL"
echo "  dry-run:   $DRY_RUN"
[[ "$LIMIT" -gt 0 ]] && echo "  limit:     $LIMIT"
[[ ${#PREFIXES[@]} -gt 0 ]] && echo "  prefixes:  ${PREFIXES[*]}"
[[ -n "$CATEGORY" ]] && echo "  category:  $CATEGORY"
echo ""

if [[ "$DRY_RUN" -eq 0 ]]; then
  if ! wait_registry; then
    cat >&2 <<EOF
ERROR: agentregistry is not reachable at $REGISTRY_URL

Start it with:
  arctl daemon start

Then re-run this script.
EOF
    exit 1
  fi
fi

if [[ "$API_MODE" == "auto" ]]; then
  if [[ "$DRY_RUN" -eq 1 ]]; then
    API_MODE="legacy"  # dry-run path does not need endpoint probe
  else
    API_MODE="$(detect_api_mode)"
  fi
fi
echo "  api mode:  $API_MODE"
echo ""

# ── build selected skill list from catalog ──────────────────────────────────
WORKDIR="${OUT_DIR:-$(mktemp -d -t skillhub-publish-XXXXXX)}"
mkdir -p "$WORKDIR"
CLEANUP_WORKDIR=0
if [[ -z "${OUT_DIR:-}" ]]; then
  CLEANUP_WORKDIR=1
  trap '[[ "$CLEANUP_WORKDIR" -eq 1 ]] && rm -rf "$WORKDIR"' EXIT
fi

export SKILLHUB_DIR CATALOG REPO_URL WORKDIR LIMIT CATEGORY API_MODE
export PREFIXES_CSV
PREFIXES_CSV="$(IFS=,; echo "${PREFIXES[*]:-}")"

python3 - <<'PY'
import hashlib
import json
import os
import re
import sys
from pathlib import Path

catalog_path = Path(os.environ["CATALOG"])
skillhub = Path(os.environ["SKILLHUB_DIR"])
repo_url = os.environ["REPO_URL"]
workdir = Path(os.environ["WORKDIR"])
limit = int(os.environ.get("LIMIT") or 0)
category_filter = (os.environ.get("CATEGORY") or "").strip()
prefixes = [p for p in (os.environ.get("PREFIXES_CSV") or "").split(",") if p]

with catalog_path.open() as f:
    items = json.load(f)
if not isinstance(items, list):
    print("catalog.json must be a JSON array", file=sys.stderr)
    sys.exit(1)

def dns_name(slug: str) -> str:
    name = slug.strip().lower().replace("/", "-").replace("_", "-")
    name = re.sub(r"[^a-z0-9.-]+", "-", name)
    name = re.sub(r"-{2,}", "-", name).strip("-.")
    if not name:
        name = "skill"
    if not name[0].isalnum():
        name = "s-" + name
    if not name[-1].isalnum():
        name = name + "0"
    if len(name) > 63:
        # Legacy name path often uses shorter identifiers; keep unique suffix.
        digest = hashlib.sha1(slug.encode()).hexdigest()[:8]
        name = name[:63 - 9].rstrip("-.") + "-" + digest
        if not name[0].isalnum():
            name = "s" + name[1:]
    return name

def yaml_escape(s: str) -> str:
    return (s or "").replace("\\", "\\\\").replace('"', '\\"').replace("\n", " ").strip()

selected = []
for item in items:
    slug = (item.get("slug") or "").strip()
    if not slug:
        continue
    if prefixes and not any(
        slug == p or slug.startswith(p.rstrip("/") + "/") or slug.startswith(p)
        for p in prefixes
    ):
        continue
    if category_filter and (item.get("category") or "") != category_filter:
        continue

    if (skillhub / "skills" / slug / "SKILL.md").is_file():
        subfolder = f"skills/{slug}"
    elif (skillhub / slug / "SKILL.md").is_file():
        subfolder = slug
    else:
        subfolder = f"skills/{slug}"

    selected.append({
        "slug": slug,
        "name": dns_name(slug),
        "title": item.get("name") or slug.split("/")[-1],
        "description": item.get("description") or f"Skill: {slug}",
        "category": item.get("category") or "Uncategorized",
        "subfolder": subfolder,
        "repo_url": repo_url,
    })

if limit > 0:
    selected = selected[:limit]

seen = set()
unique = []
for s in selected:
    if s["name"] in seen:
        digest = hashlib.sha1(s["slug"].encode()).hexdigest()[:6]
        base = s["name"][:63 - 7].rstrip("-.")
        s["name"] = f"{base}-{digest}"
    seen.add(s["name"])
    unique.append(s)
selected = unique

(workdir / "manifest.json").write_text(json.dumps(selected, indent=2))
print(f"Selected {len(selected)} skills", flush=True)

# Write modern multi-doc YAML (for apply mode) and NDJSON for legacy POSTs.
docs = []
for s in selected:
    desc = yaml_escape(s["description"])[:2000]
    title = yaml_escape(s["title"])[:200]
    cat = yaml_escape(s["category"])
    docs.append(
        f"""apiVersion: ar.dev/v1alpha1
kind: Skill
metadata:
  name: {s["name"]}
  tag: latest
  labels:
    skillhub.slug: "{yaml_escape(s["slug"])}"
    skillhub.category: "{cat}"
spec:
  title: "{title}"
  description: "{desc}"
  source:
    repository:
      url: {s["repo_url"]}
      subfolder: {s["subfolder"]}
"""
    )
(workdir / "skills-all.yaml").write_text("---\n".join(docs) if docs else "")

# Legacy payloads (one JSON object per line)
with (workdir / "legacy.ndjson").open("w") as out:
    for s in selected:
        payload = {
            "name": s["name"],
            "version": "1.0.0",
            "title": s["title"][:200],
            "description": (s["description"] or "")[:2000].replace("\n", " "),
            "category": s["category"],
            "repository": {
                "url": s["repo_url"],
                "source": "github",
            },
        }
        # Optional fields some servers ignore:
        payload["websiteUrl"] = f"https://skills.x402.wtf/skills/{s['slug']}"
        out.write(json.dumps(payload, ensure_ascii=False) + "\n")

print(f"Wrote {workdir / 'skills-all.yaml'} and {workdir / 'legacy.ndjson'}", flush=True)
PY

TOTAL=$(python3 -c "import json; print(len(json.load(open('$WORKDIR/manifest.json'))))")
if [[ "$TOTAL" -eq 0 ]]; then
  echo "No skills matched filters. Nothing to publish."
  exit 0
fi

ok=0
err=0
skip=0

if [[ "$DRY_RUN" -eq 1 ]]; then
  echo "(dry-run) would publish $TOTAL skills via api mode=$API_MODE"
  echo "  YAML:   $WORKDIR/skills-all.yaml"
  echo "  NDJSON: $WORKDIR/legacy.ndjson"
  head -n 3 "$WORKDIR/legacy.ndjson" | sed 's/^/  sample: /'
  echo "=== Summary ==="
  echo "  Selected: $TOTAL"
  echo "  Mode:     dry-run"
  exit 0
fi

publish_legacy() {
  local line name code body msg
  while IFS= read -r line; do
    [[ -z "$line" ]] && continue
    name=$(python3 -c "import json,sys; print(json.loads(sys.argv[1])['name'])" "$line")
    printf "  %-50s " "$name"
    body=$(curl -sS -w "\n%{http_code}" -X POST "$REGISTRY_URL/v0/skills" \
      -H "Content-Type: application/json" \
      -d "$line" 2>&1 || true)
    code=$(echo "$body" | tail -1)
    msg=$(echo "$body" | sed '$d')
    if [[ "$code" == "200" || "$code" == "201" ]]; then
      echo "✓"
      ok=$((ok + 1))
    elif [[ "$code" == "409" ]] || echo "$msg" | grep -qiE 'duplicate version|already exists|conflict'; then
      echo "· exists"
      skip=$((skip + 1))
      ok=$((ok + 1))
    else
      detail=$(echo "$msg" | python3 -c "import sys,json
try:
 d=json.load(sys.stdin)
 errs=d.get('errors') or []
 if errs and isinstance(errs, list) and isinstance(errs[0], dict):
  print(errs[0].get('message') or d.get('detail') or d.get('title') or d)
 else:
  print(d.get('detail') or d.get('title') or d)
except Exception:
 print(sys.stdin.read()[:160])" 2>/dev/null || echo "$msg" | head -c 160)
      echo "✗ HTTP $code: $detail"
      err=$((err + 1))
    fi
  done < "$WORKDIR/legacy.ndjson"
}

publish_apply() {
  local yaml="$WORKDIR/skills-all.yaml"
  if [[ -n "$ARCTL_BIN" ]] && "$ARCTL_BIN" apply --help >/dev/null 2>&1; then
    # Chunk the multi-doc file into batches of 40 for friendlier output
    python3 - <<'PY' "$yaml" "$WORKDIR"
import sys
from pathlib import Path
src = Path(sys.argv[1]).read_text()
workdir = Path(sys.argv[2])
docs = [d.strip() for d in src.split("\n---\n") if d.strip()]
batch = 40
paths = []
for i in range(0, len(docs), batch):
    p = workdir / f"apply-batch-{i//batch:04d}.yaml"
    p.write_text("\n---\n".join(docs[i:i+batch]) + "\n")
    paths.append(str(p))
(workdir / "apply-batches.txt").write_text("\n".join(paths) + "\n")
print(len(paths), len(docs))
PY
    while IFS= read -r batch; do
      [[ -z "$batch" ]] && continue
      echo "  apply $(basename "$batch")"
      set +e
      out=$("$ARCTL_BIN" apply -f "$batch" --registry-url "$REGISTRY_URL" 2>&1)
      rc=$?
      set -e
      echo "$out" | sed 's/^/    /' | tail -n 20
      n=$(grep -c '^kind: Skill' "$batch" || true)
      if [[ $rc -eq 0 ]]; then
        ok=$((ok + n))
      else
        err=$((err + n))
      fi
    done < "$WORKDIR/apply-batches.txt"
  else
    set +e
    resp=$(curl -sS -w "\n%{http_code}" -X POST "$REGISTRY_URL/v0/apply" \
      -H "Content-Type: application/yaml" \
      --data-binary @"$yaml" 2>&1)
    rc=$?
    set -e
    code=$(echo "$resp" | tail -1)
    if [[ $rc -eq 0 && "$code" =~ ^2 ]]; then
      ok=$TOTAL
      echo "  HTTP $code applied $TOTAL skills"
    else
      err=$TOTAL
      echo "  ✗ apply failed HTTP ${code:-?} $(echo "$resp" | sed '$d' | head -c 300)"
    fi
  fi
}

case "$API_MODE" in
  apply)  publish_apply ;;
  legacy) publish_legacy ;;
  *) echo "Unknown API_MODE=$API_MODE" >&2; exit 1 ;;
esac

echo ""
echo "=== Summary ==="
echo "  Skill Hub: $SKILLHUB_DIR"
echo "  Selected:  $TOTAL"
echo "  OK:        $ok"
echo "  Skipped:   $skip"
echo "  Error:     $err"
echo "  API:       $API_MODE"
echo ""
if [[ "$err" -eq 0 ]]; then
  echo "UI:  $REGISTRY_URL"
  if [[ -n "$ARCTL_BIN" ]]; then
    if "$ARCTL_BIN" skill list --help >/dev/null 2>&1; then
      echo "CLI: $ARCTL_BIN skill list"
    elif "$ARCTL_BIN" get --help >/dev/null 2>&1; then
      echo "CLI: $ARCTL_BIN get skills --registry-url $REGISTRY_URL"
    fi
  fi
fi

if [[ "$err" -gt 0 ]]; then
  exit 1
fi
