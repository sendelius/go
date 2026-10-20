package users

import (
	"github.com/sendelius/go/database"
	"github.com/sendelius/go/server"
)

type Module struct {
	service *Service[*User]
}

func New(db *database.Service, hooks ...Hooks[*User]) *Module {
	var h Hooks[*User]

	for _, hook := range hooks {
		h.BeforeCreate = append(h.BeforeCreate, hook.BeforeCreate...)
		h.AfterCreate = append(h.AfterCreate, hook.AfterCreate...)
	}

	service := newService(db, func() *User {
		return &User{}
	}, h)

	return &Module{
		service: service,
	}
}

func (m *Module) Start(srv *server.Service) {
	handler := newHandler()
	registerRoutes(srv, handler)
}

func (m *Module) Service() *Service[*User] {
	return m.service
}
