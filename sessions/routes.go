package sessions

import (
	"github.com/sendelius/go/server"
)

func registerRoutes[
	T Accessor,
](
	srv *server.Service,
	handler *Handler[T],
) {
	srv.ProtectedRoute("POST /v1/sessions/bearer", handler.CreateBearer)
	srv.ProtectedRoute("PUT /v1/sessions/bearer", handler.UpdateBearer)
	srv.ProtectedRoute("GET /v1/sessions", handler.GetList)
	srv.ProtectedRoute("DELETE /v1/sessions/{id}", handler.Delete)
}
