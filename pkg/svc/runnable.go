package svc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/fuseml/fuseml-core/gen/runnable"
	"github.com/fuseml/fuseml-core/pkg/domain"
)

// runnable service example implementation.
// The example methods log the requests and return zero values.
type runnablesrvc struct {
	logger *log.Logger
	store  domain.RunnableStore
}

const (
	errDuplicateInputName    = "input name must be unique across all input types"
	errDuplicateOutputName   = "output name must be unique across all output types"
	errDuplicateArtifactKind = "at most one artifact kind may be supplied"
	errMissingDefaultValue   = "default value missing for optional input"
	errMaxOneArtifactKind    = "at most one kind of artifact can be configured"
)

// NewRunnableService returns the runnable service implementation.
func NewRunnableService(logger *log.Logger, store domain.RunnableStore) runnable.Service {
	return &runnablesrvc{logger, store}
}

// RunnableInputError defines an error type that applies to a runnable input
type RunnableInputError struct {
	Name string // runnable input name
	Type string // runnable input type
	Err  string // error
}

func (e *RunnableInputError) Error() string {
	return fmt.Sprintf("input %s [%s]: %s", e.Type, e.Name, e.Err)
}

// RunnableOutputError defines an error type that applies to a runnable output
type RunnableOutputError struct {
	Name string // runnable output name
	Type string // runnable input type
	Err  string // error
}

func (e *RunnableOutputError) Error() string {
	return fmt.Sprintf("output %s [%s]: %s", e.Type, e.Name, e.Err)
}

func restToDomain(r *runnable.Runnable) (res *domain.Runnable, err error) {

	getStringValueOrDefault := func(value *string, defaultValue string) string {
		if value == nil {
			return defaultValue
		}
		return *value
	}

	getLocalContainerImage := func(image string) (string, bool) {
		if strings.HasPrefix(image, domain.LOCAL_REGISTRY_HOSTNAME+"/") {
			return strings.TrimPrefix(image, domain.LOCAL_REGISTRY_HOSTNAME+"/"), true
		}
		return image, false
	}

	res = &domain.Runnable{
		ID:          r.ID,
		Description: r.Description,
		Author:      r.Author,
		Source:      r.Source,
		Kind:        r.Kind,
		Container: domain.RunnableContainer{
			Env:        r.Container.Env,
			Args:       r.Container.Args,
			Entrypoint: getStringValueOrDefault(r.Container.Entrypoint, ""),
		},
		Inputs:            make(map[string]interface{}),
		Outputs:           make(map[string]interface{}),
		DefaultInputPath:  getStringValueOrDefault(r.DefaultInputPath, ""),
		DefaultOutputPath: getStringValueOrDefault(r.DefaultOutputPath, ""),
		Labels:            r.Labels,
	}

	res.Container.Image, res.Container.LocalImage = getLocalContainerImage(r.Container.Image)

	if r.Input != nil {
		for pName, param := range r.Input.Parameters {

			if _, iExists := res.Inputs[pName]; iExists {
				return nil, &RunnableInputError{pName, "parameter", errDuplicateInputName}
			}

			p := domain.RunnableInputParameter{
				RunnableArgDesc: domain.RunnableArgDesc{
					Name:        pName,
					Description: param.Description,
					Labels:      param.Labels,
				},
				Optional:     param.Optional,
				DefaultValue: getStringValueOrDefault(param.DefaultValue, ""),
				Path:         getStringValueOrDefault(param.Path, ""),
			}

			if param.Optional && param.DefaultValue == nil {
				return nil, &RunnableInputError{p.Name, "parameter", errMissingDefaultValue}
			}

			res.Inputs[pName] = &p
		}

		for aName, artifact := range r.Input.Artifacts {

			if _, iExists := res.Inputs[aName]; iExists {
				return nil, &RunnableInputError{aName, "artifact", errDuplicateInputName}
			}

			a := domain.RunnableInputArtifact{
				RunnableArtifactArgDesc: domain.RunnableArtifactArgDesc{
					RunnableArgDesc: domain.RunnableArgDesc{
						Name:        aName,
						Description: artifact.Description,
						Labels:      artifact.Labels,
					},
					Dimension: domain.RunnableArtifactArgDimension(artifact.Dimension),
				},
				Optional: artifact.Optional,
				Path:     getStringValueOrDefault(artifact.Path, ""),
			}

			for _, provider := range artifact.Provider {
				a.Provider = append(a.Provider, domain.ArtifactProvider(provider))
			}

			if artifact.Kind == nil {
				// generic artifact
				res.Inputs[aName] = &a
				continue
			}

			if artifact.Kind.Codeset != nil {
				// codeset artifact
				res.Inputs[aName] = &domain.RunnableInputCodeset{
					RunnableInputArtifact: a,
					RunnableCodesetArtifact: domain.RunnableCodesetArtifact{
						Type:     artifact.Kind.Codeset.Type,
						Function: artifact.Kind.Codeset.Function,
						Format:   artifact.Kind.Codeset.Format,
						Software: artifact.Kind.Codeset.Software,
					},
				}
			}

			if artifact.Kind.Model != nil {
				if _, iExists := res.Inputs[aName]; iExists {
					return nil, &RunnableInputError{aName, "artifact", errDuplicateArtifactKind}
				}

				// model artifact
				res.Inputs[aName] = &domain.RunnableInputModel{
					RunnableInputArtifact: a,
					RunnableModelArtifact: domain.RunnableModelArtifact{
						Format:     artifact.Kind.Model.Format,
						Pretrained: artifact.Kind.Model.Pretrained,
						Method:     getStringValueOrDefault(artifact.Kind.Model.Method, ""),
						Class:      getStringValueOrDefault(artifact.Kind.Model.Class, ""),
						Function:   getStringValueOrDefault(artifact.Kind.Model.Function, ""),
						Software:   artifact.Kind.Model.Software,
					},
				}
			}

			if artifact.Kind.Dataset != nil {
				if _, iExists := res.Inputs[aName]; iExists {
					return nil, &RunnableInputError{aName, "artifact", errDuplicateArtifactKind}
				}

				// dataset artifact
				res.Inputs[aName] = &domain.RunnableInputDataset{
					RunnableInputArtifact: a,
					RunnableDatasetArtifact: domain.RunnableDatasetArtifact{
						Type:        artifact.Kind.Dataset.Type,
						Format:      artifact.Kind.Dataset.Format,
						Compression: artifact.Kind.Dataset.Compression,
					},
				}
			}

			if artifact.Kind.Runnable != nil {
				if _, iExists := res.Inputs[aName]; iExists {
					return nil, &RunnableInputError{aName, "artifact", errDuplicateArtifactKind}
				}

				// runnable artifact
				res.Inputs[aName] = &domain.RunnableInputRunnable{
					RunnableInputArtifact: a,
					RunnableRunnableArtifact: domain.RunnableRunnableArtifact{
						Kind: getStringValueOrDefault(artifact.Kind.Runnable.Kind, ""),
					},
				}
			}

		}

	}

	if r.Output != nil {
		for pName, param := range r.Output.Parameters {

			if _, iExists := res.Outputs[pName]; iExists {
				return nil, &RunnableOutputError{pName, "parameter", errDuplicateOutputName}
			}

			p := domain.RunnableOutputParameter{
				RunnableArgDesc: domain.RunnableArgDesc{
					Name:        pName,
					Description: param.Description,
					Labels:      param.Labels,
				},
				Optional:     param.Optional,
				DefaultValue: getStringValueOrDefault(param.DefaultValue, ""),
				Path:         getStringValueOrDefault(param.Path, ""),
			}

			if param.Optional && param.DefaultValue == nil {
				return nil, &RunnableOutputError{p.Name, "parameter", errMissingDefaultValue}
			}

			res.Outputs[pName] = &p
		}

		for aName, artifact := range r.Output.Artifacts {

			if _, iExists := res.Outputs[aName]; iExists {
				return nil, &RunnableOutputError{aName, "artifact", errDuplicateOutputName}
			}

			a := domain.RunnableOutputArtifact{
				RunnableArtifactArgDesc: domain.RunnableArtifactArgDesc{
					RunnableArgDesc: domain.RunnableArgDesc{
						Name:        aName,
						Description: artifact.Description,
						Labels:      artifact.Labels,
					},
					Dimension: domain.RunnableArtifactArgDimension(artifact.Dimension),
				},
				Optional: artifact.Optional,
				Path:     getStringValueOrDefault(artifact.Path, ""),
			}

			for _, provider := range artifact.Provider {
				a.Provider = append(a.Provider, domain.ArtifactProvider(provider))
			}

			if artifact.Kind == nil {
				// generic artifact
				res.Outputs[aName] = &a
				continue
			}

			if artifact.Kind.Codeset != nil {
				// codeset artifact
				res.Outputs[aName] = &domain.RunnableOutputCodeset{
					RunnableOutputArtifact: a,
					RunnableCodesetArtifact: domain.RunnableCodesetArtifact{
						Type:     artifact.Kind.Codeset.Type,
						Function: artifact.Kind.Codeset.Function,
						Format:   artifact.Kind.Codeset.Format,
						Software: artifact.Kind.Codeset.Software,
					},
				}
			}

			if artifact.Kind.Model != nil {
				if _, iExists := res.Outputs[aName]; iExists {
					return nil, &RunnableOutputError{aName, "artifact", errDuplicateArtifactKind}
				}

				// model artifact
				res.Outputs[aName] = &domain.RunnableOutputModel{
					RunnableOutputArtifact: a,
					RunnableModelArtifact: domain.RunnableModelArtifact{
						Format:     artifact.Kind.Model.Format,
						Pretrained: artifact.Kind.Model.Pretrained,
						Method:     getStringValueOrDefault(artifact.Kind.Model.Method, ""),
						Class:      getStringValueOrDefault(artifact.Kind.Model.Class, ""),
						Function:   getStringValueOrDefault(artifact.Kind.Model.Function, ""),
						Software:   artifact.Kind.Model.Software,
					},
				}
			}

			if artifact.Kind.Dataset != nil {
				if _, iExists := res.Outputs[aName]; iExists {
					return nil, &RunnableOutputError{aName, "artifact", errDuplicateArtifactKind}
				}

				// dataset artifact
				res.Outputs[aName] = &domain.RunnableOutputDataset{
					RunnableOutputArtifact: a,
					RunnableDatasetArtifact: domain.RunnableDatasetArtifact{
						Type:        artifact.Kind.Dataset.Type,
						Format:      artifact.Kind.Dataset.Format,
						Compression: artifact.Kind.Dataset.Compression,
					},
				}
			}

			if artifact.Kind.Runnable != nil {
				if _, iExists := res.Outputs[aName]; iExists {
					return nil, &RunnableOutputError{aName, "artifact", errDuplicateArtifactKind}
				}

				// runnable artifact
				res.Outputs[aName] = &domain.RunnableOutputRunnable{
					RunnableOutputArtifact: a,
					RunnableRunnableArtifact: domain.RunnableRunnableArtifact{
						Kind: getStringValueOrDefault(artifact.Kind.Runnable.Kind, ""),
					},
				}
			}

		}

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

	getLocalContainerImage := func(image string, local bool) string {
		if local {
			return image + "/" + domain.LOCAL_REGISTRY_HOSTNAME
		}
		return image
	}

	created := r.Created.Format(time.RFC3339)
	res = &runnable.Runnable{
		ID:          r.ID,
		Created:     &created,
		Description: r.Description,
		Author:      r.Author,
		Source:      r.Source,
		Kind:        r.Kind,
		Container: &runnable.RunnableContainer{
			Image:      getLocalContainerImage(r.Container.Image, r.Container.LocalImage),
			Entrypoint: getNilIfEmptyString(r.Container.Entrypoint),
			Env:        r.Container.Env,
			Args:       r.Container.Args,
		},
		Input: &runnable.RunnableInput{
			Parameters: make(map[string]*runnable.RunnableInputParameter),
			Artifacts:  make(map[string]*runnable.RunnableInputArtifact),
		},
		Output: &runnable.RunnableOutput{
			Parameters: make(map[string]*runnable.RunnableOutputParameter),
			Artifacts:  make(map[string]*runnable.RunnableOutputArtifact),
		},
		DefaultInputPath:  getNilIfEmptyString(r.DefaultInputPath),
		DefaultOutputPath: getNilIfEmptyString(r.DefaultOutputPath),
		Labels:            r.Labels,
	}

	getRestInputArtifact := func(artifact *domain.RunnableInputArtifact) (res *runnable.RunnableInputArtifact) {
		res = &runnable.RunnableInputArtifact{
			Description: artifact.Description,
			Optional:    artifact.Optional,
			Path:        getNilIfEmptyString(artifact.Path),
			Dimension:   string(artifact.Dimension),
			Labels:      artifact.Labels,
		}

		for _, provider := range artifact.Provider {
			res.Provider = append(res.Provider, string(provider))
		}

		return
	}

	for iName, i := range r.Inputs {

		var artifact *runnable.RunnableInputArtifact

		switch input := i.(type) {
		case *domain.RunnableInputParameter:
			defaultValue := (*string)(nil)
			if input.Optional {
				defaultValue = &input.DefaultValue
			}

			res.Input.Parameters[iName] = &runnable.RunnableInputParameter{
				Description:  input.Description,
				Optional:     input.Optional,
				DefaultValue: defaultValue,
				Path:         getNilIfEmptyString(input.Path),
				Labels:       input.Labels,
			}
			continue
		case *domain.RunnableInputArtifact:
			artifact = getRestInputArtifact(input)
		case *domain.RunnableInputCodeset:
			artifact = getRestInputArtifact(&input.RunnableInputArtifact)
			artifact.Kind = &runnable.RunnableInputArtifactKind{
				Codeset: &runnable.CodesetArgumentDesc{
					Type:     input.Type,
					Function: input.Function,
					Format:   input.Format,
					Software: input.Software,
				},
			}
		case *domain.RunnableInputModel:
			artifact = getRestInputArtifact(&input.RunnableInputArtifact)
			artifact.Kind = &runnable.RunnableInputArtifactKind{
				Model: &runnable.ModelArgumentDesc{
					Format:     input.Format,
					Pretrained: input.Pretrained,
					Method:     getNilIfEmptyString(input.Method),
					Class:      getNilIfEmptyString(input.Class),
					Function:   getNilIfEmptyString(input.Function),
					Software:   input.Software,
				},
			}
		case *domain.RunnableInputDataset:
			artifact = getRestInputArtifact(&input.RunnableInputArtifact)
			artifact.Kind = &runnable.RunnableInputArtifactKind{
				Dataset: &runnable.DatasetArgumentDesc{
					Type:        input.Type,
					Format:      input.Format,
					Compression: input.Compression,
				},
			}
		case *domain.RunnableInputRunnable:
			artifact = getRestInputArtifact(&input.RunnableInputArtifact)
			artifact.Kind = &runnable.RunnableInputArtifactKind{
				Runnable: &runnable.RunnableArgumentDesc{
					Kind: getNilIfEmptyString(input.Kind),
				},
			}
		}

		res.Input.Artifacts[iName] = artifact
	}

	getRestOutputArtifact := func(artifact *domain.RunnableOutputArtifact) (res *runnable.RunnableOutputArtifact) {
		res = &runnable.RunnableOutputArtifact{
			Description: artifact.Description,
			Optional:    artifact.Optional,
			Path:        getNilIfEmptyString(artifact.Path),
			Dimension:   string(artifact.Dimension),
			Labels:      artifact.Labels,
		}

		for _, provider := range artifact.Provider {
			res.Provider = append(res.Provider, string(provider))
		}

		return
	}

	for iName, i := range r.Outputs {

		var artifact *runnable.RunnableOutputArtifact

		switch Output := i.(type) {
		case *domain.RunnableOutputParameter:
			defaultValue := (*string)(nil)
			if Output.Optional {
				defaultValue = &Output.DefaultValue
			}

			res.Output.Parameters[iName] = &runnable.RunnableOutputParameter{
				Description:  Output.Description,
				Optional:     Output.Optional,
				DefaultValue: defaultValue,
				Path:         getNilIfEmptyString(Output.Path),
				Labels:       Output.Labels,
			}
			continue
		case *domain.RunnableOutputArtifact:
			artifact = getRestOutputArtifact(Output)
		case *domain.RunnableOutputCodeset:
			artifact = getRestOutputArtifact(&Output.RunnableOutputArtifact)
			artifact.Kind = &runnable.RunnableOutputArtifactKind{
				Codeset: &runnable.CodesetArgumentDesc{
					Type:     Output.Type,
					Function: Output.Function,
					Format:   Output.Format,
					Software: Output.Software,
				},
			}
		case *domain.RunnableOutputModel:
			artifact = getRestOutputArtifact(&Output.RunnableOutputArtifact)
			artifact.Kind = &runnable.RunnableOutputArtifactKind{
				Model: &runnable.ModelArgumentDesc{
					Format:     Output.Format,
					Pretrained: Output.Pretrained,
					Method:     getNilIfEmptyString(Output.Method),
					Class:      getNilIfEmptyString(Output.Class),
					Function:   getNilIfEmptyString(Output.Function),
					Software:   Output.Software,
				},
			}
		case *domain.RunnableOutputDataset:
			artifact = getRestOutputArtifact(&Output.RunnableOutputArtifact)
			artifact.Kind = &runnable.RunnableOutputArtifactKind{
				Dataset: &runnable.DatasetArgumentDesc{
					Type:        Output.Type,
					Format:      Output.Format,
					Compression: Output.Compression,
				},
			}
		case *domain.RunnableOutputRunnable:
			artifact = getRestOutputArtifact(&Output.RunnableOutputArtifact)
			artifact.Kind = &runnable.RunnableOutputArtifactKind{
				Runnable: &runnable.RunnableArgumentDesc{
					Kind: getNilIfEmptyString(Output.Kind),
				},
			}
		}

		res.Output.Artifacts[iName] = artifact
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
	r, err := s.store.Get(ctx, p.ID)
	if r == nil {
		return nil, runnable.MakeNotFound(errors.New(err.Error()))
	}
	return domainToRest(r), nil
}
