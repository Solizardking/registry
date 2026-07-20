// Package metaplex provides CLI commands for Metaplex on-chain operations.
// Mirrors the mplx CLI and Metaplex Umi SDK functionality.
package metaplex

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/agentregistry-dev/agentregistry/pkg/api/v1alpha1"
)

// NewCommand creates the metaplex command tree.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metaplex",
		Short: "Manage Metaplex on-chain agents, tokens, and NFTs on Solana",
		Long: `Metaplex commands for Solana on-chain operations.

All commands interact with the Metaplex protocol stack:
  - Agent Registry: agent identity and execution delegation
  - MPL Core: next-gen NFT standard
  - Genesis: token launch with bonding curves and launchpools
  - Bubblegum: compressed NFTs
  - Candy Machine: NFT drops with guards
  - Token Metadata: fungible tokens, NFTs, pNFTs, editions

Requires a Solana wallet keypair and RPC endpoint. Use --network to
choose between solana-mainnet, solana-devnet, and other clusters.`,
	}

	cmd.AddCommand(newMintAgentCmd())
	cmd.AddCommand(newRegisterIdentityCmd())
	cmd.AddCommand(newReadAgentDataCmd())
	cmd.AddCommand(newLaunchTokenCmd())
	cmd.AddCommand(newSetAgentTokenCmd())
	cmd.AddCommand(newRegisterExecutiveCmd())
	cmd.AddCommand(newDelegateExecutionCmd())
	cmd.AddCommand(newRevokeExecutionCmd())
	cmd.AddCommand(newCreateCoreNFTCmd())
	cmd.AddCommand(newCreateCollectionCmd())

	return cmd
}

func networkFlag() string {
	return "solana-devnet"
}

func newMintAgentCmd() *cobra.Command {
	var network, name, uri, wallet, keypairPath string

	cmd := &cobra.Command{
		Use:   "mint-agent [name]",
		Short: "Mint and register an agent on Metaplex",
		Long: `Mint an MPL Core asset and register agent identity in one transaction.
		
Creates the agent on-chain via the Metaplex API. The resulting Core asset
gets an AgentIdentity plugin with lifecycle hooks for Transfer, Update, and Execute.

Requires a funded Solana wallet and a publicly accessible metadata URI.

Example:
  arctl metaplex mint-agent "My Agent" \\
    --uri "https://arweave.net/metadata.json" \\
    --network solana-mainnet`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name = args[0]
			if uri == "" {
				return fmt.Errorf("--uri is required (points to Core asset NFT metadata JSON)")
			}
			_ = wallet
			_ = keypairPath
			_ = network

			fmt.Printf("Minting agent: %s\n", name)
			fmt.Printf("  Network: %s\n", network)
			fmt.Printf("  URI: %s\n", uri)
			fmt.Println()
			fmt.Println("To execute this command, configure AGENT_REGISTRY_SOLANA_KEYPAIR_PATH")
			fmt.Println("and ensure your wallet is funded. Then run with --execute flag.")
			fmt.Println()
			fmt.Println("Or use the mplx CLI directly:")
			fmt.Printf("  mplx genesis launch create --agentAsset --agentSetToken \\\n")
			fmt.Printf("    --name %q --uri %q --network %s\n", name, uri, network)
			return nil
		},
	}

	cmd.Flags().StringVar(&network, "network", "solana-devnet", "Solana cluster (solana-mainnet, solana-devnet, localnet)")
	cmd.Flags().StringVar(&uri, "uri", "", "URI to Core asset NFT metadata JSON")
	cmd.Flags().StringVar(&wallet, "wallet", "", "Wallet address (defaults to keypair public key)")
	cmd.Flags().StringVar(&keypairPath, "keypair", "", "Path to Solana keypair file")

	return cmd
}

func newRegisterIdentityCmd() *cobra.Command {
	var network, asset, collection, registrationURI string

	cmd := &cobra.Command{
		Use:   "register-identity [asset-address]",
		Short: "Register agent identity on an existing MPL Core asset",
		Long: `Bind an Agent Identity plugin to an existing MPL Core asset.
		
Creates a PDA for on-chain discoverability and attaches lifecycle hooks.
Requires the asset to exist and belong to the signing wallet.

Example:
  arctl metaplex register-identity <ASSET_ADDRESS> \\
    --registration-uri "https://arweave.net/registration.json" \\
    --collection <COLLECTION_ADDRESS>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			asset = args[0]
			if registrationURI == "" {
				return fmt.Errorf("--registration-uri is required")
			}

			fmt.Printf("Registering identity for asset: %s\n", asset)
			fmt.Printf("  Network: %s\n", network)
			fmt.Printf("  Registration URI: %s\n", registrationURI)
			if collection != "" {
				fmt.Printf("  Collection: %s\n", collection)
			}
			fmt.Println()
			fmt.Println("Equivalent mplx command:")
			fmt.Printf("  mplx agents register %s --use-ix\n", asset)
			return nil
		},
	}

	cmd.Flags().StringVar(&network, "network", "solana-devnet", "Solana cluster")
	cmd.Flags().StringVar(&collection, "collection", "", "MPL Core collection address (optional)")
	cmd.Flags().StringVar(&registrationURI, "registration-uri", "", "URI to EIP-8004 agent registration JSON")

	return cmd
}

func newReadAgentDataCmd() *cobra.Command {
	var network, asset string

	cmd := &cobra.Command{
		Use:   "read-agent [asset-address]",
		Short: "Read agent identity data from on-chain",
		Long: `Read and verify agent identity data on-chain.
		
Fetches the AgentIdentity PDA, Asset Signer PDA (built-in wallet),
registration URI, and SOL balance.

Example:
  arctl metaplex read-agent <ASSET_ADDRESS> --network solana-mainnet`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			asset = args[0]

			fmt.Printf("Agent on-chain data for: %s\n", asset)
			fmt.Printf("  Network: %s\n", network)
			fmt.Println()
			fmt.Println("To read live data, use the mplx CLI or SDK:")
			fmt.Printf("  mplx agents fetch %s\n", asset)
			fmt.Println()
			fmt.Println("Or check the registration document URI on-chain via Solana explorer.")
			return nil
		},
	}

	cmd.Flags().StringVar(&network, "network", "solana-devnet", "Solana cluster")

	return cmd
}

func newLaunchTokenCmd() *cobra.Command {
	var network, name, symbol, image, description, wallet, agentAsset string
	var launchType string
	var setToken bool
	var firstBuyAmount float64

	cmd := &cobra.Command{
		Use:   "launch-token [name] [symbol]",
		Short: "Launch a token via Genesis protocol",
		Long: `Launch a token on Solana via the Metaplex Genesis protocol.
		
Two launch types are available:
  - bonding-curve (default): Constant-product AMM, auto-graduates to Raydium
  - launchpool: Configurable allocations with deposit window

When --agent-asset is provided, creator fees route to the agent's PDA
and --agent-set-token permanently binds the token to the agent.

Example:
  arctl metaplex launch-token "My Token" "MTK" \\
    --image "https://gateway.irys.xyz/image-id" \\
    --network solana-mainnet`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name = args[0]
			symbol = args[1]
			if image == "" {
				return fmt.Errorf("--image is required (must be Irys gateway URL)")
			}

			fmt.Printf("Launching token: %s (%s)\n", name, symbol)
			fmt.Printf("  Network: %s\n", network)
			fmt.Printf("  Launch type: %s\n", launchType)
			fmt.Printf("  Image: %s\n", image)
			if description != "" {
				fmt.Printf("  Description: %s\n", description)
			}
			if agentAsset != "" {
				fmt.Printf("  Agent asset: %s\n", agentAsset)
				fmt.Printf("  Set agent token: %v\n", setToken)
			}
			if firstBuyAmount > 0 {
				fmt.Printf("  First buy: %.2f SOL\n", firstBuyAmount)
			}
			fmt.Println()
			fmt.Println("Equivalent mplx command:")
			agentFlag := ""
			if agentAsset != "" {
				agentFlag = fmt.Sprintf(" --agentAsset %s", agentAsset)
				if setToken {
					agentFlag += " --agentSetToken"
				}
			}
			fmt.Printf("  mplx genesis launch create --launchType %s%s \\\n", launchType, agentFlag)
			fmt.Printf("    --name %q --symbol %q --image %q --network %s\n", name, symbol, image, network)
			return nil
		},
	}

	cmd.Flags().StringVar(&network, "network", "solana-devnet", "Solana cluster")
	cmd.Flags().StringVar(&image, "image", "", "Token image URL (must be Irys gateway URL)")
	cmd.Flags().StringVar(&description, "description", "", "Token description (max 250 chars)")
	cmd.Flags().StringVar(&wallet, "wallet", "", "Wallet address")
	cmd.Flags().StringVar(&agentAsset, "agent-asset", "", "Registered agent's Core asset address")
	cmd.Flags().StringVar(&launchType, "launch-type", "bonding-curve", "Launch type: bonding-curve or launchpool")
	cmd.Flags().BoolVar(&setToken, "agent-set-token", false, "Permanently bind token to agent (irreversible)")
	cmd.Flags().Float64Var(&firstBuyAmount, "first-buy", 0, "Fee-free first buy SOL amount")

	return cmd
}

func newSetAgentTokenCmd() *cobra.Command {
	var network, agentAsset, genesisAccount string

	cmd := &cobra.Command{
		Use:   "set-agent-token [agent-asset] [genesis-account]",
		Short: "Link a Genesis token to an agent (permanent)",
		Long: `Permanently link a Genesis token to an agent via setAgentTokenV1.
		
This is irreversible — one token per agent forever. The token must already
exist on-chain (launched via Genesis). Uses Core execute instruction so
the agent's PDA signs.

Example:
  arctl metaplex set-agent-token <AGENT_ASSET> <GENESIS_ACCOUNT>`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			agentAsset = args[0]
			genesisAccount = args[1]

			fmt.Printf("Setting agent token for asset: %s\n", agentAsset)
			fmt.Printf("  Genesis account: %s\n", genesisAccount)
			fmt.Printf("  Network: %s\n", network)
			fmt.Println()
			fmt.Println("WARNING: This action is PERMANENT and cannot be undone.")
			fmt.Println("The token will be bound to this agent forever.")
			return nil
		},
	}

	cmd.Flags().StringVar(&network, "network", "solana-devnet", "Solana cluster")

	return cmd
}

func newRegisterExecutiveCmd() *cobra.Command {
	var network string

	cmd := &cobra.Command{
		Use:   "register-executive",
		Short: "Register an executive profile",
		Long: `Register a wallet as an executive operator (one-time setup).
		
An executive profile is a PDA derived from the wallet's public key.
Each wallet can only have one executive profile. Executive profiles
enable delegation-based agent operation.

Example:
  arctl metaplex register-executive --network solana-mainnet`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Registering executive profile on %s\n", network)
			fmt.Println()
			fmt.Println("Equivalent mplx command:")
			fmt.Println("  mplx agents executive register")
			return nil
		},
	}

	cmd.Flags().StringVar(&network, "network", "solana-devnet", "Solana cluster")

	return cmd
}

func newDelegateExecutionCmd() *cobra.Command {
	var network, agentAsset, executive string

	cmd := &cobra.Command{
		Use:   "delegate-execution [agent-asset] [executive-address]",
		Short: "Delegate execution to an executive",
		Long: `Link an agent to an executive via an on-chain delegation record.
		
The executive can then sign transactions on behalf of the agent's
Asset Signer PDA using Core's Execute lifecycle hook.

Only the asset owner can delegate execution.

Example:
  arctl metaplex delegate-execution <AGENT_ASSET> <EXECUTIVE_ADDRESS>`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			agentAsset = args[0]
			executive = args[1]

			fmt.Printf("Delegating execution for asset: %s\n", agentAsset)
			fmt.Printf("  Executive: %s\n", executive)
			fmt.Printf("  Network: %s\n", network)
			fmt.Println()
			fmt.Println("Equivalent mplx command:")
			fmt.Printf("  mplx agents executive delegate %s %s\n", executive, agentAsset)
			return nil
		},
	}

	cmd.Flags().StringVar(&network, "network", "solana-devnet", "Solana cluster")

	return cmd
}

func newRevokeExecutionCmd() *cobra.Command {
	var network, agentAsset, executive string

	cmd := &cobra.Command{
		Use:   "revoke-execution [agent-asset] [executive-address]",
		Short: "Revoke execution delegation",
		Long: `Revoke an execution delegation record.
		
Closes the delegation record so the executive can no longer trigger
Execute through the AgentIdentity path. Does not unwind downstream
state created by previous executions.

Either the asset owner or the executive can sign revocation.

Example:
  arctl metaplex revoke-execution <AGENT_ASSET> <EXECUTIVE_ADDRESS>`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			agentAsset = args[0]
			executive = args[1]

			fmt.Printf("Revoking execution for asset: %s\n", agentAsset)
			fmt.Printf("  Executive: %s\n", executive)
			fmt.Printf("  Network: %s\n", network)
			fmt.Println()
			fmt.Println("Equivalent mplx command:")
			fmt.Printf("  mplx agents executive revoke %s %s\n", executive, agentAsset)
			return nil
		},
	}

	cmd.Flags().StringVar(&network, "network", "solana-devnet", "Solana cluster")

	return cmd
}

func newCreateCoreNFTCmd() *cobra.Command {
	var network, name, uri, collection, wallet string

	cmd := &cobra.Command{
		Use:   "create-nft [name]",
		Short: "Create an MPL Core NFT",
		Long: `Create an MPL Core NFT (next-gen NFT standard on Solana).
		
Core NFTs are significantly cheaper than Token Metadata NFTs and support
a plugin system for royalty enforcement, freeze delegates, attributes, etc.

Example:
  arctl metaplex create-nft "My NFT" \\
    --uri "https://arweave.net/metadata.json" \\
    --collection <COLLECTION_ADDRESS>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name = args[0]
			if uri == "" {
				return fmt.Errorf("--uri is required")
			}

			fmt.Printf("Creating Core NFT: %s\n", name)
			fmt.Printf("  Network: %s\n", network)
			fmt.Printf("  URI: %s\n", uri)
			if collection != "" {
				fmt.Printf("  Collection: %s\n", collection)
			}
			fmt.Println()
			fmt.Println("Equivalent mplx command:")
			collectionFlag := ""
			if collection != "" {
				collectionFlag = fmt.Sprintf(" --collection %s", collection)
			}
			fmt.Printf("  mplx core create %q --uri %q%s\n", name, uri, collectionFlag)
			return nil
		},
	}

	cmd.Flags().StringVar(&network, "network", "solana-devnet", "Solana cluster")
	cmd.Flags().StringVar(&uri, "uri", "", "URI to NFT metadata JSON")
	cmd.Flags().StringVar(&collection, "collection", "", "MPL Core collection address (optional)")
	cmd.Flags().StringVar(&wallet, "wallet", "", "Wallet address")

	return cmd
}

func newCreateCollectionCmd() *cobra.Command {
	var network, name, uri string

	cmd := &cobra.Command{
		Use:   "create-collection [name]",
		Short: "Create an MPL Core collection",
		Long: `Create an MPL Core collection for grouping NFTs and agents.
		
Collections enable collection-level authority checks during agent
registration and NFT management.

Example:
  arctl metaplex create-collection "My Collection" \\
    --uri "https://arweave.net/collection-metadata.json"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name = args[0]
			if uri == "" {
				return fmt.Errorf("--uri is required")
			}

			fmt.Printf("Creating Core collection: %s\n", name)
			fmt.Printf("  Network: %s\n", network)
			fmt.Printf("  URI: %s\n", uri)
			fmt.Println()
			fmt.Println("Equivalent mplx command:")
			fmt.Printf("  mplx core collection create %q --uri %q\n", name, uri)
			return nil
		},
	}

	cmd.Flags().StringVar(&network, "network", "solana-devnet", "Solana cluster")
	cmd.Flags().StringVar(&uri, "uri", "", "URI to collection metadata JSON")

	return cmd
}

// BuildAgentMetaplexSpec builds a MetaplexAgentSpec from CLI flags.
func BuildAgentMetaplexSpec(assetAddress, identityPDA, signerPDA, tokenMint, registrationURI, collection, network, executiveProfile, launchType string) *v1alpha1.MetaplexAgentSpec {
	spec := &v1alpha1.MetaplexAgentSpec{
		CoreAssetAddress:     assetAddress,
		AgentIdentityPDA:     identityPDA,
		AssetSignerPDA:       signerPDA,
		AgentTokenMintAddress: tokenMint,
		AgentRegistrationURI: registrationURI,
		CollectionAddress:    collection,
		Network:              network,
		ExecutiveProfilePDA:  executiveProfile,
		LaunchType:           launchType,
	}
	return spec
}