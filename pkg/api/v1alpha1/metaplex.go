package v1alpha1

// MetaplexAgentSpec holds Metaplex on-chain identity and token launch fields.
// Embedded in AgentSpec when the agent is registered on Solana via Metaplex.
type MetaplexAgentSpec struct {
	// CoreAssetAddress is the MPL Core asset address on Solana that
	// represents this agent's on-chain identity. Set during mintAndSubmitAgent
	// or registerIdentityV1.
	CoreAssetAddress string `json:"coreAssetAddress,omitempty" yaml:"coreAssetAddress,omitempty"`

	// AgentIdentityPDA is the derived PDA address for the agent's identity
	// record (AgentIdentityV2). Derived from seeds ["agent_identity", asset].
	AgentIdentityPDA string `json:"agentIdentityPDA,omitempty" yaml:"agentIdentityPDA,omitempty"`

	// AssetSignerPDA is the agent's built-in wallet PDA derived from
	// seeds ["mpl-core-execute", asset]. No private key exists.
	AssetSignerPDA string `json:"assetSignerPDA,omitempty" yaml:"assetSignerPDA,omitempty"`

	// AgentTokenMintAddress is the mint address of the agent's canonical
	// token launched via Genesis and bound via setAgentTokenV1.
	AgentTokenMintAddress string `json:"agentTokenMintAddress,omitempty" yaml:"agentTokenMintAddress,omitempty"`

	// AgentRegistrationURI is the URI pointing to the off-chain EIP-8004
	// agent registration JSON document.
	AgentRegistrationURI string `json:"agentRegistrationURI,omitempty" yaml:"agentRegistrationURI,omitempty"`

	// CollectionAddress is the MPL Core collection this agent asset
	// belongs to (optional).
	CollectionAddress string `json:"collectionAddress,omitempty" yaml:"collectionAddress,omitempty"`

	// Network is the Solana cluster: "solana-mainnet", "solana-devnet",
	// "localnet", "eclipse-mainnet", "sonic-mainnet", "sonic-devnet",
	// "fogo-mainnet", "fogo-testnet".
	Network string `json:"network,omitempty" yaml:"network,omitempty"`

	// ExecutiveProfilePDA is the PDA address for the executive profile
	// that is authorized to sign transactions on behalf of this agent.
	ExecutiveProfilePDA string `json:"executiveProfilePDA,omitempty" yaml:"executiveProfilePDA,omitempty"`

	// LaunchType indicates the Genesis launch type: "bondingCurve" or "launchpool".
	LaunchType string `json:"launchType,omitempty" yaml:"launchType,omitempty"`

	// TokenMetadata contains the on-chain token metadata fields.
	TokenMetadata *MetaplexTokenMetadata `json:"tokenMetadata,omitempty" yaml:"tokenMetadata,omitempty"`

	// RegistrationDocument contains the off-chain EIP-8004 agent metadata.
	RegistrationDocument *AgentRegistrationDocument `json:"registrationDocument,omitempty" yaml:"registrationDocument,omitempty"`
}

// MetaplexTokenMetadata describes the on-chain token metadata for a Metaplex
// Genesis token launch.
type MetaplexTokenMetadata struct {
	Name   string `json:"name,omitempty" yaml:"name,omitempty"`
	Symbol string `json:"symbol,omitempty" yaml:"symbol,omitempty"`
	Image  string `json:"image,omitempty" yaml:"image,omitempty"` // Must be Irys URL
	// URI is the off-chain token metadata URI (e.g. Arweave).
	URI string `json:"uri,omitempty" yaml:"uri,omitempty"`
}

// AgentRegistrationDocument follows the ERC-8004 standard adapted for Solana.
type AgentRegistrationDocument struct {
	Type            string                    `json:"type" yaml:"type"`
	Name            string                    `json:"name" yaml:"name"`
	Description     string                    `json:"description" yaml:"description"`
	Image           string                    `json:"image" yaml:"image"`
	Services        []AgentServiceEntry       `json:"services,omitempty" yaml:"services,omitempty"`
	Active          bool                      `json:"active,omitempty" yaml:"active,omitempty"`
	Registrations   []AgentRegistrationEntry  `json:"registrations,omitempty" yaml:"registrations,omitempty"`
	SupportedTrust  []string                  `json:"supportedTrust,omitempty" yaml:"supportedTrust,omitempty"`
}

// AgentServiceEntry describes one service endpoint the agent exposes.
type AgentServiceEntry struct {
	Name     string   `json:"name" yaml:"name"`
	Endpoint string   `json:"endpoint" yaml:"endpoint"`
	Version  string   `json:"version,omitempty" yaml:"version,omitempty"`
	Skills   []string `json:"skills,omitempty" yaml:"skills,omitempty"`
	Domains  []string `json:"domains,omitempty" yaml:"domains,omitempty"`
}

// AgentRegistrationEntry links back to an on-chain identity record.
type AgentRegistrationEntry struct {
	AgentID      string `json:"agentId" yaml:"agentId"`
	AgentRegistry string `json:"agentRegistry" yaml:"agentRegistry"` // e.g. "solana:101:metaplex"
}

// MetaplexConstants holds well-known Metaplex program IDs and seed prefixes.
var MetaplexConstants = struct {
	AgentRegistryProgramID string
	CoreProgramID          string
	GenesisProgramID       string
	BubblegumProgramID     string
	TokenMetadataProgramID string
	CandyMachineProgramID  string
	AssetSignerSeed        string
	AgentIdentitySeed      string
	ExecutiveProfileSeed   string
	ExecutionDelegateSeed  string
}{
	AgentRegistryProgramID: "metaqbxxUerqCWQLDgFgV9GxBiLdJQbU8NtV",
	CoreProgramID:          "CoRE9H7pF4iZTGZq1XZC7Tqxq9AvtQz6YZ",
	GenesisProgramID:       "GEN3xNxH3YnqXqMxQk5bY5yY5y5Y5y5Y5y5Y5",
	BubblegumProgramID:     "BUBLgUMVxJc7xYXjLxLxLxLxLxLxLxLxLx",
	TokenMetadataProgramID: "metaqbxxUerqCWQLDgFgV9GxBiLdJQbU8NtV",
	CandyMachineProgramID:  "CM4PYT1eR5YXjLxLxLxLxLxLxLxLxLxLx",
	AssetSignerSeed:        "mpl-core-execute",
	AgentIdentitySeed:      "agent_identity",
	ExecutiveProfileSeed:   "executive_profile",
	ExecutionDelegateSeed:  "execution_delegate_record",
}