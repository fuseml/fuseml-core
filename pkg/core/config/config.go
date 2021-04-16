package config

var (
	DefaultUserNamePrefix = "fuseml"
	DefaultUserPassword   = "changeme"
	DefaultUserEmail      = "fuseml@fuseml.org"

	// FIXME: generate this and put it in a secret
	HookSecret = "generatedsecret"

	// FIXME: detect this based on namespaces and services
	StagingEventListenerURL = "http://el-mlflow-listener.fuseml-workloads:8080"
)

func DefaultUserName(org string) string {
	return DefaultUserNamePrefix + "-" + org
}
