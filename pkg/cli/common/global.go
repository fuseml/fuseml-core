package common

const (
	// Version is the CLI version
	Version = "0.1"
	// ConfigFileName is the name of the FuseML configuration file (without extension)
	ConfigFileName = "config"
	// ConfigHomeSubdir is the subdirectory where the FuseML configuration files is located
	ConfigHomeSubdir = ".fuseml"
	// DefaultFuseMLURL is the default URL to use for the FuseML server
	DefaultFuseMLURL = "http://localhost:8000"
	// DefaultHTTPTimeout is the default HTTP timeout value
	DefaultHTTPTimeout = 30
)

// GlobalOptions contains global CLI configuration parameters
type GlobalOptions struct {
	// URL to the FuseML server API
	URL string
	// HTTP timeout value used for REST API calls
	Timeout int
	// Verbose mode prints out additional information
	Verbose bool
}
