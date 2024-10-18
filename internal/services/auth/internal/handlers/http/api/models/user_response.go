// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// UserResponse user response
//
// swagger:model UserResponse
type UserResponse struct {

	// age
	Age int64 `json:"age,omitempty"`

	// bio
	Bio string `json:"bio,omitempty"`

	// email
	Email string `json:"email,omitempty"`

	// name
	Name string `json:"name,omitempty"`

	// nickname
	Nickname string `json:"nickname,omitempty"`

	// roles
	Roles []string `json:"roles"`

	// user Id
	UserID string `json:"userId,omitempty"`
}

// Validate validates this user response
func (m *UserResponse) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this user response based on context it is used
func (m *UserResponse) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *UserResponse) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *UserResponse) UnmarshalBinary(b []byte) error {
	var res UserResponse
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
