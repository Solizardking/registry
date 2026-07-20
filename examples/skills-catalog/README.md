# Skills catalog

Generated `ar.dev/v1alpha1` Skill manifests imported from Cheshire Terminal skill
packages (`SKILL.md` trees).

## Sources

| Catalog label | Path under monorepo | Count (typical) |
|---------------|---------------------|-----------------|
| `agents-skills` | `agents/skills/*` | ~21 (Pump, Robinhood forge, sponge-wallet) |
| `robinhood-agents-skills` | `robinhood-agents/skills/*` | ~1 |
| `cheshire-skills` | `skills/*` (top-level) | ~179 |

Nested packs under `skills/**` (e.g. `deprecated/`, vendor copies) are skipped
unless you pass `--include-nested`.

## Apply

```bash
# Local registry
python3 scripts/import-skills.py \
  --source-root /path/to/cheshire-terminal \
  --apply --registry-url http://localhost:12121

# Fly production
python3 scripts/import-skills.py \
  --source-root /path/to/cheshire-terminal \
  --apply --registry-url https://cheshire-agentregistry.fly.dev

# Or apply checked-in batches:
for f in examples/skills-catalog/skills-batch-*.yaml; do
  arctl apply -f "$f" --registry-url https://cheshire-agentregistry.fly.dev
done
```

## Inspect

```bash
arctl get skills --registry-url https://cheshire-agentregistry.fly.dev
arctl get skill robinhood-agent-forge --tag stable -o yaml
arctl get skill pump-sdk-core --tag stable -o yaml
```

## Notes

- Catalog registration only — skills are git-backed documentation packages.
- `spec.source.repository` points at the monorepo + `subfolder` for the skill tree.
- Tag defaults to `stable`.
