# Cheshire Terminal agents catalog

Generated `ar.dev/v1alpha1` Agent manifests imported from the Cheshire Terminal
Solana agents hub (`agents-catalog.json` + per-agent JSON).

## Apply to a live registry

```bash
# Start a registry that speaks declarative apply (current arctl-server), then:
python3 scripts/import-cheshire-agents.py \
  --source /path/to/cheshire-terminal/agents \
  --apply --registry-url http://localhost:12121

# Or apply the checked-in batches:
for f in examples/cheshire-terminal-catalog/agents-batch-*.yaml; do
  arctl apply -f "$f"
done
```

## Contents

| Artifact | Purpose |
|----------|---------|
| `agents-batch-*.yaml` | Multi-doc Agent YAML (25 agents per file) |
| `import-manifest.json` | Counts, tag, batch list, source path |
| `../../scripts/import-cheshire-agents.py` | Regenerator / applier |

## Notes

- **Catalog registration only** — no container deploy or live inference.
- Agents use tag `stable` and labels `app=cheshire-terminal`, `catalog=cheshire-terminal-agents`.
- Image refs are catalog placeholders (`ghcr.io/solizardking/clawd-agents/<name>:<version>`) unless the source JSON defines an image.
- Model strings like `clawdrouter/auto` map to `modelProvider` / `modelName`.
- Re-apply is idempotent (`created` → `unchanged` / `configured`).
