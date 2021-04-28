package domain

import (
	"context"

	"github.com/fuseml/fuseml-core/gen/codeset"
)

// CodesetStore is an inteface to codeset stores
type CodesetStore interface {
	Find(ctx context.Context, project, name string) (*codeset.Codeset, error)
	GetAll(ctx context.Context, project, label *string) ([]*codeset.Codeset, error)
	Add(ctx context.Context, c *codeset.Codeset) (*codeset.Codeset, error)
}
