package svc

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/fuseml/fuseml-core/gen/workflow"
	"github.com/fuseml/fuseml-core/pkg/domain"
)

// workflow service example implementation.
// The example methods log the requests and return zero values.
type workflowsrvc struct {
	logger *log.Logger
	mgr    domain.WorkflowManager
}

// NewWorkflowService returns the workflow service implementation.
func NewWorkflowService(logger *log.Logger, workflowManager domain.WorkflowManager) workflow.Service {
	return &workflowsrvc{logger, workflowManager}
}

// List Workflows.
func (s *workflowsrvc) List(ctx context.Context, w *workflow.ListPayload) (res []*workflow.Workflow, err error) {
	s.logger.Print("workflow.list")
	workflows := s.mgr.List(ctx, w.Name)
	for _, w := range workflows {
		res = append(res, workflowDomainToRest(w))
	}
	return
}

// Create a new Workflow.
func (s *workflowsrvc) Create(ctx context.Context, w *workflow.Workflow) (res *workflow.Workflow, err error) {
	s.logger.Print("workflow.create")
	wf, err := s.mgr.Create(ctx, workflowRestToDomain(w))
	if err != nil {
		s.logger.Print(err)
		if err == domain.ErrWorkflowExists {
			return nil, workflow.MakeConflict(err)
		}
		return nil, err
	}
	return workflowDomainToRest(wf), nil
}

// Get a Workflow.
func (s *workflowsrvc) Get(ctx context.Context, w *workflow.GetPayload) (res *workflow.Workflow, err error) {
	s.logger.Print("workflow.get")
	wf, err := s.mgr.Get(ctx, w.Name)
	if err != nil {
		s.logger.Print(err)
		if err == domain.ErrWorkflowNotFound {
			return nil, workflow.MakeNotFound(err)
		}
		return nil, err
	}
	return workflowDomainToRest(wf), nil
}

// Delete a Workflow and its assignments.
func (s *workflowsrvc) Delete(ctx context.Context, d *workflow.DeletePayload) (err error) {
	s.logger.Print("workflow.delete")
	err = s.mgr.Delete(ctx, d.Name)
	if err != nil {
		s.logger.Print(err)
		return
	}
	return
}

// Assign a Workflow to a Codeset.
func (s *workflowsrvc) Assign(ctx context.Context, w *workflow.AssignPayload) (err error) {
	s.logger.Print("workflow.assign")
	_, _, err = s.mgr.AssignToCodeset(ctx, w.Name, w.CodesetProject, w.CodesetName)
	if err != nil {
		s.logger.Print(err)
		// FIXME: codeset needs to thrown a known error when trying to get a codeset that does not exist
		// to properly compare the returned error.
		if err == domain.ErrWorkflowNotFound || strings.Contains(err.Error(), "Fetching Codeset failed") {
			return workflow.MakeNotFound(err)
		}
	}
	return
}

// Unassign a Workflow from a Codeset.
func (s *workflowsrvc) Unassign(ctx context.Context, u *workflow.UnassignPayload) (err error) {
	s.logger.Print("workflow.unassign")
	err = s.mgr.UnassignFromCodeset(ctx, u.Name, u.CodesetProject, u.CodesetName)
	if err != nil {
		s.logger.Print(err)
		if err == domain.ErrWorkflowNotFound || strings.Contains(err.Error(), "Fetching Codeset failed") || err == domain.ErrWorkflowNotAssignedToCodeset {
			return workflow.MakeNotFound(err)
		}
	}
	return
}

// ListAssignments lists Workflow assignments.
func (s *workflowsrvc) ListAssignments(ctx context.Context, w *workflow.ListAssignmentsPayload) (assignments []*workflow.WorkflowAssignment, err error) {
	s.logger.Print("workflow.listAssignments")
	domainAssignments, err := s.mgr.ListAssignments(ctx, w.Name)
	if err != nil {
		return nil, err
	}

	assignments = make([]*workflow.WorkflowAssignment, len(domainAssignments))
	for i, domainAssignment := range domainAssignments {
		assignments[i] = workflowAssignmentDomainToRest(domainAssignment)
	}
	return
}

// List Workflow runs.
func (s *workflowsrvc) ListRuns(ctx context.Context, w *workflow.ListRunsPayload) ([]*workflow.WorkflowRun, error) {
	s.logger.Print("workflow.listRuns")
	filter := domain.WorkflowRunFilter{WorkflowName: w.Name}
	if w.CodesetName != nil {
		filter.CodesetName = *w.CodesetName
	}
	if w.CodesetProject != nil {
		filter.CodesetProject = *w.CodesetProject
	}
	if w.Status != nil {
		filter.Status = []string{*w.Status}
	}
	domainRuns, err := s.mgr.ListRuns(ctx, &filter)
	if err != nil {
		return nil, err
	}
	return workflowRunsDomainToRest(domainRuns), nil
}

func workflowRestToDomain(restWf *workflow.Workflow) *domain.Workflow {
	wf := &domain.Workflow{
		Name:        restWf.Name,
		Description: derefString(restWf.Description),
		Inputs:      workflowInputsRestToDomain(restWf.Inputs),
		Outputs:     workflowOutputsRestToDomain(restWf.Outputs),
		Steps:       workflowStepsRestToDomain(restWf.Steps),
	}
	return wf
}

func workflowInputsRestToDomain(restInputs []*workflow.WorkflowInput) []*domain.WorkflowInput {
	inputs := make([]*domain.WorkflowInput, len(restInputs))
	for i, restInput := range restInputs {
		inputs[i] = &domain.WorkflowInput{
			Name:        restInput.Name,
			Description: derefString(restInput.Description),
			Type:        domain.WorkflowIOType(derefString(restInput.Type)),
			Default:     derefString(restInput.Default),
			Labels:      restInput.Labels,
		}
	}
	return inputs
}

func workflowOutputsRestToDomain(restOutputs []*workflow.WorkflowOutput) []*domain.WorkflowOutput {
	outputs := make([]*domain.WorkflowOutput, len(restOutputs))
	for i, restOutput := range restOutputs {
		outputs[i] = &domain.WorkflowOutput{
			Name:        restOutput.Name,
			Description: derefString(restOutput.Description),
			Type:        domain.WorkflowIOType(derefString(restOutput.Type)),
		}
	}
	return outputs
}

func workflowStepsRestToDomain(restSteps []*workflow.WorkflowStep) []*domain.WorkflowStep {
	steps := make([]*domain.WorkflowStep, len(restSteps))
	for i, restStep := range restSteps {
		steps[i] = &domain.WorkflowStep{
			Name:    restStep.Name,
			Image:   restStep.Image,
			Inputs:  workflowStepInputsRestToDomain(restStep.Inputs),
			Outputs: workflowStepOutputsRestToDomain(restStep.Outputs),
			Env:     workflowStepEnvsRestToDomain(restStep.Env),
		}
	}
	return steps
}

func workflowStepInputsRestToDomain(restStepInputs []*workflow.WorkflowStepInput) []*domain.WorkflowStepInput {
	inputs := make([]*domain.WorkflowStepInput, len(restStepInputs))
	for i, restStepInput := range restStepInputs {
		domainStepInput := domain.WorkflowStepInput{
			Name:  restStepInput.Name,
			Value: derefString(restStepInput.Value),
		}
		if restStepInput.Codeset != nil {
			domainStepInput.Codeset = &domain.WorkflowStepInputCodeset{
				Name: restStepInput.Codeset.Name,
				Path: derefString(restStepInput.Codeset.Path),
			}
		}
		inputs[i] = &domainStepInput
	}
	return inputs
}

func workflowStepOutputsRestToDomain(restStepOutputs []*workflow.WorkflowStepOutput) []*domain.WorkflowStepOutput {
	outputs := make([]*domain.WorkflowStepOutput, len(restStepOutputs))
	for i, restStepOutput := range restStepOutputs {
		domainStepOutput := domain.WorkflowStepOutput{
			Name: restStepOutput.Name,
		}
		if restStepOutput.Image != nil {
			domainStepOutput.Image = &domain.WorkflowStepOutputImage{
				Name:       restStepOutput.Image.Name,
				Dockerfile: derefString(restStepOutput.Image.Dockerfile),
			}
		}
		outputs[i] = &domainStepOutput
	}
	return outputs
}

func workflowStepEnvsRestToDomain(restEnvs []*workflow.WorkflowStepEnv) []*domain.WorkflowStepEnv {
	envs := make([]*domain.WorkflowStepEnv, len(restEnvs))
	for i, restEnv := range restEnvs {
		envs[i] = &domain.WorkflowStepEnv{
			Name:  restEnv.Name,
			Value: restEnv.Value,
		}
	}
	return envs
}

func workflowDomainToRest(wf *domain.Workflow) *workflow.Workflow {
	created := wf.Created.Format(time.RFC3339)
	return &workflow.Workflow{
		Created:     &created,
		Name:        wf.Name,
		Description: refString(wf.Description),
		Inputs:      workflowInputsDomainToRest(wf.Inputs),
		Outputs:     workflowOutputsDomainToRest(wf.Outputs),
		Steps:       workflowStepsDomainToRest(wf.Steps),
	}
}

func workflowInputsDomainToRest(domainInputs []*domain.WorkflowInput) []*workflow.WorkflowInput {
	restInputs := make([]*workflow.WorkflowInput, len(domainInputs))
	for i, domainInput := range domainInputs {
		restInputs[i] = workflowInputDomainToRest(domainInput)
	}
	return restInputs
}

func workflowInputDomainToRest(domainInput *domain.WorkflowInput) *workflow.WorkflowInput {
	return &workflow.WorkflowInput{
		Name:        domainInput.Name,
		Description: refString(domainInput.Description),
		Type:        refString(domainInput.Type.String()),
		Default:     refString(domainInput.Default),
		Labels:      domainInput.Labels,
	}
}

func workflowOutputsDomainToRest(domainOutputs []*domain.WorkflowOutput) []*workflow.WorkflowOutput {
	restOutputs := make([]*workflow.WorkflowOutput, len(domainOutputs))
	for i, domainOutput := range domainOutputs {
		restOutputs[i] = workflowOutputDomainToRest(domainOutput)
	}
	return restOutputs
}

func workflowOutputDomainToRest(domainOutput *domain.WorkflowOutput) *workflow.WorkflowOutput {
	return &workflow.WorkflowOutput{
		Name:        domainOutput.Name,
		Description: refString(domainOutput.Description),
		Type:        refString(domainOutput.Type.String()),
	}
}

func workflowStepsDomainToRest(domainSteps []*domain.WorkflowStep) []*workflow.WorkflowStep {
	restSteps := make([]*workflow.WorkflowStep, len(domainSteps))
	for i, domainStep := range domainSteps {
		restSteps[i] = &workflow.WorkflowStep{
			Name:    domainStep.Name,
			Image:   domainStep.Image,
			Inputs:  workflowStepInputsDomainToRest(domainStep.Inputs),
			Outputs: workflowStepOutputsDomainToRest(domainStep.Outputs),
			Env:     workflowStepEnvsDomainToRest(domainStep.Env),
		}
	}
	return restSteps
}

func workflowStepInputsDomainToRest(domainStepInputs []*domain.WorkflowStepInput) []*workflow.WorkflowStepInput {
	restStepInputs := make([]*workflow.WorkflowStepInput, len(domainStepInputs))
	for i, domainStepInput := range domainStepInputs {
		restStepInput := workflow.WorkflowStepInput{
			Name:  domainStepInput.Name,
			Value: refString(domainStepInput.Value),
		}
		if domainStepInput.Codeset != nil {
			restStepInput.Codeset = &workflow.WorkflowStepInputCodeset{
				Name: domainStepInput.Codeset.Name,
				Path: refString(domainStepInput.Codeset.Path),
			}
		}
		restStepInputs[i] = &restStepInput
	}
	return restStepInputs
}

func workflowStepOutputsDomainToRest(domainStepOutputs []*domain.WorkflowStepOutput) []*workflow.WorkflowStepOutput {
	restStepOutputs := make([]*workflow.WorkflowStepOutput, len(domainStepOutputs))
	for i, domainStepOutput := range domainStepOutputs {
		restStepOutput := workflow.WorkflowStepOutput{
			Name: domainStepOutput.Name,
		}
		if domainStepOutput.Image != nil {
			restStepOutput.Image = &workflow.WorkflowStepOutputImage{
				Name:       domainStepOutput.Image.Name,
				Dockerfile: refString(domainStepOutput.Image.Dockerfile),
			}
		}
		restStepOutputs[i] = &restStepOutput
	}
	return restStepOutputs
}

func workflowStepEnvsDomainToRest(domainStepEnvs []*domain.WorkflowStepEnv) []*workflow.WorkflowStepEnv {
	restStepEnvs := make([]*workflow.WorkflowStepEnv, len(domainStepEnvs))
	for i, domainStepEnv := range domainStepEnvs {
		restStepEnv := workflow.WorkflowStepEnv{
			Name:  domainStepEnv.Name,
			Value: domainStepEnv.Value,
		}
		restStepEnvs[i] = &restStepEnv
	}
	return restStepEnvs
}

func workflowAssignmentDomainToRest(domainAssignment *domain.WorkflowAssignment) *workflow.WorkflowAssignment {
	restCodesets := make([]*workflow.Codeset, len(domainAssignment.Codesets))
	for i, domainCodeset := range domainAssignment.Codesets {
		restCodesets[i] = (*workflow.Codeset)(codesetDomainToRest(domainCodeset))
	}

	restAssignment := workflow.WorkflowAssignment{
		Workflow: domainAssignment.Workflow,
		Codesets: restCodesets,
		Status: &workflow.WorkflowAssignmentStatus{
			Available: domainAssignment.Status.Available,
			URL:       refString(domainAssignment.Status.URL),
		},
	}
	return &restAssignment
}

func workflowRunsDomainToRest(domainRuns []*domain.WorkflowRun) []*workflow.WorkflowRun {
	restRuns := make([]*workflow.WorkflowRun, len(domainRuns))
	for i, domainRun := range domainRuns {
		restRuns[i] = workflowRunDomainToRest(domainRun)
	}
	return restRuns
}

func workflowRunDomainToRest(domainRun *domain.WorkflowRun) *workflow.WorkflowRun {
	return &workflow.WorkflowRun{
		Name:           domainRun.Name,
		WorkflowRef:    domainRun.WorkflowRef,
		Inputs:         workflowRunInputsDomainToRest(domainRun.Inputs),
		Outputs:        workflowRunOutputsDomainToRest(domainRun.Outputs),
		StartTime:      domainRun.StartTime.Format(time.RFC3339),
		CompletionTime: domainRun.CompletionTime.Format(time.RFC3339),
		Status:         domainRun.Status,
		URL:            refString(domainRun.URL),
	}
}

func workflowRunInputsDomainToRest(domainRunInputs []*domain.WorkflowRunInput) []*workflow.WorkflowRunInput {
	restRunInputs := make([]*workflow.WorkflowRunInput, len(domainRunInputs))
	for i, domainRunInput := range domainRunInputs {
		restRunInputs[i] = &workflow.WorkflowRunInput{
			Input: workflowInputDomainToRest(domainRunInput.Input),
			Value: domainRunInput.Value,
		}
	}
	return restRunInputs
}

func workflowRunOutputsDomainToRest(domainRunOutputs []*domain.WorkflowRunOutput) []*workflow.WorkflowRunOutput {
	restRunOutputs := make([]*workflow.WorkflowRunOutput, len(domainRunOutputs))
	for i, domainRunOutput := range domainRunOutputs {
		restRunOutputs[i] = &workflow.WorkflowRunOutput{
			Output: workflowOutputDomainToRest(domainRunOutput.Output),
			Value:  domainRunOutput.Value,
		}
	}
	return restRunOutputs
}

func refString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func derefString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}
