package domain

const (
	// ErrProjectExists is the error message returned when trying to create a project (org) that already exists.
	ErrProjectExists = projectErr("Project with that name already exists")
)

type projectErr string

func (e projectErr) Error() string {
	return string(e)
}
