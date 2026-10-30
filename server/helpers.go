package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/sendelius/go/apperrors"
)

func withParams(pattern string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		method, path := splitPattern(pattern)
		if req.Method != method {
			response.Error(w, req, apperrors.New(ErrorMethodNotAllowed, "метод не доступен").Build(), http.StatusMethodNotAllowed)
			return
		}

		params := parsePattern(path, req.URL.Path)
		if params == nil {
			response.Error(w, req, apperrors.New(ErrorNotFound, "метод не найден").Build(), http.StatusNotFound)
			return
		}

		ctx := context.WithValue(req.Context(), contextKey{}, params)
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

func prepareError(err any) *apperrors.Error {
	var appErr *apperrors.Error
	switch e := err.(type) {
	case string:
		appErr = apperrors.New(ErrorUndefined, e).Build()
	case error:
		if errors.As(e, &appErr) {
			return appErr
		}
		appErr = apperrors.New(ErrorUndefined, "внутренняя ошибка сервера").Cause(e).Build()
	default:
		appErr = apperrors.New(ErrorUndefined, "неизвестная ошибка").Build()
	}
	return appErr
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
		msg += fmt.Sprintf(`ошибка поля '%s': %s`, strings.ToLower(fe.Field()), translateError(fe))
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
