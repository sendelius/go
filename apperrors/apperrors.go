package apperrors

import (
	"log"

	"github.com/sendelius/go/env"
)

type Error struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Field   string         `json:"field,omitempty"`
	Status  int            `json:"-"`
	Cause   error          `json:"-"`
	Data    map[string]any `json:"-"`
}

type ErrorBuilder struct {
	err *Error
}

func New(code string, message string) *ErrorBuilder {
	return &ErrorBuilder{
		err: &Error{
			Code:    code,
			Message: message,
		},
	}
}

func (b *ErrorBuilder) Cause(cause error) *ErrorBuilder {
	b.err.Cause = cause
	if cause != nil && env.Bool("DEV_MODE") {
		log.Printf("%s; причина: %+v", b.err.Error(), b.err.Cause)
	}
	return b
}

func (b *ErrorBuilder) Field(field string) *ErrorBuilder {
	b.err.Field = field
	return b
}

func (b *ErrorBuilder) Status(status int) *ErrorBuilder {
	b.err.Status = status
	return b
}

func (b *ErrorBuilder) Data(data map[string]any) *ErrorBuilder {
	b.err.Data = data
	return b
}

func (b *ErrorBuilder) Build() *Error {
	return b.err
}

func (e *Error) Error() string {
	return e.Code + ": " + e.Message
}

func (e *Error) Unwrap() error {
	return e.Cause
}
