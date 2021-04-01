package runnable

import (
	"log"

	"github.com/fuseml/fuseml-core/pkg/api/operations"
	"github.com/fuseml/fuseml-core/pkg/api/operations/runnable"
	"github.com/fuseml/fuseml-core/pkg/models"
	"github.com/go-openapi/runtime/middleware"
	"github.com/google/uuid"
)

func GetRunnableHandler(params runnable.GetRunnableParams) middleware.Responder {
	var r *models.Runnable

	id, err := uuid.Parse(params.RunnableNameOrID)
	if err == nil {
		r = store.FindRunnable(id)
	}

	if r == nil {
		return runnable.NewGetRunnableNotFound()
	}
	return runnable.NewGetRunnableOK().WithPayload(r)
}

func ListRunnablesHandler(params runnable.ListRunnablesParams) middleware.Responder {
	if params.ID == nil {
		log.Println("empty id given, returning all items")
		return runnable.NewListRunnablesOK().WithPayload(store.GetAllRunnables())
	} else {
		log.Printf("id: %s", *params.ID)
		// now we should filter by ID (but it does not make sense to ask for ID in list op)
		return runnable.NewListRunnablesOK().WithPayload(store.GetAllRunnables())
	}
}

func RegisterRunnableHandler(params runnable.RegisterRunnableParams) middleware.Responder {
	r := &models.Runnable{
		Image:   params.Runnable.Image,
		Inputs:  params.Runnable.Inputs,
		Kind:    params.Runnable.Kind,
		Labels:  params.Runnable.Labels,
		Name:    params.Runnable.Name,
		Outputs: params.Runnable.Outputs,
	}
	r = store.AddRunnable(r)
	return runnable.NewRegisterRunnableCreated().WithPayload(r)
}

func RegisterHandlers(api *operations.FusemlAPI) {
	api.RunnableGetRunnableHandler = runnable.GetRunnableHandlerFunc(GetRunnableHandler)
	api.RunnableListRunnablesHandler = runnable.ListRunnablesHandlerFunc(ListRunnablesHandler)
	api.RunnableRegisterRunnableHandler = runnable.RegisterRunnableHandlerFunc(RegisterRunnableHandler)
}
