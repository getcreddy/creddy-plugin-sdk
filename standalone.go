package sdk

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
)

// runStandalone provides a CLI interface for testing plugins without Creddy
func runStandalone(p Plugin, cfg *StandaloneConfig) {
	if cfg == nil {
		cfg = &StandaloneConfig{}
	}
	if cfg.ConfigFlag == "" {
		cfg.ConfigFlag = "config"
	}

	if len(os.Args) < 2 {
		printStandaloneUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	ctx := context.Background()

	switch cmd {
	case "info":
		runInfo(ctx, p)
	case "scopes":
		runScopes(ctx, p)
	case "schema":
		runSchema(ctx, p)
	case "constraints":
		runConstraints(ctx, p)
	case "validate":
		runValidate(ctx, p, args, cfg)
	case "get":
		runGet(ctx, p, args, cfg)
	case "revoke":
		runRevoke(ctx, p, args, cfg)
	case "help", "-h", "--help":
		printStandaloneUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printStandaloneUsage()
		os.Exit(1)
	}
}

func printStandaloneUsage() {
	fmt.Println(`Creddy Plugin Standalone Mode

Usage:
  plugin <command> [flags]

Commands:
  info        Show plugin information
  scopes      List supported scopes
  schema      Show configuration schema (JSON output)
  constraints Show TTL constraints (max/min TTL)
  validate    Validate configuration
  get         Get a credential
  revoke      Revoke a credential
  help        Show this help

Flags:
  --config   Path to JSON config file

Examples:
  # Show plugin info
  ./creddy-github info

  # Show TTL constraints
  ./creddy-github constraints

  # Show config schema (for CLI integration)
  ./creddy-github schema

  # Validate configuration
  ./creddy-github validate --config config.json

  # Get a credential
  ./creddy-github get --config config.json --scope "github:owner/repo" --ttl 10m

  # Revoke a credential
  ./creddy-github revoke --config config.json --external-id "abc123"
`)
}

func runInfo(ctx context.Context, p Plugin) {
	info, err := p.Info(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Name:              %s\n", info.Name)
	fmt.Printf("Version:           %s\n", info.Version)
	fmt.Printf("Description:       %s\n", info.Description)
	fmt.Printf("Min Creddy Version: %s\n", info.MinCreddyVersion)
}

func runScopes(ctx context.Context, p Plugin) {
	scopes, err := p.Scopes(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	for _, s := range scopes {
		fmt.Printf("Pattern: %s\n", s.Pattern)
		fmt.Printf("  Description: %s\n", s.Description)
		if len(s.Examples) > 0 {
			fmt.Printf("  Examples:\n")
			for _, ex := range s.Examples {
				fmt.Printf("    - %s\n", ex)
			}
		}
		fmt.Println()
	}
}

func runSchema(ctx context.Context, p Plugin) {
	fields, err := p.ConfigSchema(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	// Output as JSON for CLI consumption
	output, err := json.MarshalIndent(fields, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling schema: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(output))
}

func runConstraints(ctx context.Context, p Plugin) {
	constraints, err := p.Constraints(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if constraints == nil {
		fmt.Println("No TTL constraints (any TTL is acceptable)")
		return
	}
	if constraints.MaxTTL > 0 {
		fmt.Printf("Max TTL: %s\n", constraints.MaxTTL)
	} else {
		fmt.Println("Max TTL: none")
	}
	if constraints.MinTTL > 0 {
		fmt.Printf("Min TTL: %s\n", constraints.MinTTL)
	} else {
		fmt.Println("Min TTL: none")
	}
	if constraints.Description != "" {
		fmt.Printf("Description: %s\n", constraints.Description)
	}
}

func runValidate(ctx context.Context, p Plugin, args []string, cfg *StandaloneConfig) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	configFile := fs.String(cfg.ConfigFlag, "", "Path to JSON config file")
	fs.Parse(args)

	if *configFile == "" {
		fmt.Fprintln(os.Stderr, "Error: --config is required")
		os.Exit(1)
	}

	configJSON, err := os.ReadFile(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
		os.Exit(1)
	}

	if err := p.Configure(ctx, string(configJSON)); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	if err := p.Validate(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Validation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Configuration valid")
}

func runGet(ctx context.Context, p Plugin, args []string, cfg *StandaloneConfig) {
	fs := flag.NewFlagSet("get", flag.ExitOnError)
	configFile := fs.String(cfg.ConfigFlag, "", "Path to JSON config file")
	scope := fs.String("scope", "", "Scope to request")
	ttl := fs.Duration("ttl", 10*time.Minute, "TTL for credential")
	agentID := fs.String("agent-id", "test-agent", "Agent ID for testing")
	agentName := fs.String("agent-name", "Test Agent", "Agent name for testing")
	paramsJSON := fs.String("params", "{}", "JSON parameters")
	fs.Parse(args)

	if *configFile == "" {
		fmt.Fprintln(os.Stderr, "Error: --config is required")
		os.Exit(1)
	}
	if *scope == "" {
		fmt.Fprintln(os.Stderr, "Error: --scope is required")
		os.Exit(1)
	}

	configJSON, err := os.ReadFile(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
		os.Exit(1)
	}

	if err := p.Configure(ctx, string(configJSON)); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	var params map[string]string
	json.Unmarshal([]byte(*paramsJSON), &params)

	cred, err := p.GetCredential(ctx, &CredentialRequest{
		Agent: Agent{
			ID:     *agentID,
			Name:   *agentName,
			Scopes: []string{*scope},
		},
		Scope:      *scope,
		TTL:        *ttl,
		Parameters: params,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting credential: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Credential Value: %s\n", cred.Value)
	fmt.Printf("Expires At:       %s\n", cred.ExpiresAt.Format(time.RFC3339))
	if cred.Credential != "" {
		fmt.Printf("Revocation Cred:  %s\n", cred.Credential)
	}
	if len(cred.Metadata) > 0 {
		fmt.Printf("Metadata:\n")
		for k, v := range cred.Metadata {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}
}

func runRevoke(ctx context.Context, p Plugin, args []string, cfg *StandaloneConfig) {
	fs := flag.NewFlagSet("revoke", flag.ExitOnError)
	configFile := fs.String(cfg.ConfigFlag, "", "Path to JSON config file")
	externalID := fs.String("external-id", "", "External ID of credential to revoke")
	fs.Parse(args)

	if *configFile == "" {
		fmt.Fprintln(os.Stderr, "Error: --config is required")
		os.Exit(1)
	}
	if *externalID == "" {
		fmt.Fprintln(os.Stderr, "Error: --external-id is required")
		os.Exit(1)
	}

	configJSON, err := os.ReadFile(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
		os.Exit(1)
	}

	if err := p.Configure(ctx, string(configJSON)); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	if err := p.RevokeCredential(ctx, *externalID); err != nil {
		fmt.Fprintf(os.Stderr, "Error revoking credential: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Credential revoked")
}
