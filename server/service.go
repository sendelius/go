package server

import (
	"context"
	"net/http"

	"github.com/sendelius/go/apperrors"
)

type Service struct {
	mux          *http.ServeMux
	checkRequest func(http.ResponseWriter, *http.Request) (context.Context, error)
}

func newService(
	mux *http.ServeMux,
	checkRequest func(http.ResponseWriter, *http.Request) (context.Context, error),
) *Service {
	return &Service{
		mux:          mux,
		checkRequest: checkRequest,
	}
}

func (s *Service) PublicRoute(pattern string, handler http.HandlerFunc) {
	s.mux.HandleFunc(pattern, withParams(pattern, func(w http.ResponseWriter, req *http.Request) {
		if s.checkRequest != nil {
			ctx, err := s.checkRequest(w, req)
			if err == nil && ctx != nil {
				req = req.WithContext(ctx)
			}
		}
		handler.ServeHTTP(w, req)
	}))
}

func (s *Service) ProtectedRoute(pattern string, handler http.HandlerFunc) {
	s.mux.HandleFunc(pattern, withParams(pattern, func(w http.ResponseWriter, req *http.Request) {
		if s.checkRequest == nil {
			response.Error(w, req, apperrors.New(ErrorSessionCheckerInvalid, "проверка сессии не настроена").Build(), http.StatusInternalServerError)
			return
		}
		ctx, err := s.checkRequest(w, req)
		if err != nil {
			response.Error(w, req, err, http.StatusUnauthorized)
			return
		}
		if ctx == nil {
			response.Error(w, req, apperrors.New(ErrorSessionInvalid, "требуется авторизация").Build(), http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(w, req.WithContext(ctx))
	}))
}
