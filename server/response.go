package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/iancoleman/orderedmap"
	"github.com/sendelius/go/apperrors"
	"github.com/sendelius/go/env"
)

type Response struct {
}

func (h *Response) Success(w http.ResponseWriter, r *http.Request, data any, status ...int) {
	h.writeResponse(w, r, data, status...)
}

func (h *Response) Error(w http.ResponseWriter, r *http.Request, err error, status ...int) {
	code := http.StatusInternalServerError

	res := map[string]any{
		"status": "error",
		"error":  prepareError(err),
	}

	if appErr, ok := errors.AsType[*apperrors.Error](err); ok && appErr != nil {
		if appErr.Status != 0 {
			code = appErr.Status
		}

		for key, value := range appErr.Data {
			if key != "status" && key != "error" {
				res[key] = value
			}
		}
	}

	if len(status) > 0 {
		code = status[0]
	}

	h.writeResponse(w, r, res, code)
}

func (h *Response) Sanitize(v any, hidden ...string) any {
	b, _ := json.Marshal(v)

	var data any
	_ = json.Unmarshal(b, &data)
	remove := func(m map[string]any) {
		for _, field := range hidden {
			delete(m, field)
		}
	}

	switch value := v.(type) {
	case *orderedmap.OrderedMap:
		for _, field := range hidden {
			value.Delete(field)
		}
		return value
	case []any:
		// Обработка ниже через JSON
	case nil:
		return nil
	}

	switch value := data.(type) {
	case map[string]any:
		remove(value)
	case []any:
		for _, item := range value {
			if m, ok := item.(map[string]any); ok {
				remove(m)
			}
		}
	}
	return data
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
