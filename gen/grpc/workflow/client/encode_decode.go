// Code generated by goa v3.3.1, DO NOT EDIT.
//
// workflow gRPC client encoders and decoders
//
// Command:
// $ goa gen github.com/fuseml/fuseml-core/design

package client

import (
	"context"

	workflowpb "github.com/fuseml/fuseml-core/gen/grpc/workflow/pb"
	workflow "github.com/fuseml/fuseml-core/gen/workflow"
	goagrpc "goa.design/goa/v3/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// BuildListFunc builds the remote method to invoke for "workflow" service
// "list" endpoint.
func BuildListFunc(grpccli workflowpb.WorkflowClient, cliopts ...grpc.CallOption) goagrpc.RemoteFunc {
	return func(ctx context.Context, reqpb interface{}, opts ...grpc.CallOption) (interface{}, error) {
		for _, opt := range cliopts {
			opts = append(opts, opt)
		}
		if reqpb != nil {
			return grpccli.List(ctx, reqpb.(*workflowpb.ListRequest), opts...)
		}
		return grpccli.List(ctx, &workflowpb.ListRequest{}, opts...)
	}
}

// EncodeListRequest encodes requests sent to workflow list endpoint.
func EncodeListRequest(ctx context.Context, v interface{}, md *metadata.MD) (interface{}, error) {
	payload, ok := v.(*workflow.ListPayload)
	if !ok {
		return nil, goagrpc.ErrInvalidType("workflow", "list", "*workflow.ListPayload", v)
	}
	return NewListRequest(payload), nil
}

// DecodeListResponse decodes responses from the workflow list endpoint.
func DecodeListResponse(ctx context.Context, v interface{}, hdr, trlr metadata.MD) (interface{}, error) {
	message, ok := v.(*workflowpb.ListResponse)
	if !ok {
		return nil, goagrpc.ErrInvalidType("workflow", "list", "*workflowpb.ListResponse", v)
	}
	if err := ValidateListResponse(message); err != nil {
		return nil, err
	}
	res := NewListResult(message)
	return res, nil
}

// BuildRegisterFunc builds the remote method to invoke for "workflow" service
// "register" endpoint.
func BuildRegisterFunc(grpccli workflowpb.WorkflowClient, cliopts ...grpc.CallOption) goagrpc.RemoteFunc {
	return func(ctx context.Context, reqpb interface{}, opts ...grpc.CallOption) (interface{}, error) {
		for _, opt := range cliopts {
			opts = append(opts, opt)
		}
		if reqpb != nil {
			return grpccli.Register(ctx, reqpb.(*workflowpb.RegisterRequest), opts...)
		}
		return grpccli.Register(ctx, &workflowpb.RegisterRequest{}, opts...)
	}
}

// EncodeRegisterRequest encodes requests sent to workflow register endpoint.
func EncodeRegisterRequest(ctx context.Context, v interface{}, md *metadata.MD) (interface{}, error) {
	payload, ok := v.(*workflow.Workflow)
	if !ok {
		return nil, goagrpc.ErrInvalidType("workflow", "register", "*workflow.Workflow", v)
	}
	return NewRegisterRequest(payload), nil
}

// DecodeRegisterResponse decodes responses from the workflow register endpoint.
func DecodeRegisterResponse(ctx context.Context, v interface{}, hdr, trlr metadata.MD) (interface{}, error) {
	message, ok := v.(*workflowpb.RegisterResponse)
	if !ok {
		return nil, goagrpc.ErrInvalidType("workflow", "register", "*workflowpb.RegisterResponse", v)
	}
	if err := ValidateRegisterResponse(message); err != nil {
		return nil, err
	}
	res := NewRegisterResult(message)
	return res, nil
}

// BuildGetFunc builds the remote method to invoke for "workflow" service "get"
// endpoint.
func BuildGetFunc(grpccli workflowpb.WorkflowClient, cliopts ...grpc.CallOption) goagrpc.RemoteFunc {
	return func(ctx context.Context, reqpb interface{}, opts ...grpc.CallOption) (interface{}, error) {
		for _, opt := range cliopts {
			opts = append(opts, opt)
		}
		if reqpb != nil {
			return grpccli.Get(ctx, reqpb.(*workflowpb.GetRequest), opts...)
		}
		return grpccli.Get(ctx, &workflowpb.GetRequest{}, opts...)
	}
}

// EncodeGetRequest encodes requests sent to workflow get endpoint.
func EncodeGetRequest(ctx context.Context, v interface{}, md *metadata.MD) (interface{}, error) {
	payload, ok := v.(*workflow.GetPayload)
	if !ok {
		return nil, goagrpc.ErrInvalidType("workflow", "get", "*workflow.GetPayload", v)
	}
	return NewGetRequest(payload), nil
}

// DecodeGetResponse decodes responses from the workflow get endpoint.
func DecodeGetResponse(ctx context.Context, v interface{}, hdr, trlr metadata.MD) (interface{}, error) {
	message, ok := v.(*workflowpb.GetResponse)
	if !ok {
		return nil, goagrpc.ErrInvalidType("workflow", "get", "*workflowpb.GetResponse", v)
	}
	if err := ValidateGetResponse(message); err != nil {
		return nil, err
	}
	res := NewGetResult(message)
	return res, nil
}
