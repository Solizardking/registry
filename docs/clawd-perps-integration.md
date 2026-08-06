# Clawd Perps Agent — AgentRegistry Integration

## Overview

The Clawd Perps Agent provides a unified perpetuals integration surface for Solana AI agents, using the **Rise SDK** (`@ellipsis-labs/rise`) as the primary data plane with **Vulcan CLI** fallback.

This document describes how to register and discover the perps agent through the agentregistry platform using `arctl`.

## Registered Resources

### Skill: `clawd-perps-rise-agent`

Agent skill documentation for Rise SDK integration. Points at the `.claude/skills/` documentation in the `cheshire-terminal-main` repo.

- **Source:** `https://github.com/Solizardking/agents` → `core-ai/clawd-perps-agent/.claude/skills`

### MCP Server: `clawd-perps` (17 tools)

The HTTP API exposes 17 tools for perpetuals trading:

| Tool | Description |
|------|-------------|
| `perps_list_markets` | List all available Phoenix perp markets |
| `perps_get_ticker` | Get current ticker |
| `perps_get_orderbook` | L2 orderbook |
| `perps_get_candles` | OHLCV candle history |
| `perps_get_funding` | Current funding rates |
| `perps_get_market_stats` | Full market statistics |
| `perps_get_bbo` | Best bid/offer via Hawkeye |
| `perps_get_margin` | Margin status via Hawkeye |
| `perps_get_positions` | Open positions |
| `perps_get_trader_snapshot` | Full trader state |
| `perps_market_order` | Build unsigned market order |
| `perps_limit_order` | Build unsigned limit order |
| `perps_stop_loss` | Build unsigned stop loss |
| `perps_cancel_orders` | Cancel orders |
| `perps_deposit` | Build unsigned deposit |
| `perps_withdraw` | Build unsigned withdraw |
| `imperial_health` | Check Imperial API health |

- **Endpoint:** `http://localhost:8080`
- **Transport:** Streamable HTTP (`spec.remote.type: streamable-http`)

### Agent: `clawd-perps-trader`

Bundles the MCP server with runtime safety gates, wallet integration, and risk limits.

## Architecture

```
┌──────────────────────────────────────────────┐
│  CLI / Telegram / HTTP API (server.ts)       │  ← Entry points
├──────────────────────────────────────────────┤
│  ClawdPerpsRuntime (marketMaker.ts)          │  ← Orchestrator
├────────────────────┬─────────────────────────┤
│ Rise SDK Adapter   │ Vulcan CLI Bridge       │  ← Adapters
│ (phoenixRise.ts)   │ (vulcan.ts)             │
├────────────────────┴─────────────────────────┤
│ Rise SDK / Hawkeye RPC / Vulcan CLI          │  ← Data planes
└──────────────────────────────────────────────┘
```

## Quick Start

```bash
# 1. Start the perps agent server (outside agentregistry)
cd /path/to/clawd-perps-agent
npm run build
npm start   # serves at localhost:8080

# 2. Register skill + remote MCP + agent in one multi-doc apply
cd /path/to/agentregistry
arctl apply -f examples/clawd-perps-agent.yaml

# Or register MCP and agent as separate project manifests:
arctl apply -f examples/clawd-perps-mcp/mcp.yaml
arctl apply -f examples/clawd-perps-agent-project/agent.yaml

# 3. Discover
arctl get agents | grep clawd-perps
arctl get mcps | grep clawd-perps
arctl get skills | grep clawd-perps
```

### Manifest shapes

Remote MCP servers use top-level `spec.remote` (not `spec.source.remote`):

```yaml
apiVersion: ar.dev/v1alpha1
kind: MCPServer
metadata:
  name: clawd-perps
  tag: stable
spec:
  title: Clawd Perps Agent
  remote:
    type: streamable-http
    url: http://localhost:8080
```

Agent refs must include `kind: MCPServer`:

```yaml
mcpServers:
  - kind: MCPServer
    name: clawd-perps
    tag: stable
```

## Verify the Daemon is Running

```bash
arctl version
# If daemon not running, it will auto-start
```

## Environment Configuration

The agent reads these environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `SOLANA_RPC_URL` | — | Solana RPC endpoint |
| `HELIUS_API_KEY` | — | Helius API key (optional) |
| `CLAWD_PERPS_API_URL` | `https://perp-api.phoenix.trade` | Phoenix API endpoint |
| `CLAWD_PERPS_WALLET` | — | Trader wallet pubkey |
| `LIVE_TRADING` | `false` | Enable live execution |
| `PERPS_SIM_ONLY` | `true` | Simulation-only mode |
| `PERPS_ALLOWED_SYMBOLS` | `SOL,ETH,BTC` | Allowed trading symbols |
| `PERPS_MAX_NOTIONAL_USD` | `250` | Max order size |
| `PERPS_MAX_LEVERAGE` | `3` | Max leverage |

## File Locations

| Artifact | Path |
|----------|------|
| Perps agent source (example) | external `clawd-perps-agent` project |
| MCP server manifest | `examples/clawd-perps-mcp/mcp.yaml` |
| Agent manifest | `examples/clawd-perps-agent-project/agent.yaml` |
| Multi-doc YAML (dev/test) | `examples/clawd-perps-agent.yaml` |
| Skill docs | `core-ai/clawd-perps-agent/.claude/skills/perps-rise-agent.md` |
| Examples index | `examples/README.md` |
