package domain

import (
	"context"

	"github.com/fuseml/fuseml-core/gen/application"
)

// ApplicationStore is an inteface to application stores
type ApplicationStore interface {
	Find(context.Context, string) *application.Application
	GetAll(context.Context, *string, *string) ([]*application.Application, error)
	Add(context.Context, *application.Application) (*application.Application, error)
	Delete(context.Context, string) error
}
