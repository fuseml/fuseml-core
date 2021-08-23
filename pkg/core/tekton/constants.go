package tekton

const (
	pipelineRunPrefix         = "fuseml-"
	pipelineRunServiceAccount = "fuseml-workloads"
	triggersServiceAccount    = "tekton-triggers"
	workspaceAccessMode       = "ReadWriteOnce"
	workspaceSize             = "2Gi"
	codesetWorkspaceName      = "source"
	builderTaskName           = "kaniko"
	builderPrepTaskName       = "builder-prep"
	cloneTaskName             = "clone"
	codesetNameParam          = "codeset-name"
	codesetVersionParam       = "codeset-version"
	codesetProjectParam       = "codeset-project"
	codesetURLParam           = "codeset-url"
	fuseMLRegistry            = "registry.fuseml-registry"
	fuseMLRegistryLocal       = "127.0.0.1:30500"
	imageParamName            = "IMAGE"
	stepOutputVarName         = "TASK_RESULT"
	inputsVarPrefix           = "FUSEML_"
	envVarPrefix              = "FUSEML_ENV_"
	stepDefaultCmd            = "run"

	// LabelCodesetName is the label key for the codeset name
	LabelCodesetName = "fuseml/codeset-name"
	// LabelCodesetProject is the label key for the codeset project
	LabelCodesetProject = "fuseml/codeset-project"
	// LabelCodesetVersion is the label key for the codeset version
	LabelCodesetVersion = "fuseml/codeset-version"
	// LabelWorkflowRef is the label key for the reference of the workflow
	LabelWorkflowRef = "fuseml/workflow-ref"
)
