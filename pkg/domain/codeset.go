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

// CodesetSubscriber is an interface for objects interested in operations performed on
// a specific codeset
type CodesetSubscriber interface {
	OnDeletingCodeset(ctx context.Context, c *Codeset)
}

// CodesetStore is an interface to codeset stores
type CodesetStore interface {
	Find(ctx context.Context, project, name string) (*Codeset, error)
	GetAll(ctx context.Context, project, label *string) ([]*Codeset, error)
	Add(ctx context.Context, c *Codeset) (*Codeset, *string, *string, error)
	CreateWebhook(context.Context, *Codeset, string) (*int64, error)
	DeleteWebhook(context.Context, *Codeset, *int64) error
	Delete(ctx context.Context, project, name string) error
	Subscribe(ctx context.Context, watcher CodesetSubscriber, codeset *Codeset) error
	Unsubscribe(ctx context.Context, watcher CodesetSubscriber, codeset *Codeset) error
}

// GitAdminClient describes the interface of a Git admin client
type GitAdminClient interface {
	PrepareRepository(*Codeset, *string) (*string, *string, error)
	CreateRepoWebhook(string, string, *string) (*int64, error)
	DeleteRepoWebhook(string, string, *int64) error
	GetRepositories(org, label *string) ([]*Codeset, error)
	GetRepository(org, name string) (*Codeset, error)
	DeleteRepository(org, name string) error
	GetProjects() ([]*Project, error)
	GetProject(org string) (*Project, error)
	DeleteProject(org string) error
	CreateProject(string, string, bool) (*Project, error)
}
