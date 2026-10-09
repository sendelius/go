package session

import (
	"context"
	"net/http"
	"reflect"

	"github.com/sendelius/go/users"
)

type contextKey struct{}

type RequestChecker interface {
	CheckRequest(http.ResponseWriter, *http.Request) (context.Context, error)
}

type requestChecker[T Accessor] struct {
	service *Service[T]
}

func NewRequestChecker[T Accessor](service *Service[T]) RequestChecker {
	return &requestChecker[T]{service: service}
}

func (c *requestChecker[T]) CheckRequest(w http.ResponseWriter, r *http.Request) (context.Context, error) {
	current, err := c.service.Check(w, r)
	if err != nil {
		return nil, err
	}
	if isNilSession(current) {
		return r.Context(), nil
	}
	return context.WithValue(r.Context(), contextKey{}, current), nil
}

func isNilSession(value any) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface,
		reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

func FromContext(ctx context.Context) (Accessor, bool) {
	current, ok := ctx.Value(contextKey{}).(Accessor)
	return current, ok
}

func UserFromContext(ctx context.Context) (*users.User, bool) {
	current, ok := FromContext(ctx)
	if !ok || current == nil || current.GetModel() == nil {
		return nil, false
	}
	return &current.GetModel().User, true
}
