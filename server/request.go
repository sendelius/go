package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sendelius/go/apperrors"
)

type contextKey struct{}

type Request struct {
}

type QuerySetter interface {
	SetQueryParam(key string, value any)
}

func (h *Request) Validate(w http.ResponseWriter, r *http.Request, dst any) bool {
	var err error

	if r.Method == http.MethodGet {
		params := extractDotQuery(r)
		query := r.URL.Query()

		for key := range query {
			parts := strings.SplitN(key, ".", 2)

			if len(parts) == 2 {
				query.Del(key)
			}
		}

		err = queryDecoder.Decode(dst, query)
		if err == nil {
			setQueryParams(dst, params)
		}
	} else {
		err = json.NewDecoder(r.Body).Decode(dst)
	}

	if err != nil {
		var ute *json.UnmarshalTypeError
		var se *json.SyntaxError
		switch {
		case errors.As(err, &ute):
			response.Error(w, r, apperrors.New(ErrorRequestTypeInvalid, fmt.Sprintf("поле '%s' имеет неверный тип, ожидается %s", ute.Field, humanType(ute.Type))).Cause(err).Build(), http.StatusBadRequest)
		case errors.As(err, &se):
			response.Error(w, r, apperrors.New(ErrorRequestJsonInvalid, "некорректный JSON в теле запроса").Cause(err).Build(), http.StatusBadRequest)
		case errors.Is(err, io.EOF):
			response.Error(w, r, apperrors.New(ErrorRequestEmptyBody, "тело запроса пустое").Cause(err).Build(), http.StatusBadRequest)
		default:
			response.Error(w, r, apperrors.New(ErrorRequestBodyInvalid, "не удалось обработать тело запроса").Cause(err).Build(), http.StatusBadRequest)
		}
		return false
	}

	if err = validate.Struct(dst); err != nil {
		response.Error(w, r, apperrors.New(ErrorRequestValidation, formatValidationError(err).Error()).Cause(err).Build(), http.StatusBadRequest)
		return false
	}

	return true
}

func (h *Request) GetParam(r *http.Request, key string) string {
	params, _ := r.Context().Value(contextKey{}).(map[string]string)
	return params[key]
}
