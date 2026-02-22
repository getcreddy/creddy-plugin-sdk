# Creddy Plugin SDK

SDK for building Creddy credential plugins.

## Overview

Creddy uses a plugin architecture for credential backends. Each integration (GitHub, Anthropic, Doppler, etc.) is a separate binary that communicates with Creddy core via gRPC.

This SDK provides:
- The `Plugin` interface that all plugins must implement
- gRPC server/client implementation (via Hashicorp go-plugin)
- Standalone CLI mode for testing plugins without Creddy
- Test utilities

## Quick Start

```go
package main

import (
    "context"
    
    sdk "github.com/getcreddy/creddy-plugin-sdk"
)

type MyPlugin struct {
    config MyConfig
}

func (p *MyPlugin) Info(ctx context.Context) (*sdk.PluginInfo, error) {
    return &sdk.PluginInfo{
        Name:             "my-integration",
        Version:          "0.1.0",
        Description:      "My custom Creddy integration",
        MinCreddyVersion: "0.4.0",
    }, nil
}

func (p *MyPlugin) Scopes(ctx context.Context) ([]sdk.ScopeSpec, error) {
    return []sdk.ScopeSpec{
        {
            Pattern:     "my-integration:*",
            Description: "Access to my integration",
            Examples:    []string{"my-integration:resource:read"},
        },
    }, nil
}

// ... implement other methods ...

func main() {
    sdk.ServeWithStandalone(&MyPlugin{}, nil)
}
```

## Plugin Interface

```go
type Plugin interface {
    // Metadata
    Info(ctx context.Context) (*PluginInfo, error)
    Scopes(ctx context.Context) ([]ScopeSpec, error)
    
    // Configuration
    Configure(ctx context.Context, config string) error
    Validate(ctx context.Context) error
    
    // Credentials
    GetCredential(ctx context.Context, req *CredentialRequest) (*Credential, error)
    RevokeCredential(ctx context.Context, externalID string) error
    MatchScope(ctx context.Context, scope string) (bool, error)
}
```

## Development Mode

Plugins can run standalone for testing without Creddy:

```bash
# Show plugin info
./my-plugin info

# List supported scopes
./my-plugin scopes

# Validate configuration
./my-plugin validate --config config.json

# Get a credential
./my-plugin get --config config.json --scope "my-integration:resource" --ttl 10m

# Revoke a credential
./my-plugin revoke --config config.json --external-id "abc123"
```

## Building

```bash
# Build plugin
go build -o creddy-my-integration .

# Run tests
go test ./...
```

## Official Plugins

- [creddy-github](https://github.com/getcreddy/creddy-github) - GitHub App installation tokens
- [creddy-anthropic](https://github.com/getcreddy/creddy-anthropic) - Anthropic API keys
- [creddy-doppler](https://github.com/getcreddy/creddy-doppler) - Doppler service tokens

## License

Apache 2.0
