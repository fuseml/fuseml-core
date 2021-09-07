// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
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
	"github.com/google/wire"
	"github.com/timshannon/badgerhold/v3"
	"log"
)

// Injectors from wire.go:

func InitializeCore(logger *log.Logger, storeOptions badgerhold.Options, fuseMLNamespace string) (*coreInit, error) {
	store, err := badgerhold.Open(storeOptions)
	if err != nil {
		return nil, err
	}
	applicationStore := badger.NewApplicationStore(store)
	service := svc.NewApplicationService(logger, applicationStore)
	applicationEndpoints := application.NewEndpoints(service)
	adminClient, err := gitea.NewAdminClient(logger)
	if err != nil {
		return nil, err
	}
	gitCodesetStore := core.NewGitCodesetStore(adminClient)
	codesetService := svc.NewCodesetService(logger, gitCodesetStore)
	codesetEndpoints := codeset.NewEndpoints(codesetService)
	gitProjectStore := core.NewGitProjectStore(adminClient)
	projectService := svc.NewProjectService(logger, gitProjectStore)
	projectEndpoints := project.NewEndpoints(projectService)
	runnableStore := core.NewRunnableStore()
	runnableService := svc.NewRunnableService(logger, runnableStore)
	runnableEndpoints := runnable.NewEndpoints(runnableService)
	versionService := svc.NewVersionService(logger)
	versionEndpoints := version.NewEndpoints(versionService)
	workflowBackend, err := tekton.NewWorkflowBackend(logger, fuseMLNamespace)
	if err != nil {
		return nil, err
	}
	workflowStore := badger.NewWorkflowStore(store)
	extensionStore := badger.NewExtensionStore(store)
	extensionRegistry := manager.NewExtensionRegistry(extensionStore)
	workflowManager := manager.NewWorkflowManager(workflowBackend, workflowStore, gitCodesetStore, extensionRegistry)
	workflowService := svc.NewWorkflowService(logger, workflowManager)
	workflowEndpoints := workflow.NewEndpoints(workflowService)
	extensionService := svc.NewExtensionRegistryService(logger, extensionRegistry)
	extensionEndpoints := extension.NewEndpoints(extensionService)
	mainEndpoints := &endpoints{
		application: applicationEndpoints,
		codeset:     codesetEndpoints,
		project:     projectEndpoints,
		runnable:    runnableEndpoints,
		version:     versionEndpoints,
		workflow:    workflowEndpoints,
		extension:   extensionEndpoints,
	}
	mainCoreInit := &coreInit{
		endpoints: mainEndpoints,
		store:     store,
	}
	return mainCoreInit, nil
}

// wire.go:

var storeSet = wire.NewSet(badgerhold.Open, badger.NewApplicationStore, wire.Bind(new(domain.ApplicationStore), new(*badger.ApplicationStore)), gitea.NewAdminClient, wire.Bind(new(domain.GitAdminClient), new(*gitea.AdminClient)), core.NewGitCodesetStore, wire.Bind(new(domain.CodesetStore), new(*core.GitCodesetStore)), core.NewGitProjectStore, wire.Bind(new(domain.ProjectStore), new(*core.GitProjectStore)), core.NewRunnableStore, wire.Bind(new(domain.RunnableStore), new(*core.RunnableStore)), badger.NewWorkflowStore, wire.Bind(new(domain.WorkflowStore), new(*badger.WorkflowStore)), badger.NewExtensionStore, wire.Bind(new(domain.ExtensionStore), new(*badger.ExtensionStore)))

var managerSet = wire.NewSet(manager.NewWorkflowManager, wire.Bind(new(domain.WorkflowManager), new(*manager.WorkflowManager)), manager.NewExtensionRegistry, wire.Bind(new(domain.ExtensionRegistry), new(*manager.ExtensionRegistry)))

var backendSet = wire.NewSet(tekton.NewWorkflowBackend, wire.Bind(new(domain.WorkflowBackend), new(*tekton.WorkflowBackend)))

var endpointsSet = wire.NewSet(svc.NewApplicationService, application.NewEndpoints, svc.NewCodesetService, codeset.NewEndpoints, svc.NewProjectService, project.NewEndpoints, svc.NewRunnableService, runnable.NewEndpoints, svc.NewVersionService, version.NewEndpoints, svc.NewWorkflowService, workflow.NewEndpoints, svc.NewExtensionRegistryService, extension.NewEndpoints)
