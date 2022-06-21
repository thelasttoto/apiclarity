// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// NewGetAPIEventsEventIDReconstructedSpecDiffParams creates a new GetAPIEventsEventIDReconstructedSpecDiffParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewGetAPIEventsEventIDReconstructedSpecDiffParams() *GetAPIEventsEventIDReconstructedSpecDiffParams {
	return &GetAPIEventsEventIDReconstructedSpecDiffParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewGetAPIEventsEventIDReconstructedSpecDiffParamsWithTimeout creates a new GetAPIEventsEventIDReconstructedSpecDiffParams object
// with the ability to set a timeout on a request.
func NewGetAPIEventsEventIDReconstructedSpecDiffParamsWithTimeout(timeout time.Duration) *GetAPIEventsEventIDReconstructedSpecDiffParams {
	return &GetAPIEventsEventIDReconstructedSpecDiffParams{
		timeout: timeout,
	}
}

// NewGetAPIEventsEventIDReconstructedSpecDiffParamsWithContext creates a new GetAPIEventsEventIDReconstructedSpecDiffParams object
// with the ability to set a context for a request.
func NewGetAPIEventsEventIDReconstructedSpecDiffParamsWithContext(ctx context.Context) *GetAPIEventsEventIDReconstructedSpecDiffParams {
	return &GetAPIEventsEventIDReconstructedSpecDiffParams{
		Context: ctx,
	}
}

// NewGetAPIEventsEventIDReconstructedSpecDiffParamsWithHTTPClient creates a new GetAPIEventsEventIDReconstructedSpecDiffParams object
// with the ability to set a custom HTTPClient for a request.
func NewGetAPIEventsEventIDReconstructedSpecDiffParamsWithHTTPClient(client *http.Client) *GetAPIEventsEventIDReconstructedSpecDiffParams {
	return &GetAPIEventsEventIDReconstructedSpecDiffParams{
		HTTPClient: client,
	}
}

/* GetAPIEventsEventIDReconstructedSpecDiffParams contains all the parameters to send to the API endpoint
   for the get API events event ID reconstructed spec diff operation.

   Typically these are written to a http.Request.
*/
type GetAPIEventsEventIDReconstructedSpecDiffParams struct {

	/* EventID.

	   API event ID

	   Format: uint32
	*/
	EventID uint32

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the get API events event ID reconstructed spec diff params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetAPIEventsEventIDReconstructedSpecDiffParams) WithDefaults() *GetAPIEventsEventIDReconstructedSpecDiffParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the get API events event ID reconstructed spec diff params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetAPIEventsEventIDReconstructedSpecDiffParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the get API events event ID reconstructed spec diff params
func (o *GetAPIEventsEventIDReconstructedSpecDiffParams) WithTimeout(timeout time.Duration) *GetAPIEventsEventIDReconstructedSpecDiffParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get API events event ID reconstructed spec diff params
func (o *GetAPIEventsEventIDReconstructedSpecDiffParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get API events event ID reconstructed spec diff params
func (o *GetAPIEventsEventIDReconstructedSpecDiffParams) WithContext(ctx context.Context) *GetAPIEventsEventIDReconstructedSpecDiffParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get API events event ID reconstructed spec diff params
func (o *GetAPIEventsEventIDReconstructedSpecDiffParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get API events event ID reconstructed spec diff params
func (o *GetAPIEventsEventIDReconstructedSpecDiffParams) WithHTTPClient(client *http.Client) *GetAPIEventsEventIDReconstructedSpecDiffParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get API events event ID reconstructed spec diff params
func (o *GetAPIEventsEventIDReconstructedSpecDiffParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithEventID adds the eventID to the get API events event ID reconstructed spec diff params
func (o *GetAPIEventsEventIDReconstructedSpecDiffParams) WithEventID(eventID uint32) *GetAPIEventsEventIDReconstructedSpecDiffParams {
	o.SetEventID(eventID)
	return o
}

// SetEventID adds the eventId to the get API events event ID reconstructed spec diff params
func (o *GetAPIEventsEventIDReconstructedSpecDiffParams) SetEventID(eventID uint32) {
	o.EventID = eventID
}

// WriteToRequest writes these params to a swagger request
func (o *GetAPIEventsEventIDReconstructedSpecDiffParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param eventId
	if err := r.SetPathParam("eventId", swag.FormatUint32(o.EventID)); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}