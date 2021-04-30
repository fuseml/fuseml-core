package config

const (
	// FuseMlNamespace is the kubernetes namespace where FuseML workfloads are created
	FuseMlNamespace = "fuseml-workloads"
)

var (
	// DefaultUserNamePrefix is the default prefix for user names created for each project
	DefaultUserNamePrefix = "fuseml"
	// DefaultUserPassword is the default password used when creating new per-project users
	DefaultUserPassword = "changeme"
	// DefaultUserEmail is the default value of email used when creating new per-project users
	DefaultUserEmail = "fuseml@fuseml.org"

	// HookSecret is the default secret used when creating repository hooks
	HookSecret = "generatedsecret"
)

// DefaultUserName returns default user name for new per-project user
func DefaultUserName(org string) string {
	return DefaultUserNamePrefix + "-" + org
}
