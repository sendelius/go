package users

import (
	"github.com/sendelius/go/database"
	"github.com/sendelius/go/server"
)

type Module[T Accessor] struct {
	service *Service[T]
}

func New(db *database.Service) *Module[*User] {
	return NewCustom[*User](db, func() *User {
		return &User{}
	})
}

func NewCustom[T Accessor](db *database.Service, factory func() T) *Module[T] {
	return &Module[T]{
		service: newService(db, factory),
	}
}

func (m *Module[T]) Start(srv *server.Service) {
	handler := newHandler()
	registerRoutes(srv, handler)
}

func (m *Module[T]) Service() *Service[T] {
	return m.service
}
