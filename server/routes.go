package server

import (
	"context"
	"net/http"
	"strings"
)

func (s *Server) PublicRoute(pattern string, handler http.HandlerFunc) {
	s.mux.HandleFunc(pattern, withParams(pattern, handler))
}

func (s *Server) ProtectedRoute(pattern string, handler http.HandlerFunc) {
	s.mux.HandleFunc(pattern, withParams(pattern, func(w http.ResponseWriter, req *http.Request) {
		session, err := s.sessions.Check(w, req)
		if err != nil {
			writeResponse(
				w,
				req,
				map[string]string{
					"status": "error",
					"error":  err.Error(),
				},
				http.StatusUnauthorized,
			)
			return
		}

		ctx := context.WithValue(req.Context(), SessionKey, session)

		handler.ServeHTTP(w, req.WithContext(ctx))
	}))
}

func withParams(pattern string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		method, path := splitPattern(pattern)

		// проверка метода
		if req.Method != method {
			writeResponse(
				w,
				req,
				map[string]string{
					"status": "error",
					"error":  "метод не доступен",
				},
				http.StatusMethodNotAllowed,
			)
			return
		}

		params := parsePattern(path, req.URL.Path)

		if params == nil {
			writeResponse(
				w,
				req,
				map[string]string{
					"status": "error",
					"error":  "метод не найден",
				},
				http.StatusNotFound,
			)
			return
		}

		ctx := context.WithValue(req.Context(), ParamsKey, params)
		next.ServeHTTP(w, req.WithContext(ctx))
	}
}

func splitPattern(pattern string) (method, path string) {
	parts := strings.SplitN(pattern, " ", 2)
	if len(parts) != 2 {
		return "", pattern
	}
	return parts[0], parts[1]
}

func parsePattern(pattern, path string) map[string]string {
	clean := func(s string) []string {
		s = strings.Trim(s, "/")
		if s == "" {
			return []string{}
		}
		return strings.Split(s, "/")
	}

	pParts := clean(pattern)
	uParts := clean(path)

	if len(pParts) != len(uParts) {
		return nil
	}

	params := make(map[string]string)

	for i := range pParts {
		p := pParts[i]
		u := uParts[i]

		if len(p) > 2 && p[0] == '{' && p[len(p)-1] == '}' {
			params[p[1:len(p)-1]] = u
			continue
		}

		if p != u {
			return nil
		}
	}

	return params
}
