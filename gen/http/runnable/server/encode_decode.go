// Code generated by goa v3.3.1, DO NOT EDIT.
//
// runnable HTTP server encoders and decoders
//
// Command:
// $ goa gen github.com/fuseml/fuseml-core/design

package server

import (
	"context"
	"io"
	"net/http"

	runnable "github.com/fuseml/fuseml-core/gen/runnable"
	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
)

// EncodeListResponse returns an encoder for responses returned by the runnable
// list endpoint.
func EncodeListResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.([]*runnable.Runnable)
		enc := encoder(ctx, w)
		body := NewListResponseBody(res)
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}

// DecodeListRequest returns a decoder for requests sent to the runnable list
// endpoint.
func DecodeListRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			kind *string
		)
		kindRaw := r.URL.Query().Get("kind")
		if kindRaw != "" {
			kind = &kindRaw
		}
		payload := NewListPayload(kind)

		return payload, nil
	}
}

// EncodeListError returns an encoder for errors returned by the list runnable
// endpoint.
func EncodeListError(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder, formatter func(err error) goahttp.Statuser) func(context.Context, http.ResponseWriter, error) error {
	encodeError := goahttp.ErrorEncoder(encoder, formatter)
	return func(ctx context.Context, w http.ResponseWriter, v error) error {
		en, ok := v.(ErrorNamer)
		if !ok {
			return encodeError(ctx, w, v)
		}
		switch en.ErrorName() {
		case "NotFound":
			res := v.(*goa.ServiceError)
			enc := encoder(ctx, w)
			var body interface{}
			if formatter != nil {
				body = formatter(res)
			} else {
				body = NewListNotFoundResponseBody(res)
			}
			w.Header().Set("goa-error", "NotFound")
			w.WriteHeader(http.StatusNotFound)
			return enc.Encode(body)
		default:
			return encodeError(ctx, w, v)
		}
	}
}

// EncodeRegisterResponse returns an encoder for responses returned by the
// runnable register endpoint.
func EncodeRegisterResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*runnable.Runnable)
		enc := encoder(ctx, w)
		body := NewRegisterResponseBody(res)
		w.WriteHeader(http.StatusCreated)
		return enc.Encode(body)
	}
}

// DecodeRegisterRequest returns a decoder for requests sent to the runnable
// register endpoint.
func DecodeRegisterRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body RegisterRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}
		err = ValidateRegisterRequestBody(&body)
		if err != nil {
			return nil, err
		}
		payload := NewRegisterRunnable(&body)

		return payload, nil
	}
}

// EncodeRegisterError returns an encoder for errors returned by the register
// runnable endpoint.
func EncodeRegisterError(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder, formatter func(err error) goahttp.Statuser) func(context.Context, http.ResponseWriter, error) error {
	encodeError := goahttp.ErrorEncoder(encoder, formatter)
	return func(ctx context.Context, w http.ResponseWriter, v error) error {
		en, ok := v.(ErrorNamer)
		if !ok {
			return encodeError(ctx, w, v)
		}
		switch en.ErrorName() {
		case "BadRequest":
			res := v.(*goa.ServiceError)
			enc := encoder(ctx, w)
			var body interface{}
			if formatter != nil {
				body = formatter(res)
			} else {
				body = NewRegisterBadRequestResponseBody(res)
			}
			w.Header().Set("goa-error", "BadRequest")
			w.WriteHeader(http.StatusBadRequest)
			return enc.Encode(body)
		default:
			return encodeError(ctx, w, v)
		}
	}
}

// EncodeGetResponse returns an encoder for responses returned by the runnable
// get endpoint.
func EncodeGetResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(*runnable.Runnable)
		enc := encoder(ctx, w)
		body := NewGetResponseBody(res)
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}

// DecodeGetRequest returns a decoder for requests sent to the runnable get
// endpoint.
func DecodeGetRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			runnableNameOrID string

			params = mux.Vars(r)
		)
		runnableNameOrID = params["runnableNameOrId"]
		payload := NewGetPayload(runnableNameOrID)

		return payload, nil
	}
}

// EncodeGetError returns an encoder for errors returned by the get runnable
// endpoint.
func EncodeGetError(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder, formatter func(err error) goahttp.Statuser) func(context.Context, http.ResponseWriter, error) error {
	encodeError := goahttp.ErrorEncoder(encoder, formatter)
	return func(ctx context.Context, w http.ResponseWriter, v error) error {
		en, ok := v.(ErrorNamer)
		if !ok {
			return encodeError(ctx, w, v)
		}
		switch en.ErrorName() {
		case "BadRequest":
			res := v.(*goa.ServiceError)
			enc := encoder(ctx, w)
			var body interface{}
			if formatter != nil {
				body = formatter(res)
			} else {
				body = NewGetBadRequestResponseBody(res)
			}
			w.Header().Set("goa-error", "BadRequest")
			w.WriteHeader(http.StatusBadRequest)
			return enc.Encode(body)
		case "NotFound":
			res := v.(*goa.ServiceError)
			enc := encoder(ctx, w)
			var body interface{}
			if formatter != nil {
				body = formatter(res)
			} else {
				body = NewGetNotFoundResponseBody(res)
			}
			w.Header().Set("goa-error", "NotFound")
			w.WriteHeader(http.StatusNotFound)
			return enc.Encode(body)
		default:
			return encodeError(ctx, w, v)
		}
	}
}

// marshalRunnableRunnableToRunnableResponse builds a value of type
// *RunnableResponse from a value of type *runnable.Runnable.
func marshalRunnableRunnableToRunnableResponse(v *runnable.Runnable) *RunnableResponse {
	res := &RunnableResponse{
		ID:      v.ID,
		Name:    v.Name,
		Kind:    v.Kind,
		Created: v.Created,
	}
	if v.Image != nil {
		res.Image = marshalRunnableRunnableImageToRunnableImageResponse(v.Image)
	}
	if v.Inputs != nil {
		res.Inputs = make([]*RunnableInputResponse, len(v.Inputs))
		for i, val := range v.Inputs {
			res.Inputs[i] = marshalRunnableRunnableInputToRunnableInputResponse(val)
		}
	}
	if v.Outputs != nil {
		res.Outputs = make([]*RunnableOutputResponse, len(v.Outputs))
		for i, val := range v.Outputs {
			res.Outputs[i] = marshalRunnableRunnableOutputToRunnableOutputResponse(val)
		}
	}
	if v.Labels != nil {
		res.Labels = make([]string, len(v.Labels))
		for i, val := range v.Labels {
			res.Labels[i] = val
		}
	}

	return res
}

// marshalRunnableRunnableImageToRunnableImageResponse builds a value of type
// *RunnableImageResponse from a value of type *runnable.RunnableImage.
func marshalRunnableRunnableImageToRunnableImageResponse(v *runnable.RunnableImage) *RunnableImageResponse {
	res := &RunnableImageResponse{
		RegistryURL: v.RegistryURL,
		Repository:  v.Repository,
		Tag:         v.Tag,
	}

	return res
}

// marshalRunnableRunnableInputToRunnableInputResponse builds a value of type
// *RunnableInputResponse from a value of type *runnable.RunnableInput.
func marshalRunnableRunnableInputToRunnableInputResponse(v *runnable.RunnableInput) *RunnableInputResponse {
	res := &RunnableInputResponse{
		Name: v.Name,
		Kind: v.Kind,
	}
	if v.Runnable != nil {
		res.Runnable = marshalRunnableRunnableRefToRunnableRefResponse(v.Runnable)
	}
	if v.Parameter != nil {
		res.Parameter = marshalRunnableInputParameterToInputParameterResponse(v.Parameter)
	}

	return res
}

// marshalRunnableRunnableRefToRunnableRefResponse builds a value of type
// *RunnableRefResponse from a value of type *runnable.RunnableRef.
func marshalRunnableRunnableRefToRunnableRefResponse(v *runnable.RunnableRef) *RunnableRefResponse {
	if v == nil {
		return nil
	}
	res := &RunnableRefResponse{
		Name: v.Name,
		Kind: v.Kind,
	}
	if v.Labels != nil {
		res.Labels = make([]string, len(v.Labels))
		for i, val := range v.Labels {
			res.Labels[i] = val
		}
	}

	return res
}

// marshalRunnableInputParameterToInputParameterResponse builds a value of type
// *InputParameterResponse from a value of type *runnable.InputParameter.
func marshalRunnableInputParameterToInputParameterResponse(v *runnable.InputParameter) *InputParameterResponse {
	if v == nil {
		return nil
	}
	res := &InputParameterResponse{
		Datatype: v.Datatype,
		Optional: v.Optional,
		Default:  v.Default,
	}

	return res
}

// marshalRunnableRunnableOutputToRunnableOutputResponse builds a value of type
// *RunnableOutputResponse from a value of type *runnable.RunnableOutput.
func marshalRunnableRunnableOutputToRunnableOutputResponse(v *runnable.RunnableOutput) *RunnableOutputResponse {
	res := &RunnableOutputResponse{
		Name: v.Name,
		Kind: v.Kind,
	}
	if v.Runnable != nil {
		res.Runnable = marshalRunnableRunnableRefToRunnableRefResponse(v.Runnable)
	}
	if v.Metadata != nil {
		res.Metadata = marshalRunnableInputParameterToInputParameterResponse(v.Metadata)
	}

	return res
}

// unmarshalRunnableImageRequestBodyToRunnableRunnableImage builds a value of
// type *runnable.RunnableImage from a value of type *RunnableImageRequestBody.
func unmarshalRunnableImageRequestBodyToRunnableRunnableImage(v *RunnableImageRequestBody) *runnable.RunnableImage {
	res := &runnable.RunnableImage{
		RegistryURL: v.RegistryURL,
		Repository:  v.Repository,
		Tag:         v.Tag,
	}

	return res
}

// unmarshalRunnableInputRequestBodyToRunnableRunnableInput builds a value of
// type *runnable.RunnableInput from a value of type *RunnableInputRequestBody.
func unmarshalRunnableInputRequestBodyToRunnableRunnableInput(v *RunnableInputRequestBody) *runnable.RunnableInput {
	res := &runnable.RunnableInput{
		Name: v.Name,
		Kind: v.Kind,
	}
	if v.Runnable != nil {
		res.Runnable = unmarshalRunnableRefRequestBodyToRunnableRunnableRef(v.Runnable)
	}
	if v.Parameter != nil {
		res.Parameter = unmarshalInputParameterRequestBodyToRunnableInputParameter(v.Parameter)
	}

	return res
}

// unmarshalRunnableRefRequestBodyToRunnableRunnableRef builds a value of type
// *runnable.RunnableRef from a value of type *RunnableRefRequestBody.
func unmarshalRunnableRefRequestBodyToRunnableRunnableRef(v *RunnableRefRequestBody) *runnable.RunnableRef {
	if v == nil {
		return nil
	}
	res := &runnable.RunnableRef{
		Name: v.Name,
		Kind: v.Kind,
	}
	if v.Labels != nil {
		res.Labels = make([]string, len(v.Labels))
		for i, val := range v.Labels {
			res.Labels[i] = val
		}
	}

	return res
}

// unmarshalInputParameterRequestBodyToRunnableInputParameter builds a value of
// type *runnable.InputParameter from a value of type
// *InputParameterRequestBody.
func unmarshalInputParameterRequestBodyToRunnableInputParameter(v *InputParameterRequestBody) *runnable.InputParameter {
	if v == nil {
		return nil
	}
	res := &runnable.InputParameter{
		Datatype: v.Datatype,
		Optional: v.Optional,
		Default:  v.Default,
	}

	return res
}

// unmarshalRunnableOutputRequestBodyToRunnableRunnableOutput builds a value of
// type *runnable.RunnableOutput from a value of type
// *RunnableOutputRequestBody.
func unmarshalRunnableOutputRequestBodyToRunnableRunnableOutput(v *RunnableOutputRequestBody) *runnable.RunnableOutput {
	res := &runnable.RunnableOutput{
		Name: v.Name,
		Kind: v.Kind,
	}
	if v.Runnable != nil {
		res.Runnable = unmarshalRunnableRefRequestBodyToRunnableRunnableRef(v.Runnable)
	}
	if v.Metadata != nil {
		res.Metadata = unmarshalInputParameterRequestBodyToRunnableInputParameter(v.Metadata)
	}

	return res
}

// marshalRunnableRunnableImageToRunnableImageResponseBody builds a value of
// type *RunnableImageResponseBody from a value of type *runnable.RunnableImage.
func marshalRunnableRunnableImageToRunnableImageResponseBody(v *runnable.RunnableImage) *RunnableImageResponseBody {
	res := &RunnableImageResponseBody{
		RegistryURL: v.RegistryURL,
		Repository:  v.Repository,
		Tag:         v.Tag,
	}

	return res
}

// marshalRunnableRunnableInputToRunnableInputResponseBody builds a value of
// type *RunnableInputResponseBody from a value of type *runnable.RunnableInput.
func marshalRunnableRunnableInputToRunnableInputResponseBody(v *runnable.RunnableInput) *RunnableInputResponseBody {
	res := &RunnableInputResponseBody{
		Name: v.Name,
		Kind: v.Kind,
	}
	if v.Runnable != nil {
		res.Runnable = marshalRunnableRunnableRefToRunnableRefResponseBody(v.Runnable)
	}
	if v.Parameter != nil {
		res.Parameter = marshalRunnableInputParameterToInputParameterResponseBody(v.Parameter)
	}

	return res
}

// marshalRunnableRunnableRefToRunnableRefResponseBody builds a value of type
// *RunnableRefResponseBody from a value of type *runnable.RunnableRef.
func marshalRunnableRunnableRefToRunnableRefResponseBody(v *runnable.RunnableRef) *RunnableRefResponseBody {
	if v == nil {
		return nil
	}
	res := &RunnableRefResponseBody{
		Name: v.Name,
		Kind: v.Kind,
	}
	if v.Labels != nil {
		res.Labels = make([]string, len(v.Labels))
		for i, val := range v.Labels {
			res.Labels[i] = val
		}
	}

	return res
}

// marshalRunnableInputParameterToInputParameterResponseBody builds a value of
// type *InputParameterResponseBody from a value of type
// *runnable.InputParameter.
func marshalRunnableInputParameterToInputParameterResponseBody(v *runnable.InputParameter) *InputParameterResponseBody {
	if v == nil {
		return nil
	}
	res := &InputParameterResponseBody{
		Datatype: v.Datatype,
		Optional: v.Optional,
		Default:  v.Default,
	}

	return res
}

// marshalRunnableRunnableOutputToRunnableOutputResponseBody builds a value of
// type *RunnableOutputResponseBody from a value of type
// *runnable.RunnableOutput.
func marshalRunnableRunnableOutputToRunnableOutputResponseBody(v *runnable.RunnableOutput) *RunnableOutputResponseBody {
	res := &RunnableOutputResponseBody{
		Name: v.Name,
		Kind: v.Kind,
	}
	if v.Runnable != nil {
		res.Runnable = marshalRunnableRunnableRefToRunnableRefResponseBody(v.Runnable)
	}
	if v.Metadata != nil {
		res.Metadata = marshalRunnableInputParameterToInputParameterResponseBody(v.Metadata)
	}

	return res
}
