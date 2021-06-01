package client

import (
	"context"
	"net/http"
	"sort"
	"time"

	goahttp "goa.design/goa/v3/http"

	workflowc "github.com/fuseml/fuseml-core/gen/http/workflow/client"
	"github.com/fuseml/fuseml-core/gen/workflow"
)

// WorkflowClient holds a client for Workflow
type WorkflowClient struct {
	c *workflowc.Client
}

// NewWorkflowClient initializes a WorkflowClient
func NewWorkflowClient(scheme string, host string, doer goahttp.Doer, encoder func(*http.Request) goahttp.Encoder,
	decoder func(*http.Response) goahttp.Decoder, verbose bool) *WorkflowClient {
	wc := &WorkflowClient{workflowc.NewClient(scheme, host, doer, encoder, decoder, verbose)}
	return wc
}

// Assign a Workflow to a Codeset.
func (wc *WorkflowClient) Assign(name, codesetName, codesetProject string) (err error) {
	request, err := workflowc.BuildAssignPayload(name, codesetName, codesetProject)
	if err != nil {
		return
	}

	_, err = wc.c.Assign()(context.Background(), request)
	return
}

// Create a new Workflow.
func (wc *WorkflowClient) Create(workflowDef string) (*workflow.Workflow, error) {
	request, err := workflowc.BuildCreatePayload(workflowDef)
	if err != nil {
		return nil, err
	}

	response, err := wc.c.Create()(context.Background(), request)
	if err != nil {
		return nil, err
	}

	return response.(*workflow.Workflow), nil
}

// Get a Workflow.
func (wc *WorkflowClient) Get(name string) (*workflow.Workflow, error) {
	request, err := workflowc.BuildGetPayload(name)
	if err != nil {
		return nil, err
	}

	response, err := wc.c.Get()(context.Background(), request)
	if err != nil {
		return nil, err
	}

	return response.(*workflow.Workflow), nil
}

// List Workflows.
func (wc *WorkflowClient) List(name string) ([]*workflow.Workflow, error) {
	request, err := workflowc.BuildListPayload(name)
	if err != nil {
		return nil, err
	}

	wfs, err := wc.c.List()(context.Background(), request)
	if err != nil {
		return nil, err
	}

	return wfs.([]*workflow.Workflow), nil
}

// ListAssignments lists Workflow assignments.
func (wc *WorkflowClient) ListAssignments(name string) ([]*workflow.WorkflowAssignment, error) {
	request, err := workflowc.BuildListAssignmentsPayload(name)
	if err != nil {
		return nil, err
	}

	response, err := wc.c.ListAssignments()(context.Background(), request)
	if err != nil {
		return nil, err
	}

	return response.([]*workflow.WorkflowAssignment), nil
}

// ListRuns lists Workflow runs.
func (wc *WorkflowClient) ListRuns(name, codesetName, codesetProject, status string) ([]*workflow.WorkflowRun, error) {
	request, err := workflowc.BuildListRunsPayload(name, codesetName, codesetProject, status)
	if err != nil {
		return nil, err
	}

	res, err := wc.c.ListRuns()(context.Background(), request)
	if err != nil {
		return nil, err
	}

	wrs := res.([]*workflow.WorkflowRun)
	sortWorkflowRunsByStartTime(wrs)

	return wrs, nil
}

func sortWorkflowRunsByStartTime(wrs []*workflow.WorkflowRun) {
	sort.Sort(workflowRunsByStartTime(wrs))
}

type workflowRunsByStartTime []*workflow.WorkflowRun

func (wrs workflowRunsByStartTime) Len() int      { return len(wrs) }
func (wrs workflowRunsByStartTime) Swap(i, j int) { wrs[i], wrs[j] = wrs[j], wrs[i] }
func (wrs workflowRunsByStartTime) Less(i, j int) bool {
	if wrs[j].StartTime == nil {
		return false
	}
	if wrs[i].StartTime == nil {
		return true
	}
	layout := time.RFC3339
	stj, _ := time.Parse(layout, *wrs[j].StartTime)
	sti, _ := time.Parse(layout, *wrs[i].StartTime)
	return stj.Before(sti)
}
