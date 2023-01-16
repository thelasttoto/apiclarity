// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// InternalErrorResponse An object that is return in all cases of failures.
//
// swagger:model InternalErrorResponse
type InternalErrorResponse struct {

	// message
	Message string `json:"message,omitempty"`
}

// Validate validates this internal error response
func (m *InternalErrorResponse) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this internal error response based on context it is used
func (m *InternalErrorResponse) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *InternalErrorResponse) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *InternalErrorResponse) UnmarshalBinary(b []byte) error {
	var res InternalErrorResponse
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
