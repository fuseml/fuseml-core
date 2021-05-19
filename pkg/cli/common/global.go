package common

const (
	Version = "0.1"
)

// GlobalOptions contains global CLI configuration parameters
type GlobalOptions struct {
	Url     string
	Timeout int
	Verbose bool
}
