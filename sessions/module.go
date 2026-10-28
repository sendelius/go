package sessions

import (
	"context"
	"net/http"

	"github.com/sendelius/go/database"
	"github.com/sendelius/go/server"
)

type Module[T Accessor] struct {
	service *Service[T]
	worker  *Worker[T]
}

func New(db *database.Service) *Module[*Session] {
	return NewCustom[*Session](db, func() *Session {
		return &Session{}
	})
}

func NewCustom[T Accessor](db *database.Service, factory func() T) *Module[T] {
	service := newService(db, factory)
	worker := newWorker(service)
	return &Module[T]{
		service: service,
		worker:  worker,
	}
}

func (m *Module[T]) Start(srv *server.Service) {
	handler := newHandler()
	registerRoutes(srv, handler)
	go m.worker.Clean()
}

func (m *Module[T]) Service() *Service[T] {
	return m.service
}

func (m *Module[T]) CheckRequest(w http.ResponseWriter, r *http.Request) (context.Context, error) {
	current, err := m.service.Check(w, r)
	if err != nil {
		return nil, err
	}
	ctx := context.WithValue(r.Context(), contextKey{}, current)
	return ctx, nil
}
