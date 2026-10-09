package server

import (
	"errors"
	"log"
	"net"
	"net/http"

	"github.com/sendelius/go/session"
	"gorm.io/gorm"
)

type Server struct {
	host     string
	mux      *http.ServeMux
	DB       *gorm.DB
	sessions session.RequestChecker
}

func New(host string, sessions ...session.RequestChecker) *Server {
	server := &Server{
		host: host,
		mux:  http.NewServeMux(),
	}
	if len(sessions) > 0 {
		server.sessions = sessions[0]
	}
	return server
}

func (s *Server) Start() {
	// 404 ошибка
	s.mux.HandleFunc("/", error404)

	h := &http.Server{
		Addr:    s.host,
		Handler: s.mux,
	}

	ln, err := net.Listen("tcp", h.Addr)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Сервер запущен на %s", h.Addr)

	if err := h.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func error404(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, r, map[string]string{
		"status": "error",
		"error":  "метод не найден",
	}, http.StatusNotFound)
}
