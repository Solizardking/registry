# Cheshire Agent Registry deploy

## Live endpoints

| Surface | URL |
|---------|-----|
| **API + embedded UI (Fly)** | https://registry.cheshireterminal.ai |
| **Health** | https://registry.cheshireterminal.ai/v0/health |
| **Cheshire SPA page** | https://cheshireterminal.ai/agent-registry |
| **Alias path** | https://cheshireterminal.ai/registry |
| **Same-origin proxy** | https://cheshire-clawd-terminal.fly.dev/api/agent-registry/* |
| **Vercel UI (catalog)** | see Vercel project `agentregistry-ui` (points at Fly API) |

## Source / catalog links

| Resource | URL |
|----------|-----|
| This registry codebase | https://github.com/Solizardking/agentregistry-main |
| Registry alias repo | https://github.com/Solizardking/registry |
| Agents catalog / forge | https://github.com/Solizardking/agents |
| Skills hub | https://github.com/Solizardking/skills |
| Agent Auth docs | [docs/agent-auth.md](./agent-auth.md) |

Default agent `source.repository.url` in catalog YAML and import scripts is
`https://github.com/Solizardking/agents`. Skill catalogs still use skills/skillhub hosts.

## Apps

- `cheshire-agentregistry` — Go `arctl-server` + embedded Next UI (**backend**)
- `cheshire-agentregistry-db` — Fly Postgres (flex)
- `cheshire-clawd-terminal` — sets `AGENT_REGISTRY_URL=https://registry.cheshireterminal.ai`
- Vercel `agentregistry-ui` — Next catalog UI (optional frontend mirror)

## Deploy / redeploy

```bash
cd /Users/8bit/agentregistry-main
fly deploy -c fly.toml --remote-only -a cheshire-agentregistry
```

Or: `./scripts/deploy-fly.sh`

## Publish catalog YAML

```bash
curl -X POST https://registry.cheshireterminal.ai/v0/apply \
  -H 'Content-Type: application/yaml' \
  --data-binary @examples/clawd-perps-agent.yaml
```

Catalog batches live under `examples/cheshire-terminal-catalog/`.

## Env (server)

| Variable | Notes |
|----------|--------|
| `AGENT_REGISTRY_DATABASE_URL` | Set via entrypoint from Fly `DATABASE_URL` |
| `AGENT_REGISTRY_JWT_PRIVATE_KEY` | 32-byte hex secret |
| `AGENT_REGISTRY_SERVER_ADDRESS` | `:8080` |
| `AGENT_REGISTRY_PLATFORM_MODE` | `docker` |

## Cheshire wiring

- Proxy: `server/routes/agent-registry.ts` → `/api/agent-registry`
- SPA: `client/src/pages/AgentRegistryPage.tsx` → `/agent-registry`
- Nav: Agents menu → **Agent Registry**
