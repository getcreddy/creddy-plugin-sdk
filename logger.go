package sdk

import (
	"os"

	"github.com/hashicorp/go-hclog"
)

// Logger returns an hclog.Logger for plugin use.
// When creddy runs with --debug, this logger will output debug messages.
// The logger is automatically configured by go-plugin to match the host's log level.
var Logger hclog.Logger

func init() {
	// Default logger - will be replaced by go-plugin's automatic logger injection
	// when running as a plugin. In standalone mode, check CREDDY_DEBUG.
	level := hclog.Info
	if os.Getenv("CREDDY_DEBUG") == "1" {
		level = hclog.Debug
	}
	
	Logger = hclog.New(&hclog.LoggerOptions{
		Name:   "plugin",
		Level:  level,
		Output: os.Stderr,
	})
}

// SetLogger allows overriding the default logger (used by go-plugin internally)
func SetLogger(l hclog.Logger) {
	if l != nil {
		Logger = l
	}
}

// Debug logs a debug message
func Debug(msg string, args ...interface{}) {
	Logger.Debug(msg, args...)
}

// Info logs an info message
func Info(msg string, args ...interface{}) {
	Logger.Info(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...interface{}) {
	Logger.Warn(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...interface{}) {
	Logger.Error(msg, args...)
}
