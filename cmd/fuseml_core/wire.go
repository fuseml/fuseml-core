//go:build wireinject
// +build wireinject

package main

import (
	"log"

	"github.com/google/wire"
	"github.com/timshannon/badgerhold/v3"

	"github.com/fuseml/fuseml-core/gen/application"
	"github.com/fuseml/fuseml-core/gen/codeset"
	"github.com/fuseml/fuseml-core/gen/extension"
	"github.com/fuseml/fuseml-core/gen/project"
	"github.com/fuseml/fuseml-core/gen/runnable"
	"github.com/fuseml/fuseml-core/gen/version"
	"github.com/fuseml/fuseml-core/gen/workflow"
	"github.com/fuseml/fuseml-core/pkg/core"
	"github.com/fuseml/fuseml-core/pkg/core/gitea"
	"github.com/fuseml/fuseml-core/pkg/core/manager"
	"github.com/fuseml/fuseml-core/pkg/core/store/badger"
	"github.com/fuseml/fuseml-core/pkg/core/tekton"
	"github.com/fuseml/fuseml-core/pkg/domain"
	"github.com/fuseml/fuseml-core/pkg/svc"
)

var storeSet = wire.NewSet(
	core.NewApplicationStore,
	wire.Bind(new(domain.ApplicationStore), new(*core.ApplicationStore)),
	gitea.NewAdminClient,
	wire.Bind(new(domain.GitAdminClient), new(*gitea.AdminClient)),
	core.NewGitCodesetStore,
	wire.Bind(new(domain.CodesetStore), new(*core.GitCodesetStore)),
	core.NewGitProjectStore,
	wire.Bind(new(domain.ProjectStore), new(*core.GitProjectStore)),
	core.NewRunnableStore,
	wire.Bind(new(domain.RunnableStore), new(*core.RunnableStore)),
	badger.NewWorkflowStore,
	wire.Bind(new(domain.WorkflowStore), new(*badger.WorkflowStore)),
	core.NewExtensionStore,
	wire.Bind(new(domain.ExtensionStore), new(*core.ExtensionStore)),
)

var managerSet = wire.NewSet(
	manager.NewWorkflowManager,
	wire.Bind(new(domain.WorkflowManager), new(*manager.WorkflowManager)),
	manager.NewExtensionRegistry,
	wire.Bind(new(domain.ExtensionRegistry), new(*manager.ExtensionRegistry)),
)

var backendSet = wire.NewSet(
	tekton.NewWorkflowBackend,
	wire.Bind(new(domain.WorkflowBackend), new(*tekton.WorkflowBackend)),
)

var endpointsSet = wire.NewSet(
	svc.NewApplicationService,
	application.NewEndpoints,
	svc.NewCodesetService,
	codeset.NewEndpoints,
	svc.NewProjectService,
	project.NewEndpoints,
	svc.NewRunnableService,
	runnable.NewEndpoints,
	svc.NewVersionService,
	version.NewEndpoints,
	svc.NewWorkflowService,
	workflow.NewEndpoints,
	svc.NewExtensionRegistryService,
	extension.NewEndpoints,
)

func InitializeCore(logger *log.Logger, storeOptions badgerhold.Options, fuseMLNamespace string) (*coreInit, error) {
	wire.Build(
		storeSet,
		managerSet,
		backendSet,
		endpointsSet,
		wire.Struct(new(endpoints), "*"),
		wire.Struct(new(stores), "*"),
		wire.Struct(new(coreInit), "*"),
	)
	return &coreInit{}, nil
}
