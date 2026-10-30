package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/schema"
	"github.com/iancoleman/orderedmap"
	"github.com/sendelius/go/apperrors"
	"github.com/sendelius/go/env"
)

var validate = validator.New()
var queryDecoder = schema.NewDecoder()

type Response struct {
}

func (h *Response) Success(w http.ResponseWriter, r *http.Request, data any, status ...int) {
	h.writeResponse(w, r, data, status...)
}

func (h *Response) Error(w http.ResponseWriter, r *http.Request, err error, status ...int) {
	code := http.StatusInternalServerError

	if appErr, ok := errors.AsType[*apperrors.Error](err); ok &&
		appErr != nil && appErr.Status != 0 {
		code = appErr.Status
	}

	if len(status) > 0 {
		code = status[0]
	}

	h.writeResponse(w, r, map[string]any{
		"status": "error",
		"error":  prepareError(err),
	}, code)
}

func (h *Response) Sanitize(v any, hidden ...string) *orderedmap.OrderedMap {
	b, _ := json.Marshal(v)

	m := orderedmap.New()
	_ = json.Unmarshal(b, &m)

	for _, field := range hidden {
		m.Delete(field)
	}

	return m
}

func (h *Response) writeResponse(w http.ResponseWriter, r *http.Request, data any, status ...int) {
	w.Header().Set("Content-Type", "application/json")

	w.Header().Set("Access-Control-Allow-Origin", "https://"+env.String("BASE_DOMAIN"))
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	code := http.StatusOK
	if len(status) > 0 {
		code = status[0]
	}

	if data == nil {
		data = map[string]string{}
	}

	switch v := data.(type) {
	case map[string]any:
		if _, ok := v["status"]; !ok {
			v["status"] = "success"
		}
	case map[string]string:
		if _, ok := v["status"]; !ok {
			v["status"] = "success"
		}
	}

	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(data)
}
