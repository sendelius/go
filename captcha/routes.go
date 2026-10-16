package captcha

import "github.com/sendelius/go/server"

func registerRoutes(srv *server.Service, handler *Handler) {
	srv.PublicRoute("GET /v1/captcha/challenge", handler.Challenge)
	srv.PublicRoute("POST /v1/captcha/verify", handler.Verify)
}
