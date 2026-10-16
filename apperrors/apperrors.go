package apperrors

import (
	"log"

	"github.com/sendelius/go/env"
)

type Error struct {
	Code    string
	Message string
	Cause   error
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
