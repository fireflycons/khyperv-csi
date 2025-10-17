package rest

import (
	"google.golang.org/grpc/codes"
)

type Error struct {
	// GRPC error code
	Code codes.Code `json:"code" swaggertype:"primitive,integer"`

	// Error message
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}

func (*Error) Is(err error) bool {

	if err == nil {
		return false
	}

	_, ok := err.(*Error)
	return ok
}

func NewError(code codes.Code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}
