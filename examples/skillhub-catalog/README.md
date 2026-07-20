# SkillHub catalog

Full [SkillHub](https://github.com/Solizardking/skillhub-main) import (584 skills
from `catalog.json`) as `ar.dev/v1alpha1` Skill resources.

Upstream: https://github.com/Solizardking/skillhub-main · site: https://skills.x402.wtf

## Apply

```bash
# From a clone of the hub
git clone --depth 1 https://github.com/Solizardking/skillhub-main.git /tmp/skillhub-main

python3 scripts/import-skillhub.py \
  --source /tmp/skillhub-main \
  --repo-url https://github.com/Solizardking/skillhub-main \
  --apply --registry-url https://cheshire-agentregistry.fly.dev

# Or apply checked-in batches:
for f in examples/skillhub-catalog/skills-batch-*.yaml; do
  arctl apply -f "$f" --registry-url https://cheshire-agentregistry.fly.dev
done
```

## Tag strategy

| Catalog | Tag | Notes |
|---------|-----|--------|
| Cheshire monorepo top-level | `stable` | `examples/skills-catalog/` |
| SkillHub full hub | `skillhub` | This directory — path-based names (`google-cloud-gke-basics`) |

```bash
arctl get skills --tag skillhub
arctl get skill metaplex-agent --tag skillhub -o yaml
arctl get skill nvidia-jetson-quick-start --tag skillhub -o yaml
```

## Zones (from catalog.json)

NVIDIA · Solana/Blockchain · Dev Tools/Agents · Google Cloud/Ads/Analytics ·
Media · Productivity · Utilities · Local/Web Services

## Notes

- Names are DNS-safe slugs: `/` → `-` (e.g. `anthropic-skills/claude-api` → `anthropic-skills-claude-api`).
- `spec.source.repository` → `https://github.com/Solizardking/skillhub-main` + `skills/<slug>`.
- Labels: `catalog=skillhub`, `category=<zone>`, `namespace=<first path segment>`.
