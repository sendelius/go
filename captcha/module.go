package captcha

import (
	"github.com/sendelius/go/redis"
	"github.com/sendelius/go/server"
)

type Module struct {
	service *Service
}

func New(srv *server.Service, rds *redis.Service) *Module {
	service := newService(rds)
	handler := newHandler(service)

	registerRoutes(srv, handler)

	return &Module{
		service: service,
	}
}

func (m *Module) Service() *Service {
	return m.service
}
