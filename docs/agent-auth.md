# Agent Auth (Better Auth)

AgentRegistry exposes an **[Agent Auth Protocol](https://agentauthprotocol.com)** provider via the Next.js UI using [`@better-auth/agent-auth`](https://www.better-auth.com/docs/plugins/agent-auth).

AI agents can discover the registry, register, request human approval, and execute **scoped admin capabilities** with short-lived signed JWTs.

> This plugin is under active development upstream. Report issues at [github.com/better-auth/agent-auth](https://github.com/better-auth/agent-auth).

## Features wired in this repo

| Piece | Location |
| --- | --- |
| Server plugin (`agentAuth` + `admin`) | `ui/lib/auth.ts` |
| Capability definitions + registry proxy | `ui/lib/agent-auth/capabilities.ts` |
| Better Auth HTTP handler | `ui/app/api/auth/[...all]/route.ts` |
| Discovery document | `ui/app/.well-known/agent-configuration/route.ts` |
| Browser client | `ui/lib/auth-client.ts` |
| Catalog admin agent | `examples/admin-agent.yaml` |

## Prerequisites

1. Registry server running (default `http://localhost:12121`)
2. UI packages installed (`cd ui && npm install --legacy-peer-deps`)
3. Env vars (see below)

## Environment

Add to `ui/.env.local` (or the process environment):

```bash
# Required in production
BETTER_AUTH_SECRET="$(openssl rand -hex 32)"
BETTER_AUTH_URL="http://localhost:3000"

# Where capability onExecute proxies registry HTTP calls
REGISTRY_URL="http://localhost:12121"
# or: NEXT_PUBLIC_API_URL=http://localhost:12121

# Optional SQLite path for Better Auth agent/grant tables
BETTER_AUTH_DB_PATH="./.data/better-auth.sqlite"
```

`BETTER_AUTH_SECRET` falls back to `AGENT_REGISTRY_JWT_PRIVATE_KEY` if set, then a **dev-only** default. Never ship the default.

## Database migration

After install, generate/migrate Agent Auth + Admin tables:

```bash
cd ui
npx auth migrate
# or
npx @better-auth/cli@latest migrate
```

If the CLI needs a config path:

```bash
npx @better-auth/cli@latest migrate --config lib/auth.ts
```

## Run

```bash
# Terminal 1 — registry
make run-server   # or your usual arctl-server start

# Terminal 2 — UI (serves Agent Auth)
cd ui && npm run dev
```

Discovery:

```bash
curl -s http://localhost:3000/.well-known/agent-configuration | jq .
```

Register the catalog agent:

```bash
arctl apply -f examples/admin-agent.yaml
```

## Admin capabilities

| Capability | Registry call |
| --- | --- |
| `list_agents` | `GET /v0/agents` |
| `get_agent` | `GET /v0/agents/{name}[/{tag}]` |
| `list_servers` | `GET /v0/servers` |
| `get_server` | `GET /v0/servers/{name}[/{tag}]` |
| `list_skills` | `GET /v0/skills` |
| `list_prompts` | `GET /v0/prompts` |
| `list_deployments` | `GET /v0/deployments` |
| `get_deployment` | `GET /v0/deployments/{name}?namespace=` |
| `apply_manifest` | `POST /v0/apply` |
| `delete_agent` | `DELETE /v0/agents/{name}[/{tag}]` |
| `registry_health` | `GET /v0/health` |
| `registry_version` | `GET /v0/version` |

Implementation: `ui/lib/agent-auth/capabilities.ts` → `executeRegistryCapability`.

## Flow

1. Agent reads `/.well-known/agent-configuration`
2. Lists capabilities and registers (delegated or autonomous mode)
3. Human approves grants (device auth / CIBA — plugin flows)
4. Agent signs short-lived JWTs (`aud` = execute / capability `location`)
5. `onExecute` proxies to the Go registry with the actor context

## Admin plugin (human operators)

The Better Auth **admin** plugin is enabled alongside Agent Auth so operators can:

- Assign `admin` role
- Ban / unban users who approve agents
- Impersonate for support (time-limited)

See [Better Auth Admin docs](https://www.better-auth.com/docs/plugins/admin). Default role is `user`; admin roles: `["admin"]`.

## Client usage

```ts
import { authClient } from "@/lib/auth-client";

// After sign-in as a human operator who can approve agents:
// use authClient methods exposed by agentAuthClient / adminClient
```

## Tests

```bash
cd ui
npm test -- lib/agent-auth
```

## Security notes

- Capability execute is as powerful as the registry's public authz mode allows. Pair with a non-public `AuthzProvider` in production.
- Prefer **delegated** mode so actions map to a real human user.
- Rotate `BETTER_AUTH_SECRET` independently of registry JWT keys when possible.
- SQLite under `ui/.data/` is for local agent/grant state only — not a substitute for registry Postgres.

## Related

- [Agent Auth protocol](https://agentauthprotocol.com)
- [docs/auth/authz-matrix.md](./auth/authz-matrix.md) — registry HTTP permission matrix
- [examples/admin-agent.yaml](../examples/admin-agent.yaml) — catalog entry for the admin agent
