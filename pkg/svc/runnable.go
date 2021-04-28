package svc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/fuseml/fuseml-core/gen/runnable"
	"github.com/fuseml/fuseml-core/pkg/domain"
	"github.com/google/uuid"
)

// runnable service example implementation.
// The example methods log the requests and return zero values.
type runnablesrvc struct {
	logger *log.Logger
	store  domain.RunnableStore
}

const (
	errMissingDefaultValue = "default value missing for optional input"
	errMaxOnePassByMode    = "'passByValue' and 'passByReference' cannot both be supplied"
)

// NewRunnable returns the runnable service implementation.
func NewRunnableService(logger *log.Logger, store domain.RunnableStore) runnable.Service {
	return &runnablesrvc{logger, store}
}

func restToDomain(r *runnable.Runnable) (res *domain.Runnable, err error) {

	getStringValueOrDefault := func(value *string, defaultValue string) string {
		if value == nil {
			return defaultValue
		}
		return *value
	}

	getIntValueOrDefault := func(value *int, defaultValue int) int {
		if value == nil {
			return defaultValue
		}
		return *value
	}

	getInputErr := func(name string, err string) error {
		return fmt.Errorf("input [%s]: %s", name, err)
	}

	getOutputErr := func(name string, err string) error {
		return fmt.Errorf("output [%s]: %s", name, err)
	}

	res = &domain.Runnable{
		Id:          getStringValueOrDefault(r.ID, uuid.NewString()),
		Description: r.Description,
		Kind:        r.Kind,
		Container: domain.RunnableContainer{
			Image:      r.Container.Image,
			Env:        r.Container.Env,
			Args:       r.Container.Args,
			Entrypoint: getStringValueOrDefault(r.Container.Entrypoint, ""),
		},
		Inputs:            make(map[string]*domain.RunnableInput, len(r.Inputs)),
		Outputs:           make(map[string]*domain.RunnableOutput, len(r.Outputs)),
		DefaultInputPath:  r.DefaultInputPath,
		DefaultOutputPath: r.DefaultOutputPath,
		Labels:            r.Labels,
	}

	for iName, input := range r.Inputs {

		i := domain.RunnableInput{
			Type:         domain.RunnableArgumentType(input.Type),
			Name:         iName,
			Description:  input.Description,
			Optional:     input.Optional,
			DefaultValue: getStringValueOrDefault(input.DefaultValue, ""),
			Labels:       input.Labels,
		}

		if input.Optional && input.DefaultValue == nil {
			return nil, getInputErr(i.Name, errMissingDefaultValue)
		}

		if i.Type == domain.RAT_OPAQUE {
			// default value passing mode for opaque artifact is by-reference
			i.ValuePass.PassBy = domain.PBM_REFERENCE
		} else {
			// default value passing mode for everything else is by-value
			i.ValuePass.PassBy = domain.PBM_VALUE
		}

		if input.PassByValue != nil {
			i.ValuePass.PassBy = domain.PBM_VALUE
			i.ValuePass.Path = getStringValueOrDefault(input.PassByValue.ToPath, "")
		}
		if input.PassByReference != nil {
			if input.PassByValue != nil {
				return nil, getInputErr(i.Name, errMaxOnePassByMode)
			}
			i.ValuePass.PassBy = domain.PBM_REFERENCE
			i.ValuePass.Path = getStringValueOrDefault(input.PassByValue.ToPath, "")
		}

		if i.Type != domain.RAT_PARAM && input.Artifact != nil {
			i.Artifact = &domain.RunnableArtifactSpec{
				StoreType: domain.ArtifactStoreType(getStringValueOrDefault(input.Artifact.StoreType, "")),
				Store:     getStringValueOrDefault(input.Artifact.Store, ""),
				Project:   getStringValueOrDefault(input.Artifact.Project, ""),
				Name:      getStringValueOrDefault(input.Artifact.Name, ""),
				Version:   getStringValueOrDefault(input.Artifact.Version, ""),
				MinCount:  getIntValueOrDefault(input.Artifact.MinCount, 1),
				MaxCount:  getIntValueOrDefault(input.Artifact.MaxCount, 1),
				Labels:    input.Labels,
			}
		}

		res.Inputs[i.Name] = &i
	}

	for oName, output := range r.Outputs {

		o := domain.RunnableOutput{
			Type:        domain.RunnableArgumentType(output.Type),
			Name:        oName,
			Description: output.Description,
			Labels:      output.Labels,
		}

		if o.Type == domain.RAT_OPAQUE {
			// default value passing mode for opaque artifact is by-reference
			o.ValuePass.PassBy = domain.PBM_REFERENCE
		} else {
			// default value passing mode for everything else is by-value
			o.ValuePass.PassBy = domain.PBM_VALUE
		}

		if output.PassByValue != nil {
			o.ValuePass.PassBy = domain.PBM_VALUE
			o.ValuePass.Path = getStringValueOrDefault(output.PassByValue.FromPath, "")
		}
		if output.PassByReference != nil {
			if output.PassByValue != nil {
				return nil, getOutputErr(o.Name, errMaxOnePassByMode)
			}
			o.ValuePass.PassBy = domain.PBM_REFERENCE
			o.ValuePass.Path = getStringValueOrDefault(output.PassByValue.FromPath, "")
		}

		if o.Type != domain.RAT_PARAM && output.Artifact != nil {
			o.Artifact = &domain.RunnableArtifactSpec{
				StoreType: domain.ArtifactStoreType(getStringValueOrDefault(output.Artifact.StoreType, "")),
				Store:     getStringValueOrDefault(output.Artifact.Store, ""),
				Project:   getStringValueOrDefault(output.Artifact.Project, ""),
				Name:      getStringValueOrDefault(output.Artifact.Name, ""),
				Version:   getStringValueOrDefault(output.Artifact.Version, ""),
				Labels:    output.Labels,
			}
		}

		res.Outputs[o.Name] = &o
	}

	return
}

func domainToRest(r *domain.Runnable) (res *runnable.Runnable) {

	getNilIfEmptyString := func(value string) *string {
		if value == "" {
			return nil
		}
		return &value
	}

	created := r.Created.Format(time.RFC3339)
	res = &runnable.Runnable{
		ID:          &r.Id,
		Description: r.Description,
		Kind:        r.Kind,
		Container: &runnable.RunnableContainer{
			Image:      r.Container.Image,
			Entrypoint: getNilIfEmptyString(r.Container.Entrypoint),
			Env:        r.Container.Env,
			Args:       r.Container.Args,
		},
		Inputs:            make(map[string]*runnable.RunnableInput, len(r.Inputs)),
		Outputs:           make(map[string]*runnable.RunnableOutput, len(r.Outputs)),
		DefaultInputPath:  r.DefaultInputPath,
		DefaultOutputPath: r.DefaultOutputPath,
		Created:           &created,
		Labels:            r.Labels,
	}
	for iName, input := range r.Inputs {
		defaultValue := (*string)(nil)
		if input.Optional {
			defaultValue = &input.DefaultValue
		}
		i := runnable.RunnableInput{
			Type:         string(input.Type),
			Description:  input.Description,
			Optional:     input.Optional,
			DefaultValue: defaultValue,
			Labels:       input.Labels,
		}

		if input.Type != domain.RAT_PARAM && input.Artifact != nil {
			i.Artifact = &runnable.ArtifactArgSpec{
				StoreType: getNilIfEmptyString(string(input.Artifact.StoreType)),
				Store:     getNilIfEmptyString(input.Artifact.Store),
				Name:      getNilIfEmptyString(input.Artifact.Name),
				Project:   getNilIfEmptyString(input.Artifact.Project),
				Version:   getNilIfEmptyString(input.Artifact.Version),
				MinCount:  &input.Artifact.MinCount,
				MaxCount:  &input.Artifact.MaxCount,
			}
		}

		if input.ValuePass.PassBy == domain.PBM_REFERENCE {
			i.PassByReference = &runnable.InputPassByStrategy{
				ToPath: getNilIfEmptyString(input.ValuePass.Path),
			}
		} else if input.ValuePass.PassBy == domain.PBM_VALUE {
			i.PassByValue = &runnable.InputPassByStrategy{
				ToPath: getNilIfEmptyString(input.ValuePass.Path),
			}
		}
		res.Inputs[iName] = &i
	}

	for oName, output := range r.Outputs {
		o := runnable.RunnableOutput{
			Type:        string(output.Type),
			Description: output.Description,
			Labels:      output.Labels,
		}

		if output.Type != domain.RAT_PARAM && output.Artifact != nil {
			o.Artifact = &runnable.ArtifactArgSpec{
				StoreType: getNilIfEmptyString(string(output.Artifact.StoreType)),
				Store:     getNilIfEmptyString(output.Artifact.Store),
				Name:      getNilIfEmptyString(output.Artifact.Name),
				Project:   getNilIfEmptyString(output.Artifact.Project),
				Version:   getNilIfEmptyString(output.Artifact.Version),
			}
		}

		if output.ValuePass.PassBy == domain.PBM_REFERENCE {
			o.PassByReference = &runnable.OutputPassByStrategy{
				FromPath: getNilIfEmptyString(output.ValuePass.Path),
			}
		} else if output.ValuePass.PassBy == domain.PBM_VALUE {
			o.PassByValue = &runnable.OutputPassByStrategy{
				FromPath: getNilIfEmptyString(output.ValuePass.Path),
			}
		}
		res.Outputs[oName] = &o
	}

	return
}

// Retrieve information about runnables registered in FuseML.
func (s *runnablesrvc) List(ctx context.Context, p *runnable.ListPayload) (res []*runnable.Runnable, err error) {
	s.logger.Print("runnable.list")
	idQuery := ""
	if p.ID != nil {
		idQuery = *p.ID
	}
	kindQuery := ""
	if p.Kind != nil {
		kindQuery = *p.Kind
	}
	items, err := s.store.Find(ctx, idQuery, kindQuery, p.Labels)
	res = make([]*runnable.Runnable, 0, len(items))
	for _, r := range items {
		res = append(res, domainToRest(r))
	}
	return res, err
}

// Register a runnable with the FuseML runnable runnableStore.
func (s *runnablesrvc) Register(ctx context.Context, p *runnable.Runnable) (res *runnable.Runnable, err error) {
	s.logger.Print("runnable.register")
	r, err := restToDomain(p)
	if err != nil {
		return p, runnable.MakeBadRequest(err)
	}
	r, err = s.store.Register(ctx, r)
	if err != nil {
		return nil, runnable.MakeBadRequest(err)
	}
	return domainToRest(r), nil
}

// Retrieve a Runnable from FuseML.
func (s *runnablesrvc) Get(ctx context.Context, p *runnable.GetPayload) (res *runnable.Runnable, err error) {
	s.logger.Print("runnable.get")
	r, f_err := s.store.Get(ctx, p.ID)
	if r == nil {
		return nil, runnable.MakeNotFound(errors.New(f_err.Error()))
	}
	return domainToRest(r), nil
}
