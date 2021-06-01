package domain

import (
	"context"
)

// Codeset represents a codeset artifact
type Codeset struct {
	// The name of the Codeset
	Name string
	// The project this Codeset belongs to
	Project string
	// Codeset description
	Description string
	// Additional Codeset labels that helps with identifying the type
	Labels []string
	// Full URL to the Codeset
	URL string
}

// CodesetStore is an inteface to codeset stores
type CodesetStore interface {
	Find(ctx context.Context, project, name string) (*Codeset, error)
	GetAll(ctx context.Context, project, label *string) ([]*Codeset, error)
	Add(ctx context.Context, c *Codeset) (*Codeset, *string, *string, error)
	CreateWebhook(context.Context, *Codeset, string) (*int64, error)
	DeleteWebhook(context.Context, *Codeset, *int64) error
	Delete(ctx context.Context, project, name string) error
}
