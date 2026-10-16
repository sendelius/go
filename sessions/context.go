package sessions

import (
	"context"

	"github.com/sendelius/go/users"
)

type contextKey struct{}

func FromContext(ctx context.Context) (Accessor, bool) {
	current, ok := ctx.Value(contextKey{}).(Accessor)
	return current, ok
}

func UserFromContext(ctx context.Context) (*users.User, bool) {
	current, ok := FromContext(ctx)
	if !ok || current.GetModel() == nil {
		return nil, false
	}
	return &current.GetModel().User, true
}
