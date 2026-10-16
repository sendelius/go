package sessions

import (
	"context"
	"net/http"

	"github.com/sendelius/go/database"
	"github.com/sendelius/go/server"
)

type Module struct {
	service *Service[*Session]
	worker  *Worker[*Session]
}

func New(db *database.Service) *Module {
	service := newService(db, func() *Session {
		return &Session{}
	})
	worker := newWorker(service)
	return &Module{
		service: service,
		worker:  worker,
	}
}

func (m *Module) Start(srv *server.Service) {
	handler := newHandler()
	registerRoutes(srv, handler)
	go m.worker.Clean()
}

func (m *Module) Service() *Service[*Session] {
	return m.service
}

func (m *Module) CheckRequest(w http.ResponseWriter, r *http.Request) (context.Context, error) {
	current, err := m.service.Check(w, r)
	if err != nil {
		return nil, err
	}

	ctx := r.Context()
	if current != nil {
		ctx = context.WithValue(ctx, contextKey{}, current)
	}

	return ctx, nil
}
