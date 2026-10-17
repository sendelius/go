package auth

import (
	"github.com/sendelius/go/captcha"
	"github.com/sendelius/go/server"
	"github.com/sendelius/go/sessions"
	"github.com/sendelius/go/users"
)

type Module struct {
	service *Service
}

func New(
	users *users.Service[*users.User],
	sessions *sessions.Service[*sessions.Session],
) *Module {
	service := newService(users, sessions)
	return &Module{
		service: service,
	}
}

func (m *Module) Start(srv *server.Service, captcha *captcha.Service) {
	handler := newHandler(m.service, captcha)
	registerRoutes(srv, handler)
}

func (m *Module) Service() *Service {
	return m.service
}
