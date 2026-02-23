// Package sdk provides the interface for Creddy plugins.
//
// Creddy uses Hashicorp's go-plugin with gRPC for plugin communication.
// This allows plugins to be separate binaries that can be versioned and
// distributed independently from Creddy core.
package sdk

import (
	"context"
	"time"
)

// PluginInfo contains metadata about a plugin
type PluginInfo struct {
	// Name is the plugin identifier (e.g., "github", "doppler")
	Name string

	// Version is the semantic version (e.g., "1.2.3")
	Version string

	// Description is a human-readable description
	Description string

	// MinCreddyVersion is the minimum Creddy version required
	MinCreddyVersion string
}

// ScopeSpec describes a scope pattern this plugin handles
type ScopeSpec struct {
	// Pattern is the scope pattern (e.g., "github:*", "doppler:*")
	Pattern string

	// Description explains what this scope grants
	Description string

	// Examples shows example scope values
	Examples []string
}

// Agent represents the agent requesting credentials
type Agent struct {
	// ID is the unique agent identifier
	ID string

	// Name is the human-readable agent name
	Name string

	// Scopes are the scopes this agent is authorized for
	Scopes []string
}

// Credential represents an issued credential
type Credential struct {
	// Value is the credential value (token, key, etc.)
	Value string

	// ExpiresAt is when the credential expires
	ExpiresAt time.Time

	// ExternalID is an optional ID for revocation (empty if not revocable)
	ExternalID string

	// Metadata is optional additional data about the credential
	Metadata map[string]string
}

// CredentialRequest contains parameters for requesting a credential
type CredentialRequest struct {
	// Agent is the agent requesting the credential
	Agent Agent

	// Scope is the requested scope (e.g., "github:owner/repo:read")
	Scope string

	// TTL is the requested time-to-live
	TTL time.Duration

	// Parameters are scope-specific parameters
	Parameters map[string]string
}

// ConfigField describes a configuration field for dynamic CLI flags
type ConfigField struct {
	// Name is the field name in the JSON config (e.g., "app_id")
	Name string

	// Type is the field type:
	//   "string" - plain text
	//   "int"    - integer value
	//   "bool"   - true/false
	//   "file"   - path to file, CLI reads content
	//   "secret" - sensitive value, masked in interactive mode
	Type string

	// Description is a human-readable description
	Description string

	// Required indicates if this field must be provided
	Required bool

	// Default is the default value (as a string, empty if none)
	Default string
}

// Plugin is the interface that all Creddy plugins must implement
type Plugin interface {
	// Info returns plugin metadata
	Info(ctx context.Context) (*PluginInfo, error)

	// Scopes returns the scope patterns this plugin handles
	Scopes(ctx context.Context) ([]ScopeSpec, error)

	// ConfigSchema returns the configuration schema for dynamic CLI flags
	// This allows the CLI to generate proper flags instead of requiring raw JSON
	ConfigSchema(ctx context.Context) ([]ConfigField, error)

	// Constraints returns the TTL constraints for this plugin
	// Return nil if there are no constraints (any TTL is acceptable)
	Constraints(ctx context.Context) (*Constraints, error)

	// Configure sets up the plugin with the given configuration
	// Config is a JSON string with plugin-specific settings
	Configure(ctx context.Context, config string) error

	// Validate tests the plugin configuration (e.g., API connectivity)
	Validate(ctx context.Context) error

	// GetCredential issues a new credential
	GetCredential(ctx context.Context, req *CredentialRequest) (*Credential, error)

	// RevokeCredential revokes a previously issued credential
	// Returns nil if revocation is not supported or credential not found
	RevokeCredential(ctx context.Context, externalID string) error

	// MatchScope checks if the given scope matches this plugin
	MatchScope(ctx context.Context, scope string) (bool, error)
}

// PluginConfig is the standard configuration structure
type PluginConfig struct {
	// Raw JSON config from Creddy
	Raw string

	// Parsed config values (plugin-specific)
	Values map[string]interface{}
}

// Constraints describes the TTL limits for a plugin
type Constraints struct {
	// MaxTTL is the maximum TTL this plugin supports (0 = no maximum)
	MaxTTL time.Duration

	// MinTTL is the minimum TTL this plugin supports (0 = no minimum)
	MinTTL time.Duration

	// Description explains the constraint (e.g., "GitHub installation tokens have a maximum lifetime of 1 hour")
	Description string
}
