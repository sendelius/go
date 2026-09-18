package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Request struct {
}

type QuerySetter interface {
	SetQueryParam(key string, value any)
}

func (h *Request) GetSession(r *http.Request) any {
	return r.Context().Value(SessionKey)
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
			writeResponse(
				w,
				r,
				map[string]string{
					"status": "error",
					"error": fmt.Sprintf(
						"поле '%s' имеет неверный тип, ожидается %s",
						ute.Field,
						humanType(ute.Type),
					),
				},
				http.StatusBadRequest,
			)

		case errors.As(err, &se):
			writeResponse(
				w,
				r,
				map[string]string{
					"status": "error",
					"error":  "некорректный JSON в теле запроса",
				},
				http.StatusBadRequest,
			)

		case errors.Is(err, io.EOF):
			writeResponse(
				w,
				r,
				map[string]string{
					"status": "error",
					"error":  "тело запроса пустое",
				},
				http.StatusBadRequest,
			)

		default:
			writeResponse(
				w,
				r,
				map[string]string{
					"status": "error",
					"error":  "не удалось обработать тело запроса: " + err.Error(),
				},
				http.StatusBadRequest,
			)
		}

		return false
	}

	if err := validate.Struct(dst); err != nil {
		writeResponse(
			w,
			r,
			map[string]string{
				"status": "error",
				"error":  formatValidationError(err).Error(),
			},
			http.StatusBadRequest,
		)
		return false
	}

	return true
}

func (h *Request) GetParam(r *http.Request, key string) string {
	params, _ := r.Context().Value(ParamsKey).(map[string]string)
	return params[key]
}

func humanType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "строка"
	case reflect.Bool:
		return "булево значение"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "число"
	case reflect.Float32, reflect.Float64:
		return "число"
	case reflect.Slice, reflect.Array:
		return "массив"
	case reflect.Map:
		return "объект"
	case reflect.Struct:
		return "объект"
	case reflect.Pointer:
		return humanType(t.Elem())
	default:
		return "значение"
	}
}

func formatValidationError(err error) error {
	var ve validator.ValidationErrors

	if !errors.As(err, &ve) {
		return fmt.Errorf("ошибка валидации: %w", err)
	}

	var msg string

	for i, fe := range ve {
		if i > 0 {
			msg += "; "
		}

		msg += fmt.Sprintf(
			`ошибка поля '%s': %s`,
			strings.ToLower(fe.Field()),
			translateError(fe),
		)
	}

	return errors.New(msg)
}

func translateError(fe validator.FieldError) string {
	switch fe.Tag() {

	case "required":
		return "поле обязательно"

	case "min":
		return "слишком короткое значение"

	case "max":
		return "слишком длинное значение"

	case "oneof":
		return "недопустимое значение"

	case "email":
		return "некорректный email"

	case "omitempty":
		return ""

	default:
		return "некорректное значение"
	}
}

func extractDotQuery(r *http.Request) map[string]any {
	query := r.URL.Query()
	result := make(map[string]any)

	for key, values := range query {
		if len(values) == 0 {
			continue
		}

		parts := strings.SplitN(key, ".", 2)

		if len(parts) != 2 {
			continue
		}

		group := parts[0]
		field := parts[1]

		data, ok := result[group].(map[string]string)
		if !ok {
			data = make(map[string]string)
			result[group] = data
		}

		data[field] = values[0]
	}

	return result
}

func setQueryParams(dst any, params map[string]any) {
	setter, ok := dst.(QuerySetter)
	if !ok {
		return
	}

	for key, value := range params {
		setter.SetQueryParam(key, value)
	}
}
