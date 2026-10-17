package users

import (
	"github.com/sendelius/go/database"
	"github.com/sendelius/go/server"
)

type Module struct {
	service *Service[*User]
}

func New(db *database.Service) *Module {
	service := newService(db, func() *User {
		return &User{}
	})
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
