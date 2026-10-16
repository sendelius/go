package users

import "github.com/sendelius/go/infrastructure"

type Handler struct {
	infrastructure.Handler
}

func newHandler() *Handler {
	return &Handler{}
}
