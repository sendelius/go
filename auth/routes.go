package auth

import (
	"github.com/sendelius/go/server"
	"github.com/sendelius/go/sessions"
	"github.com/sendelius/go/users"
)

func registerRoutes[
	TUser users.Accessor,
	TSession sessions.Accessor,
](
	srv *server.Service,
	handler *Handler[TUser, TSession],
) {
	srv.PublicRoute("POST /v1/auth/login", handler.Login)
	srv.PublicRoute("POST /v1/auth/registration", handler.Register)
	srv.PublicRoute("POST /v1/auth/forgot", handler.Forgot)
	srv.ProtectedRoute("DELETE /v1/auth/logout", handler.Logout)
	srv.ProtectedRoute("GET /v1/auth/me", handler.Me)
}
