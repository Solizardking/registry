# Metaplex Integration

The AgentRegistry integrates with the **Metaplex protocol stack** on Solana to provide on-chain identity, token launch, NFT creation, and execution delegation for AI agents. This mirrors the full Metaplex agent ecosystem at [metaplex.com/agents](https://metaplex.com/agents).

## Overview

The integration covers six Metaplex programs:

| Program | Purpose | CLI Support | MCP Tools |
|---------|---------|-------------|-----------|
| **Agent Registry** | On-chain agent identity, wallets, execution delegation | `mplx agents` | `mint_agent`, `register_agent_identity`, `read_agent_data`, `set_agent_token`, `register_executive`, `delegate_execution`, `revoke_execution` |
| **MPL Core** | Next-gen NFT standard (cheaper than Token Metadata) | `mplx core` | `create_core_nft` |
| **Genesis** | Token launch with bonding curves and launchpools | `mplx genesis` | `launch_agent_token`, `set_agent_token` |
| **Bubblegum** | Compressed NFTs at massive scale | `mplx bg` | — |
| **Token Metadata** | Original NFT standard (fungible tokens, NFTs, pNFTs) | `mplx tm` | — |
| **Candy Machine** | NFT drops with minting rules/guards | `mplx cm` | — |

## Key Concepts

### Agent Identity

Every agent registered via Metaplex gets:
- **MPL Core Asset**: An on-chain NFT representing the agent
- **AgentIdentity PDA**: A derived address for discoverability (`["agent_identity", asset]`)
- **Asset Signer PDA**: The agent's built-in wallet (`["mpl-core-execute", asset]`) — no private key exists
- **EIP-8004 Document**: Off-chain JSON metadata describing the agent's services, trust model, and registrations

### Agent Token

An agent can launch a canonical token via the **Genesis** protocol:
- **Bonding Curve** (default): Constant-product AMM that starts trading immediately and auto-graduates to Raydium CPMM
- **Launchpool**: Configurable allocations with a 48-hour deposit window
- **setAgentTokenV1**: Irreversibly binds the token mint to the agent (one token per agent forever)

### Execution Delegation

Agents use **mpl-agent-tools** for off-chain operation:
- **Executive Profile**: One-time registration per wallet (PDA at `["executive_profile", authority]`)
- **Delegate Execution**: Links an agent to an executive via `ExecutionDelegateRecordV1`
- **Revoke Execution**: Closes the delegation record to stop future Execute calls

## Configuration

### Environment Variables

```bash
# Solana RPC endpoints
AGENT_REGISTRY_SOLANA_RPC_URL=https://api.mainnet-beta.solana.com
AGENT_REGISTRY_SOLANA_DEVNET_RPC_URL=https://api.devnet.solana.com

# Metaplex API
AGENT_REGISTRY_METAPLEX_API_BASE_URL=https://api.metaplex.com

# Solana keypair file (for signing transactions)
AGENT_REGISTRY_SOLANA_KEYPAIR_PATH=~/.config/solana/id.json

# Irys gateway (for token image uploads)
AGENT_REGISTRY_IRYS_GATEWAY_URL=https://gateway.irys.xyz
```

## CLI Usage

The `arctl metaplex` command provides access to all Metaplex operations:

```bash
# Mint an agent (create Core asset + register identity)
arctl metaplex mint-agent "My Agent" \
  --uri "https://arweave.net/metadata.json" \
  --network solana-mainnet

# Register identity on existing Core asset
arctl metaplex register-identity <ASSET_ADDRESS> \
  --registration-uri "https://arweave.net/registration.json" \
  --collection <COLLECTION_ADDRESS>

# Read agent data from on-chain
arctl metaplex read-agent <ASSET_ADDRESS> --network solana-mainnet

# Launch a token via Genesis bonding curve
arctl metaplex launch-token "Agent Token" "AGT" \
  --image "https://gateway.irys.xyz/image-id" \
  --network solana-mainnet

# Launch an agent token (routes fees to agent PDA)
arctl metaplex launch-token "Agent Token" "AGT" \
  --image "https://gateway.irys.xyz/image-id" \
  --agent-asset <AGENT_ASSET_ADDRESS> \
  --agent-set-token \
  --network solana-mainnet

# Permanently link a token to an agent
arctl metaplex set-agent-token <AGENT_ASSET> <GENESIS_ACCOUNT>

# Register an executive profile
arctl metaplex register-executive --network solana-mainnet

# Delegate execution to an executive
arctl metaplex delegate-execution <AGENT_ASSET> <EXECUTIVE_ADDRESS>

# Revoke execution delegation
arctl metaplex revoke-execution <AGENT_ASSET> <EXECUTIVE_ADDRESS>

# Create an MPL Core NFT
arctl metaplex create-nft "My NFT" \
  --uri "https://arweave.net/metadata.json" \
  --collection <COLLECTION_ADDRESS>

# Create an MPL Core collection
arctl metaplex create-collection "My Collection" \
  --uri "https://arweave.net/collection-metadata.json"
```

## MCP Tools

When the MCP server is configured with Metaplex settings, the following additional tools are available through the MCP bridge:

### Agent Operations
- `mint_agent` — Mint a Core asset + register agent identity in one transaction
- `register_agent_identity` — Attach AgentIdentity plugin to an existing Core asset
- `read_agent_data` — Read agent identity PDA, wallet, and balance from on-chain
- `set_agent_token` — Permanently link a Genesis token to an agent (irreversible)

### Token Operations
- `launch_agent_token` — Launch a token via Genesis protocol (bonding curve or launchpool)

### Executive Operations
- `register_executive` — Register an executive operator profile
- `delegate_execution` — Delegate execution from agent to executive
- `revoke_execution` — Revoke execution delegation

### NFT Operations
- `create_core_nft` — Create an MPL Core NFT

## Agent YAML Example

```yaml
apiVersion: ar.dev/v1alpha1
kind: Agent
metadata:
  name: metaplex-trader
  labels:
    app.kubernetes.io/metaplex: "true"
spec:
  title: "Metaplex Trading Agent"
  description: "An autonomous agent on Solana with Metaplex identity"
  modelProvider: gemini
  modelName: gemini-2.5-flash
  source:
    image: ghcr.io/acme/metaplex-trader:v1.0.0
  metaplex:
    coreAssetAddress: "<MPL_CORE_ASSET>"
    agentIdentityPDA: "<AGENT_IDENTITY_PDA>"
    assetSignerPDA: "<ASSET_SIGNER_PDA>"
    agentTokenMintAddress: "<TOKEN_MINT>"
    agentRegistrationURI: "https://arweave.net/agent-registration.json"
    network: "solana-mainnet"
    launchType: "bondingCurve"
    tokenMetadata:
      name: "Metaplex Trader Token"
      symbol: "MPT"
      image: "https://gateway.irys.xyz/image-id"
    registrationDocument:
      type: "https://eips.ethereum.org/EIPS/eip-8004#registration-v1"
      name: "Metaplex Trading Agent"
      description: "An autonomous agent executing DeFi strategies on Solana"
      image: "https://arweave.net/agent-avatar.png"
      services:
        - name: "web"
          endpoint: "https://metaplex.com/agent/<ASSET_ADDRESS>"
      active: true
      registrations:
        - agentId: "<MINT_ADDRESS>"
          agentRegistry: "solana:101:metaplex"
      supportedTrust:
        - reputation
        - crypto-economic
```

## Quick Reference

| Operation | `mplx` CLI | `arctl metaplex` | MCP Tool |
|-----------|-----------|------------------|----------|
| Mint agent | `mplx genesis launch create --agentAsset` | `mint-agent` | `mint_agent` |
| Register identity | `mplx agents register --use-ix` | `register-identity` | `register_agent_identity` |
| Fetch agent | `mplx agents fetch` | `read-agent` | `read_agent_data` |
| Set agent token | `mplx agents set-agent-token` | `set-agent-token` | `set_agent_token` |
| Register executive | `mplx agents executive register` | `register-executive` | `register_executive` |
| Delegate | `mplx agents executive delegate` | `delegate-execution` | `delegate_execution` |
| Revoke | `mplx agents executive revoke` | `revoke-execution` | `revoke_execution` |
| Create Core NFT | `mplx core create` | `create-nft` | `create_core_nft` |
| Create collection | `mplx core collection create` | `create-collection` | — |
| Launch token | `mplx genesis launch create` | `launch-token` | `launch_agent_token` |

## Related Resources

- [Metaplex Agent Documentation](https://metaplex.com/agents)
- [Metaplex Agent Registry SDK](https://github.com/metaplex-foundation/mpl-agent-registry)
- [Genesis Protocol](https://metaplex.com/genesis)
- [MPL Core](https://metaplex.com/core)
- [Solana Developer Docs](https://docs.solana.com)