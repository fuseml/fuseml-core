package domain

import (
	"context"
)

// Project represents a project artifact
type Project struct {
	// The name of the Project
	Name string
	// Project description
	Description string
	// Users assigned to the project
	Users []*User
}

// User represents user assigned to the project
type User struct {
	Name  string
	Email string
}

// ProjectStore is an inteface to project stores
type ProjectStore interface {
	Find(ctx context.Context, name string) (*Project, error)
	GetAll(ctx context.Context) ([]*Project, error)
	Delete(ctx context.Context, name string) error
}
