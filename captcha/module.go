package captcha

import (
	"github.com/sendelius/go/redis"
	"github.com/sendelius/go/server"
)

type Module struct {
	service *Service
}

func New(rds *redis.Service) *Module {
	service := newService(rds)
	return &Module{
		service: service,
	}
}

func (m *Module) Start(srv *server.Service) {
	handler := newHandler(m.service)
	registerRoutes(srv, handler)
}

func (m *Module) Service() *Service {
	return m.service
}
