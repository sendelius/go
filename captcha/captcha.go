package captcha

import (
	"github.com/sendelius/go/redis"
	"github.com/sendelius/go/server"
)

func New(srv *server.Server, redisService *redis.Redis) *Service {
	service := newService(redisService)
	handler := newHandler(service)

	srv.PublicRoute("GET /v1/captcha/challenge", handler.Challenge)
	srv.PublicRoute("POST /v1/captcha/verify", handler.Verify)

	return service
}
