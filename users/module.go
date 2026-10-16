package users

import (
	"github.com/sendelius/go/database"
	"github.com/sendelius/go/server"
)

type Module struct {
	service *Service[*User]
}

func New(srv *server.Service, db *database.Service) *Module {
	service := newService(db, func() *User {
		return &User{}
	})
	handler := newHandler()

	registerRoutes(srv, handler)

	return &Module{
		service: service,
	}
}

func (m *Module) Service() *Service[*User] {
	return m.service
}
