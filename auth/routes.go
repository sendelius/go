package auth

import "github.com/sendelius/go/server"

func registerRoutes(srv *server.Service, handler *Handler) {
	srv.PublicRoute("POST /v1/auth/login", handler.Login)
	srv.PublicRoute("POST /v1/auth/registration", handler.Register)
	srv.PublicRoute("POST /v1/auth/forgot", handler.Forgot)
	srv.ProtectedRoute("DELETE /v1/auth/logout", handler.Logout)
	srv.ProtectedRoute("GET /v1/auth/me", handler.Me)
}
