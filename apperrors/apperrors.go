package apperrors

import (
	"log"

	"github.com/sendelius/go/env"
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Cause   error  `json:"-"`
}

func New(code, message string, cause ...error) *Error {
	err := &Error{
		Code:    code,
		Message: message,
	}
	if len(cause) > 0 {
		err.Cause = cause[0]
		if err.Cause != nil && env.Bool("DEV_MODE") {
			log.Printf("%s; причина: %+v", err.Error(), err.Cause)
		}
	}
	return err
}

func (e *Error) Error() string {
	return e.Code + ": " + e.Message
}

func (e *Error) Unwrap() error {
	return e.Cause
}
