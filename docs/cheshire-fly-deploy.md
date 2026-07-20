# Cheshire Agent Registry on Fly

## Live endpoints

| Surface | URL |
|---------|-----|
| **API + native UI** | https://cheshire-agentregistry.fly.dev |
| **Health** | https://cheshire-agentregistry.fly.dev/v0/health |
| **Cheshire SPA page** | https://cheshireterminal.ai/agent-registry |
| **Same-origin proxy** | https://cheshire-clawd-terminal.fly.dev/api/agent-registry/* |

## Apps

- `cheshire-agentregistry` — Go `arctl-server` + embedded Next UI
- `cheshire-agentregistry-db` — Fly Postgres (flex)
- `cheshire-clawd-terminal` — sets `AGENT_REGISTRY_URL=https://cheshire-agentregistry.fly.dev`

## Deploy / redeploy

```bash
cd /Users/8bit/agentregistry-main
fly deploy -c fly.toml --remote-only -a cheshire-agentregistry
```

Or: `./scripts/deploy-fly.sh`

## Publish catalog YAML

```bash
curl -X POST https://cheshire-agentregistry.fly.dev/v0/apply \
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
