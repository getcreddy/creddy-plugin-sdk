package sdk

import (
	"os"

	"github.com/hashicorp/go-plugin"
)

// HandshakeConfig is used to validate the plugin and host are compatible
var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "CREDDY_PLUGIN",
	MagicCookieValue: "creddy",
}

// PluginMap is the map of plugin types Creddy supports
var PluginMap = map[string]plugin.Plugin{
	"credential": &CredentialGRPCPlugin{},
}

// Serve starts the plugin server. Call this from your plugin's main().
//
// Example:
//
//	func main() {
//	    sdk.Serve(&MyPlugin{})
//	}
func Serve(p Plugin) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"credential": &CredentialGRPCPlugin{Impl: p},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}

// ServeWithStandalone starts the plugin with optional standalone CLI mode.
// When CREDDY_PLUGIN_STANDALONE=1, it runs in standalone mode for testing.
//
// Example:
//
//	func main() {
//	    sdk.ServeWithStandalone(&MyPlugin{}, &sdk.StandaloneConfig{
//	        ConfigFlag: "config",
//	    })
//	}
func ServeWithStandalone(p Plugin, cfg *StandaloneConfig) {
	if os.Getenv("CREDDY_PLUGIN_STANDALONE") == "1" || len(os.Args) > 1 {
		runStandalone(p, cfg)
		return
	}
	Serve(p)
}

// StandaloneConfig configures standalone mode behavior
type StandaloneConfig struct {
	// ConfigFlag is the flag name for config file (default: "config")
	ConfigFlag string
}
