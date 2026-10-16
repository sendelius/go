package server

import (
	"errors"
	"log"
	"net"
	"net/http"

	"github.com/sendelius/go/apperrors"
	"github.com/sendelius/go/session"
)

type Service struct {
	host     string
	mux      *http.ServeMux
	sessions session.RequestChecker
}

func newService(host string) *Service {
	return &Service{
		host: host,
		mux:  http.NewServeMux(),
	}
}

func (s *Service) Start() {
	// 404 ошибка
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeResponse(w, r, map[string]any{
			"status": "error",
			"error":  apperrors.New(ErrorNotFound, "метод не найден"),
		}, http.StatusNotFound)
	})

	h := &http.Server{
		Addr:    s.host,
		Handler: s.mux,
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
