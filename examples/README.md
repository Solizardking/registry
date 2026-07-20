# Examples

Ready-to-use declarative YAML for `arctl apply`. Every file is valid against
the `ar.dev/v1alpha1` scheme (namespace is defaulted by the server on apply).

## Quick apply

```bash
# Start the local daemon if needed
arctl version

# Minimal single-resource examples
arctl apply -f examples/mcp.yaml
arctl apply -f examples/mcp-remote.yaml
arctl apply -f examples/skill.yaml
arctl apply -f examples/prompt.yaml
arctl apply -f examples/agent.yaml

# Agent + MCP + skill + prompt in one multi-doc file
arctl apply -f examples/full-stack.yaml

# Metaplex agent (stubs satisfy MCP refs; catalog-only, no live endpoints required)
arctl apply -f examples/metaplex-mcp-stubs.yaml
arctl apply -f examples/metaplex-agent.yaml

# Clawd Perps multi-doc (catalog-only; perps HTTP server need not be up)
arctl apply -f examples/clawd-perps-agent.yaml
```

## Catalog

| File | Kind(s) | Notes |
|------|---------|--------|
| [`agent.yaml`](./agent.yaml) | Agent | Minimal agent with OCI image |
| [`mcp.yaml`](./mcp.yaml) | MCPServer | OCI package + stdio transport |
| [`mcp-remote.yaml`](./mcp-remote.yaml) | MCPServer | Pre-running remote URL (`spec.remote`) |
| [`skill.yaml`](./skill.yaml) | Skill | Git-backed skill package |
| [`prompt.yaml`](./prompt.yaml) | Prompt | Inline system prompt |
| [`full-stack.yaml`](./full-stack.yaml) | MCPServer + Skill + Prompt + Agent | Dependencies first, then agent |
| [`metaplex-mcp-stubs.yaml`](./metaplex-mcp-stubs.yaml) | MCPServer | Catalog stubs for metaplex agent MCP refs |
| [`metaplex-agent.yaml`](./metaplex-agent.yaml) | Agent | Metaplex on-chain identity fields |
| [`clawd-perps-agent.yaml`](./clawd-perps-agent.yaml) | Skill + MCPServer + Agent | Multi-doc Clawd Perps stack |
| [`clawd-perps-mcp/mcp.yaml`](./clawd-perps-mcp/mcp.yaml) | MCPServer | Remote perps MCP (localhost:8080) |
| [`clawd-perps-agent-project/agent.yaml`](./clawd-perps-agent-project/agent.yaml) | Agent | Agent-only project referencing clawd-perps |

## Robinhood Agents forge catalog

Import the dual-chain forge package (`agents/*.json` + examples) under tag `robinhood`:

```bash
python3 scripts/import-robinhood-agents.py \
  --source /path/to/robinhood-agents \
  --apply --registry-url http://localhost:12121
```

Batches: [`robinhood-agents-catalog/`](./robinhood-agents-catalog/).

## Cheshire Terminal catalog (bulk)

Import the full Cheshire Terminal Solana agents hub (143 agents) as catalog
entries:

```bash
python3 scripts/import-cheshire-agents.py \
  --source /path/to/cheshire-terminal/agents \
  --apply --registry-url http://localhost:12121
```

Checked-in batches live under [`cheshire-terminal-catalog/`](./cheshire-terminal-catalog/).

## Clawd Perps

See [docs/clawd-perps-integration.md](../docs/clawd-perps-integration.md).

```bash
# 1. Start the perps agent HTTP server (outside this repo)
cd /path/to/clawd-perps-agent && npm start

# 2. Register everything in one shot
arctl apply -f examples/clawd-perps-agent.yaml

# Or register MCP + agent separately
arctl apply -f examples/clawd-perps-mcp/mcp.yaml
arctl apply -f examples/clawd-perps-agent-project/agent.yaml
```

## Schema reminders

- **Remote MCP**: use top-level `spec.remote` (not `spec.source.remote`).
- **Agent MCP refs**: every entry needs `kind: MCPServer` plus `name` (and optional `tag`).
- **Multi-doc YAML**: do not put a bare `---` before the first document when leading comments would form an empty first doc — start the first resource, then use `---` between subsequent resources.
- **Skills / Prompts**: catalog resources only; no `arctl build` step.
- **Agents / source MCP servers**: `arctl build ./project` needs an `arctl init` scaffold with `arctl.yaml` + framework code. Catalog-only manifests (image or remote URL) just need `arctl apply`.
