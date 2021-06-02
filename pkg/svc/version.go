package svc

import (
	"context"
	"log"

	gversion "github.com/fuseml/fuseml-core/gen/version"
	"github.com/fuseml/fuseml-core/pkg/version"
)

// version service implementation.
type versionsrvc struct {
	logger *log.Logger
}

// NewVersionService returns the version service implementation.
func NewVersionService(logger *log.Logger) gversion.Service {
	return &versionsrvc{logger}
}

// Retrieve an Codeset from FuseML.
func (s *versionsrvc) Get(ctx context.Context) (res *gversion.VersionInfo, err error) {
	s.logger.Print("version.get")

	v := version.GetInfo()

	return &gversion.VersionInfo{
		Version:        &v.Version,
		GitCommit:      &v.GitCommit,
		BuildDate:      &v.BuildDate,
		GolangVersion:  &v.GoVersion,
		GolangCompiler: &v.Compiler,
		Platform:       &v.Platform,
	}, nil
}
