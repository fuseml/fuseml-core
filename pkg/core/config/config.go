package config

const (
	// FuseMLNamespace is the kubernetes namespace where FuseML workfloads are created
	FuseMLNamespace = "fuseml-workloads"
)

var (
	// DefaultUserNamePrefix is the default prefix for user names created for each project
	DefaultUserNamePrefix = "fuseml"
	// DefaultUserPassword is the default password used when creating new per-project users
	DefaultUserPassword = "changeme"
	// DefaultUserEmailDomain is the default domain for user email
	DefaultUserEmailDomain = "@fuseml.org"

	// HookSecret is the default secret used when creating repository hooks
	HookSecret = "generatedsecret"
)

// DefaultUserName returns default user name for new per-project user
func DefaultUserName(org string) string {
	return DefaultUserNamePrefix + "-" + org
}

// DefaultUserEmail returns default email for new per-project user
func DefaultUserEmail(org string) string {
	return DefaultUserName(org) + DefaultUserEmailDomain
}
