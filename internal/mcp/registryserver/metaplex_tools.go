package registryserver

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/agentregistry-dev/agentregistry/internal/registry/config"
	"github.com/agentregistry-dev/agentregistry/internal/registry/service/metaplex"
)

// addMetaplexTools registers Metaplex on-chain operation tools with the
// MCP server. These tools mirror the mplx CLI and Umi SDK for agent
// minting, identity registration, token launching, and execution delegation.
func addMetaplexTools(server *mcp.Server, cfg *config.Config) {
	svc := metaplex.New(cfg)

	// mint_agent — creates a Core asset + registers identity in one tx
	mcp.AddTool(server, &mcp.Tool{
		Name:        "mint_agent",
		Description: "Mint an MPL Core asset and register agent identity via the Metaplex API in one transaction",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args struct {
		Wallet      string `json:"wallet"      doc:"Wallet address that pays" required:"true"`
		Name        string `json:"name"        doc:"Agent display name" required:"true"`
		URI         string `json:"uri"         doc:"URI to Core asset NFT metadata JSON" required:"true"`
		Network     string `json:"network,omitempty"    doc:"Solana cluster (solana-mainnet, solana-devnet)"`
		Description string `json:"description,omitempty" doc:"Agent description for the registration document"`
	}) (*mcp.CallToolResult, *metaplex.MintAgentResult, error) {
		agentMetadata := map[string]interface{}{
			"type":        "agent",
			"name":        args.Name,
			"description": args.Description,
			"services":    []interface{}{},
			"registrations": []interface{}{},
			"supportedTrust": []string{"reputation"},
		}
		result, err := svc.MintAgent(ctx, &metaplex.MintAgentRequest{
			Wallet:        args.Wallet,
			Network:       args.Network,
			Name:          args.Name,
			URI:           args.URI,
			AgentMetadata: agentMetadata,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("mint agent: %w", err)
		}
		return nil, result, nil
	})

	// register_identity — attach AgentIdentity to existing Core asset
	mcp.AddTool(server, &mcp.Tool{
		Name:        "register_agent_identity",
		Description: "Register an agent identity on an existing MPL Core asset (registerIdentityV1)",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args struct {
		Asset              string `json:"asset"                doc:"MPL Core asset address" required:"true"`
		RegistrationURI    string `json:"registrationUri"      doc:"URI to EIP-8004 agent registration JSON" required:"true"`
		Collection         string `json:"collection,omitempty" doc:"MPL Core collection address (optional)"`
		Network            string `json:"network,omitempty"    doc:"Solana cluster"`
	}) (*mcp.CallToolResult, *metaplex.RegisterIdentityResult, error) {
		result, err := svc.RegisterIdentity(ctx, &metaplex.RegisterIdentityRequest{
			Asset:                args.Asset,
			Collection:           args.Collection,
			AgentRegistrationURI: args.RegistrationURI,
			Network:              args.Network,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("register identity: %w", err)
		}
		return nil, result, nil
	})

	// launch_token — Genesis token launch
	mcp.AddTool(server, &mcp.Tool{
		Name:        "launch_agent_token",
		Description: "Launch a token via Genesis protocol (bonding curve or launchpool)",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args struct {
		Wallet        string  `json:"wallet"                 doc:"Wallet address" required:"true"`
		Name          string  `json:"name"                   doc:"Token name (1-32 chars)" required:"true"`
		Symbol        string  `json:"symbol"                 doc:"Token symbol (1-10 chars)" required:"true"`
		Image         string  `json:"image"                  doc:"Token image URL (must be Irys gateway)" required:"true"`
		Network       string  `json:"network,omitempty"      doc:"Solana cluster"`
		LaunchType    string  `json:"launchType,omitempty"   doc:"bonding-curve or launchpool (default bonding-curve)"`
		AgentAsset    string  `json:"agentAsset,omitempty"   doc:"Registered agent Core asset address for fee routing"`
		SetToken      bool    `json:"setToken,omitempty"     doc:"Permanently bind token to agent (irreversible)"`
		FirstBuyAmount float64 `json:"firstBuyAmount,omitempty" doc:"Fee-free first buy SOL amount"`
		Description   string  `json:"description,omitempty"  doc:"Token description (max 250 chars)"`
	}) (*mcp.CallToolResult, *metaplex.CreateTokenLaunchResult, error) {
		launchType := args.LaunchType
		if launchType == "" {
			launchType = "bonding-curve"
		}
		token := map[string]interface{}{
			"name":   args.Name,
			"symbol": args.Symbol,
			"image":  args.Image,
		}
		if args.Description != "" {
			token["description"] = args.Description
		}
		launch := map[string]interface{}{}
		if args.FirstBuyAmount > 0 {
			launch["firstBuyAmount"] = args.FirstBuyAmount
		}
		req := &metaplex.CreateTokenLaunchRequest{
			Wallet:     args.Wallet,
			Network:    args.Network,
			LaunchType: launchType,
			Token:      token,
			Launch:     launch,
		}
		if args.AgentAsset != "" {
			req.Agent = map[string]interface{}{
				"mint":     args.AgentAsset,
				"setToken": args.SetToken,
			}
		}
		result, err := svc.CreateTokenLaunch(ctx, req)
		if err != nil {
			return nil, nil, fmt.Errorf("launch token: %w", err)
		}
		return nil, result, nil
	})

	// set_agent_token — permanent token-agent binding
	mcp.AddTool(server, &mcp.Tool{
		Name:        "set_agent_token",
		Description: "Permanently link a Genesis token to an agent via setAgentTokenV1 (irreversible)",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args struct {
		AssetAddress   string `json:"assetAddress"   doc:"Registered agent's Core asset address" required:"true"`
		GenesisAccount string `json:"genesisAccount" doc:"Genesis account/mint address" required:"true"`
		Network        string `json:"network,omitempty" doc:"Solana cluster"`
	}) (*mcp.CallToolResult, *metaplex.SetAgentTokenResult, error) {
		result, err := svc.SetAgentToken(ctx, &metaplex.SetAgentTokenRequest{
			AssetAddress:   args.AssetAddress,
			GenesisAccount: args.GenesisAccount,
			Network:        args.Network,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("set agent token: %w", err)
		}
		return nil, result, nil
	})

	// read_agent_data — fetch on-chain agent identity
	mcp.AddTool(server, &mcp.Tool{
		Name:        "read_agent_data",
		Description: "Read agent identity data from on-chain (PDA, wallet, balance)",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args struct {
		AssetAddress string `json:"assetAddress" doc:"MPL Core asset address" required:"true"`
		Network      string `json:"network,omitempty" doc:"Solana cluster"`
	}) (*mcp.CallToolResult, *metaplex.ReadAgentDataResult, error) {
		result, err := svc.ReadAgentData(ctx, &metaplex.ReadAgentDataRequest{
			AssetAddress: args.AssetAddress,
			Network:      args.Network,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("read agent data: %w", err)
		}
		return nil, result, nil
	})

	// register_executive — create executive profile
	mcp.AddTool(server, &mcp.Tool{
		Name:        "register_executive",
		Description: "Register a wallet as an executive operator profile (one-time setup per wallet)",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args struct {
		Authority string `json:"authority,omitempty" doc:"Wallet address (defaults to signer)"`
		Network   string `json:"network,omitempty"   doc:"Solana cluster"`
	}) (*mcp.CallToolResult, *metaplex.RegisterExecutiveProfileResult, error) {
		result, err := svc.RegisterExecutiveProfile(ctx, &metaplex.RegisterExecutiveProfileRequest{
			Authority: args.Authority,
			Network:   args.Network,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("register executive: %w", err)
		}
		return nil, result, nil
	})

	// delegate_execution — link agent to executive
	mcp.AddTool(server, &mcp.Tool{
		Name:        "delegate_execution",
		Description: "Delegate execution authority from an agent to an executive (creates delegation record)",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args struct {
		AgentAsset string `json:"agentAsset" doc:"Agent's MPL Core asset address" required:"true"`
		Executive  string `json:"executive"  doc:"Executive wallet address" required:"true"`
		Network    string `json:"network,omitempty" doc:"Solana cluster"`
	}) (*mcp.CallToolResult, *metaplex.DelegateExecutionResult, error) {
		result, err := svc.DelegateExecution(ctx, &metaplex.DelegateExecutionRequest{
			AgentAsset: args.AgentAsset,
			Executive:  args.Executive,
			Network:    args.Network,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("delegate execution: %w", err)
		}
		return nil, result, nil
	})

	// revoke_execution — revoke delegation
	mcp.AddTool(server, &mcp.Tool{
		Name:        "revoke_execution",
		Description: "Revoke an execution delegation record (stops future Execute calls)",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args struct {
		AgentAsset string `json:"agentAsset" doc:"Agent's MPL Core asset address" required:"true"`
		Executive  string `json:"executive"  doc:"Executive wallet address" required:"true"`
		Network    string `json:"network,omitempty" doc:"Solana cluster"`
	}) (*mcp.CallToolResult, *metaplex.RevokeExecutionResult, error) {
		result, err := svc.RevokeExecution(ctx, &metaplex.RevokeExecutionRequest{
			AgentAsset: args.AgentAsset,
			Executive:  args.Executive,
			Network:    args.Network,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("revoke execution: %w", err)
		}
		return nil, result, nil
	})

	// create_core_nft — create an MPL Core NFT
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_core_nft",
		Description: "Create an MPL Core NFT (next-gen NFT standard, cheaper than TM)",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args struct {
		Wallet     string `json:"wallet"                doc:"Wallet address" required:"true"`
		Name       string `json:"name"                  doc:"NFT name" required:"true"`
		URI        string `json:"uri"                   doc:"URI to NFT metadata JSON" required:"true"`
		Collection string `json:"collection,omitempty"  doc:"MPL Core collection address"`
		Network    string `json:"network,omitempty"     doc:"Solana cluster"`
	}) (*mcp.CallToolResult, *metaplex.CreateNFTResult, error) {
		result, err := svc.CreateCoreNFT(ctx, &metaplex.CreateNFTRequest{
			Wallet:     args.Wallet,
			Name:       args.Name,
			URI:        args.URI,
			Collection: args.Collection,
			Network:    args.Network,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("create core nft: %w", err)
		}
		return nil, result, nil
	})
}