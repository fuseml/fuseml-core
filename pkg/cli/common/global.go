package common

const (
	// Version is the CLI version
	Version = "0.1"
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
