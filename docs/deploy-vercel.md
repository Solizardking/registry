# Deploy AgentRegistry UI to Vercel

The **API backend** runs on Fly (`cheshire-agentregistry` →
https://registry.cheshireterminal.ai). The optional **Next.js catalog UI** can
mirror on Vercel and talk to that API.

## Prerequisites

- Vercel CLI logged in (`vercel whoami`)
- Fly registry healthy: `curl -s https://registry.cheshireterminal.ai/v0/health`

## One-shot deploy (from repo)

```bash
cd ui
vercel link --yes --project agentregistry-ui   # first time
vercel env add NEXT_PUBLIC_API_URL production  # https://registry.cheshireterminal.ai
vercel env add NEXT_PUBLIC_GATEWAY_URL production
vercel env add BETTER_AUTH_URL production      # https://<your-ui>.vercel.app
vercel env add BETTER_AUTH_SECRET production   # openssl rand -hex 32
vercel --prod
```

Or from repo root:

```bash
vercel --cwd ui --prod \
  --env NEXT_PUBLIC_API_URL=https://registry.cheshireterminal.ai \
  --env NEXT_PUBLIC_GATEWAY_URL=https://registry.cheshireterminal.ai \
  --env BETTER_AUTH_URL=https://agentregistry-ui.vercel.app
```

## Domain mapping

| Host / path | Target |
|-------------|--------|
| `registry.cheshireterminal.ai` | Fly app `cheshire-agentregistry` (API + embedded UI) |
| `cheshireterminal.ai/registry` | Cheshire Terminal SPA route / proxy |
| `cheshireterminal.ai/agent-registry` | Cheshire Terminal SPA route |
| Vercel project domain | Catalog UI → API at `NEXT_PUBLIC_API_URL` |

To attach a custom host (e.g. `ui.registry.cheshireterminal.ai`):

```bash
vercel domains add ui.registry.cheshireterminal.ai --project agentregistry-ui
```

Then add a CNAME at Cloudflare → `cname.vercel-dns.com`.

## GitHub connections

| Repo | Role |
|------|------|
| [Solizardking/registry](https://github.com/Solizardking/registry) | This control-plane code (alias of agentregistry-main) |
| [Solizardking/agentregistry-main](https://github.com/Solizardking/agentregistry-main) | Primary git remote |
| [Solizardking/agents](https://github.com/Solizardking/agents) | Default agent source / forge catalog |
| [Solizardking/skills](https://github.com/Solizardking/skills) | Skills hub |

Import scripts and example agent YAMLs default `repository.url` to
`https://github.com/Solizardking/agents`.

## Agent Auth note

Better Auth Agent Auth uses SQLite locally (`BETTER_AUTH_DB_PATH`). On Vercel
serverless, prefer pointing Agent Auth at the Fly-hosted UI or use an external
Postgres `DATABASE_URL` later. Catalog browse/apply works with only
`NEXT_PUBLIC_API_URL` set.

## Fly backend redeploy

```bash
fly deploy -c fly.toml --remote-only -a cheshire-agentregistry
# or
./scripts/deploy-fly.sh
```
