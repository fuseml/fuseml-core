package common

import "fmt"

const (
	// ConfigFileName is the name of the FuseML configuration file (without extension)
	ConfigFileName = "config"
	// ConfigHomeSubdir is the subdirectory where the FuseML configuration files is located
	ConfigHomeSubdir = ".fuseml"
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
	// CurrentProject says which project to use if "project" flag is not passed
	CurrentProject string
	// CurrentCodeset sets which codeset to use if the name is not provided
	CurrentCodeset string
}

// Validate validates the global configuration
func (o *GlobalOptions) Validate() error {

	if o.URL == "" {
		return fmt.Errorf("the URL to the FuseML server must be provided as an argument, or through the FUSEML_SERVER_URL envrionment variable or in the CLI configuration file")
	}

	return nil
}
