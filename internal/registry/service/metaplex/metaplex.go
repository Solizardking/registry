// Package metaplex provides Solana on-chain operations through the Metaplex
// protocol stack: Agent Registry, MPL Core, Genesis, Bubblegum, Token Metadata,
// and Candy Machine.
//
// This package mirrors the Metaplex agent ecosystem at
// https://metaplex.com/agents — every Go function corresponds to an equivalent
// mplx CLI command or Umi SDK operation.
package metaplex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/agentregistry-dev/agentregistry/internal/registry/config"
	"github.com/agentregistry-dev/agentregistry/pkg/api/v1alpha1"
)

// Service provides Metaplex on-chain operations. It talks to the Metaplex
// API (https://api.metaplex.com) and Solana RPC endpoints to construct and
// submit transactions.
type Service struct {
	cfg     *config.Config
	client  *http.Client
}

// New creates a new Metaplex service.
func New(cfg *config.Config) *Service {
	return &Service{
		cfg: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// MintAgentRequest describes the input to mintAndSubmitAgent (Metaplex API).
// Mirrors the @metaplex-foundation/mpl-agent-registry mintAgent input.
type MintAgentRequest struct {
	Wallet        string                 `json:"wallet"`
	Network       string                 `json:"network,omitempty"`
	Name          string                 `json:"name"`
	URI           string                 `json:"uri"`
	AgentMetadata map[string]interface{} `json:"agentMetadata"`
}

// MintAgentResult is the response from mintAgent.
type MintAgentResult struct {
	AssetAddress string `json:"assetAddress"`
	Signature    string `json:"signature"`
	Transaction  string `json:"transaction"` // base64-encoded unsigned transaction
}

// MintAgent mints a new Core asset and registers agent identity in one
// transaction via the Metaplex API.
func (s *Service) MintAgent(ctx context.Context, req *MintAgentRequest) (*MintAgentResult, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	rpcURL := s.resolveRPC(req.Network)
	apiURL := s.cfg.MetaplexAPIBaseURL

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		apiURL+"/v1/agents/mint?rpc="+rpcURL,
		bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("calling Metaplex API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Metaplex API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result MintAgentResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

// RegisterIdentityRequest describes the input to registerIdentityV1.
type RegisterIdentityRequest struct {
	Asset                 string `json:"asset"`
	Collection            string `json:"collection,omitempty"`
	AgentRegistrationURI  string `json:"agentRegistrationUri"`
	Network               string `json:"network,omitempty"`
}

// RegisterIdentityResult is the response from registerIdentityV1.
type RegisterIdentityResult struct {
	Signature string `json:"signature"`
}

// RegisterIdentity binds an identity record to an existing MPL Core asset.
// Equivalent to registerIdentityV1 in mpl-agent-registry.
func (s *Service) RegisterIdentity(ctx context.Context, req *RegisterIdentityRequest) (*RegisterIdentityResult, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	rpcURL := s.resolveRPC(req.Network)
	apiURL := s.cfg.MetaplexAPIBaseURL

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		apiURL+"/v1/agents/register-identity?rpc="+rpcURL,
		bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("calling Metaplex API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Metaplex API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result RegisterIdentityResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

// CreateTokenLaunchRequest describes the input to createAndRegisterLaunch
// on Genesis protocol.
type CreateTokenLaunchRequest struct {
	Wallet      string                 `json:"wallet"`
	Network     string                 `json:"network,omitempty"`
	LaunchType  string                 `json:"launchType"` // "bondingCurve" or "launchpool"
	Token       map[string]interface{} `json:"token"`
	Launch      map[string]interface{} `json:"launch"`
	Agent       map[string]interface{} `json:"agent,omitempty"`
}

// CreateTokenLaunchResult is the response from createAndRegisterLaunch.
type CreateTokenLaunchResult struct {
	MintAddress string `json:"mintAddress"`
	Signature   string `json:"signature"`
	Link        string `json:"link"`
}

// CreateTokenLaunch launches a token via Genesis protocol.
// When agent params are provided, creator fees route to the agent's PDA
// and setAgentTokenV1 binds the token permanently to the agent.
func (s *Service) CreateTokenLaunch(ctx context.Context, req *CreateTokenLaunchRequest) (*CreateTokenLaunchResult, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	rpcURL := s.resolveRPC(req.Network)
	apiURL := s.cfg.MetaplexAPIBaseURL

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		apiURL+"/v1/genesis/launch?rpc="+rpcURL,
		bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("calling Metaplex API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Metaplex API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result CreateTokenLaunchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

// ReadAgentDataRequest describes fetching agent identity data from on-chain.
type ReadAgentDataRequest struct {
	AssetAddress string `json:"assetAddress"`
	Network      string `json:"network,omitempty"`
}

// ReadAgentDataResult holds the agent identity data fetched on-chain.
type ReadAgentDataResult struct {
	AssetAddress     string `json:"assetAddress"`
	AgentIdentityPDA string `json:"agentIdentityPDA"`
	AssetSignerPDA   string `json:"assetSignerPDA"`
	Registered       bool   `json:"registered"`
	RegistrationURI  string `json:"registrationURI,omitempty"`
	Balance          uint64 `json:"balance,omitempty"` // lamports in Asset Signer
}

// ReadAgentData fetches agent identity data on-chain and returns
// the identity PDA, Asset Signer PDA, and balance.
func (s *Service) ReadAgentData(ctx context.Context, req *ReadAgentDataRequest) (*ReadAgentDataResult, error) {
	rpcURL := s.resolveRPC(req.Network)
	apiURL := s.cfg.MetaplexAPIBaseURL

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/v1/agents/%s?rpc=%s", apiURL, req.AssetAddress, rpcURL),
		nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("calling Metaplex API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Metaplex API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result ReadAgentDataResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

// SetAgentTokenRequest describes linking a Genesis token to an agent.
type SetAgentTokenRequest struct {
	AssetAddress    string `json:"assetAddress"`
	GenesisAccount  string `json:"genesisAccount"`
	Network         string `json:"network,omitempty"`
}

// SetAgentTokenResult is the response from setAgentTokenV1.
type SetAgentTokenResult struct {
	Signature string `json:"signature"`
}

// SetAgentToken permanently links a Genesis token to the agent via
// setAgentTokenV1. This is irreversible — one token per agent forever.
func (s *Service) SetAgentToken(ctx context.Context, req *SetAgentTokenRequest) (*SetAgentTokenResult, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	rpcURL := s.resolveRPC(req.Network)
	apiURL := s.cfg.MetaplexAPIBaseURL

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		apiURL+"/v1/agents/set-token?rpc="+rpcURL,
		bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("calling Metaplex API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Metaplex API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result SetAgentTokenResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

// RegisterExecutiveProfileRequest describes registering an executive profile.
type RegisterExecutiveProfileRequest struct {
	Authority string `json:"authority"`
	Network   string `json:"network,omitempty"`
}

// RegisterExecutiveProfileResult is the response from registerExecutiveV1.
type RegisterExecutiveProfileResult struct {
	Signature string `json:"signature"`
}

// RegisterExecutiveProfile registers a wallet as an executive operator.
// Equivalent to registerExecutiveV1 in mpl-agent-tools.
func (s *Service) RegisterExecutiveProfile(ctx context.Context, req *RegisterExecutiveProfileRequest) (*RegisterExecutiveProfileResult, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	rpcURL := s.resolveRPC(req.Network)
	apiURL := s.cfg.MetaplexAPIBaseURL

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		apiURL+"/v1/agents/executive/register?rpc="+rpcURL,
		bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("calling Metaplex API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Metaplex API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result RegisterExecutiveProfileResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

// DelegateExecutionRequest describes delegating execution to an executive.
type DelegateExecutionRequest struct {
	AgentAsset    string `json:"agentAsset"`
	Executive     string `json:"executive"`
	Network       string `json:"network,omitempty"`
}

// DelegateExecutionResult is the response from delegateExecutionV1.
type DelegateExecutionResult struct {
	Signature string `json:"signature"`
}

// DelegateExecution creates a delegation record linking an agent to an executive.
func (s *Service) DelegateExecution(ctx context.Context, req *DelegateExecutionRequest) (*DelegateExecutionResult, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	rpcURL := s.resolveRPC(req.Network)
	apiURL := s.cfg.MetaplexAPIBaseURL

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		apiURL+"/v1/agents/executive/delegate?rpc="+rpcURL,
		bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("calling Metaplex API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Metaplex API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result DelegateExecutionResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

// RevokeExecutionRequest describes revoking execution delegation.
type RevokeExecutionRequest struct {
	AgentAsset string `json:"agentAsset"`
	Executive  string `json:"executive"`
	Network    string `json:"network,omitempty"`
}

// RevokeExecutionResult is the response from revokeExecutionV1.
type RevokeExecutionResult struct {
	Signature string `json:"signature"`
}

// RevokeExecution revokes an execution delegation record.
func (s *Service) RevokeExecution(ctx context.Context, req *RevokeExecutionRequest) (*RevokeExecutionResult, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	rpcURL := s.resolveRPC(req.Network)
	apiURL := s.cfg.MetaplexAPIBaseURL

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		apiURL+"/v1/agents/executive/revoke?rpc="+rpcURL,
		bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("calling Metaplex API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Metaplex API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result RevokeExecutionResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

// CreateNFTRequest describes creating an MPL Core NFT.
type CreateNFTRequest struct {
	Wallet        string `json:"wallet"`
	Name          string `json:"name"`
	URI           string `json:"uri"`
	Collection    string `json:"collection,omitempty"`
	Network       string `json:"network,omitempty"`
}

// CreateNFTResult is the response from creating a Core NFT.
type CreateNFTResult struct {
	AssetAddress string `json:"assetAddress"`
	Signature    string `json:"signature"`
}

// CreateCoreNFT creates an MPL Core NFT via the Metaplex API.
func (s *Service) CreateCoreNFT(ctx context.Context, req *CreateNFTRequest) (*CreateNFTResult, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	rpcURL := s.resolveRPC(req.Network)
	apiURL := s.cfg.MetaplexAPIBaseURL

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		apiURL+"/v1/core/nft?rpc="+rpcURL,
		bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("calling Metaplex API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Metaplex API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result CreateNFTResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}

// resolveRPC returns the appropriate RPC URL based on network.
func (s *Service) resolveRPC(network string) string {
	switch network {
	case "solana-devnet", "":
		return s.cfg.SolanaDevnetRPCURL
	case "solana-mainnet":
		return s.cfg.SolanaRPCURL
	default:
		return s.cfg.SolanaRPCURL
	}
}

// BuildAgentRegistrationDocument builds a standard EIP-8004 registration
// document for the given agent and returns the JSON bytes for upload.
func BuildAgentRegistrationDocument(name, description, image string, services []v1alpha1.AgentServiceEntry, registrations []v1alpha1.AgentRegistrationEntry) ([]byte, error) {
	doc := v1alpha1.AgentRegistrationDocument{
		Type:        "https://eips.ethereum.org/EIPS/eip-8004#registration-v1",
		Name:        name,
		Description: description,
		Image:       image,
		Services:    services,
		Active:      true,
		Registrations: registrations,
		SupportedTrust: []string{"reputation", "crypto-economic"},
	}
	return json.MarshalIndent(doc, "", "  ")
}