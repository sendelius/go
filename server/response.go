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
	writeResponse(w, r, data, status...)
}

func (h *Response) Error(w http.ResponseWriter, r *http.Request, err any, status ...int) {
	code := http.StatusInternalServerError
	if len(status) > 0 {
		code = status[0]
	}

	writeResponse(w, r, map[string]any{
		"status": "error",
		"error":  prepareError(err),
	}, code)
}

func (h *Response) ErrorWithField(w http.ResponseWriter, r *http.Request, err any, field string, status ...int) {
	code := http.StatusInternalServerError
	if len(status) > 0 {
		code = status[0]
	}

	writeResponse(w, r, map[string]any{
		"status": "error",
		"error":  prepareError(err),
		"field":  field,
	}, code)
}

func (h *Response) ErrorWithData(w http.ResponseWriter, r *http.Request, err any, data map[string]any, status ...int) {
	code := http.StatusInternalServerError
	if len(status) > 0 {
		code = status[0]
	}

	response := map[string]any{
		"status": "error",
		"error":  prepareError(err),
	}
	for key, value := range data {
		if key != "status" && key != "error" {
			response[key] = value
		}
	}

	writeResponse(w, r, response, code)
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

func prepareError(err any) *apperrors.Error {
	var appErr *apperrors.Error

	switch e := err.(type) {
	case string:
		appErr = apperrors.New(ErrorUndefined, e)
	case error:
		if !errors.As(e, &appErr) {
			appErr = apperrors.New(ErrorUndefined, e.Error())
		}
	default:
		appErr = apperrors.New(ErrorUndefined, "неизвестная ошибка")
	}

	return appErr
}

func writeResponse(w http.ResponseWriter, r *http.Request, data any, status ...int) {
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
