package auth

import (
	"github.com/sendelius/go/captcha"
	"github.com/sendelius/go/server"
	"github.com/sendelius/go/sessions"
	"github.com/sendelius/go/users"
)

type Module[
	TUser users.Accessor,
	TSession sessions.Accessor,
] struct {
	service *Service[TUser, TSession]
}

func New(
	usersService *users.Service[*users.User],
	sessionsService *sessions.Service[*sessions.Session],
) *Module[*users.User, *sessions.Session] {
	return NewCustom[*users.User, *sessions.Session](usersService, sessionsService)
}

func NewCustom[
	TUser users.Accessor,
	TSession sessions.Accessor,
](
	usersService users.ServiceInterface[TUser],
	sessionsService sessions.ServiceInterface[TSession],
) *Module[TUser, TSession] {
	return &Module[TUser, TSession]{
		service: newService(usersService, sessionsService),
	}
}

func (m *Module[TUser, TSession]) Start(srv *server.Service, captcha *captcha.Service) {
	handler := newHandler(m.service, captcha)
	registerRoutes(srv, handler)
}

func (m *Module[TUser, TSession]) Service() *Service[TUser, TSession] {
	return m.service
}
