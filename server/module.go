package server

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"

	"github.com/sendelius/go/apperrors"
)

var response Response

type Module struct {
	host         string
	mux          *http.ServeMux
	service      *Service
	checkRequest func(http.ResponseWriter, *http.Request) (context.Context, error)
}

func New(
	host string,
	checkRequest func(http.ResponseWriter, *http.Request) (context.Context, error),
) *Module {
	mux := http.NewServeMux()
	service := newService(mux, checkRequest)
	return &Module{
		host:         host,
		mux:          mux,
		service:      service,
		checkRequest: checkRequest,
	}
}

func (m *Module) Start() {
	// 404 ошибка
	m.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, r, apperrors.New(ErrorNotFound, "метод не найден").Build(), http.StatusNotFound)
	})

	h := &http.Server{
		Addr:    m.host,
		Handler: m.mux,
	}

	ln, err := net.Listen("tcp", h.Addr)
	if err != nil {
		panic(err)
	}

	log.Printf("Сервер запущен на %s", h.Addr)

	if err := h.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}

func (m *Module) Service() *Service {
	return m.service
}
