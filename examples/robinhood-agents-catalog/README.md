# Robinhood Agents catalog

Generated `ar.dev/v1alpha1` Agent manifests imported from the
[Robinhood Agents](https://github.com/Solizardking/agents) forge package
(`agents/*.json` + dual-chain `examples/`).

## Apply to a live registry

```bash
python3 scripts/import-robinhood-agents.py \
  --source /path/to/robinhood-agents \
  --apply --registry-url http://localhost:12121

# Or apply checked-in batches:
for f in examples/robinhood-agents-catalog/agents-batch-*.yaml; do
  arctl apply -f "$f"
done
```

## Tag strategy

| Catalog | Tag | Notes |
|---------|-----|--------|
| Cheshire Terminal hub | `stable` | `examples/cheshire-terminal-catalog/` |
| Robinhood Agents forge | `robinhood` | This directory — same identifiers, different tag |

Shared names (e.g. `airdrop-hunter`) therefore appear as distinct rows:

```bash
arctl get agent airdrop-hunter --tag stable
arctl get agent airdrop-hunter --tag robinhood
```

## Notes

- Catalog registration only — no EVM broadcast, no Solana mint, no deploy.
- Labels: `app=robinhood-agents`, `catalog=robinhood-agents`, `forge=cheshire-terminal`.
- Character personas (`*character*`, investor characters) get `kind=character`.
- Forge examples become `robinhood-open-research-agent` / `solana-…` name slugs.
