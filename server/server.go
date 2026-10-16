package server

import (
	"github.com/sendelius/go/session"
)

func New(host string, sessions ...session.RequestChecker) *Service {
	if len(sessions) > 1 {
		panic("server: допускается только одна проверка сессии")
	}

	service := newService(host)

	if len(sessions) == 1 {
		service.sessions = sessions[0]
	}

	return service
}
