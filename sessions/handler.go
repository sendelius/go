package sessions

import (
	"net/http"

	"github.com/sendelius/go/infrastructure"
)

type Handler[
	T Accessor,
] struct {
	infrastructure.Handler
	sessions *Service[T]
}

func newHandler[
	T Accessor,
](
	sessions *Service[T],
) *Handler[T] {
	return &Handler[T]{
		sessions: sessions,
	}
}

func (h *Handler[T]) GetList(w http.ResponseWriter, r *http.Request) {
	request := struct {
		UserID string `json:"user_id" validate:"omitempty"`
		*infrastructure.ListRequest
	}{
		ListRequest: h.sessions.GetListRequest(),
	}
	if !h.Request.Validate(w, r, &request) {
		return
	}

	if request.UserID == "" {
		current, ok := FromContext(r.Context())
		if ok {
			request.UserID = current.GetModel().UserID
		}
	}

	request.SetFilter("user_id", request.UserID)

	sessions, total, pagination, err := h.sessions.GetList(request.ListRequest, &infrastructure.ListConfig{
		FilterFields: []string{"is_bearer", "user_id"},
		SortFields:   []string{"id", "created_at", "updated_at", "token", "comment", "expires_at"},
	})
	if err != nil {
		h.Response.Error(w, r, err)
		return
	}

	h.Response.Success(w, r, map[string]any{
		"sessions":   h.Response.Sanitize(sessions, "user_id", "user"),
		"total":      total,
		"pagination": pagination,
	})
}

func (h *Handler[T]) CreateBearer(w http.ResponseWriter, r *http.Request) {
	request := new(struct {
		UserID  string `json:"user_id" validate:"omitempty"`
		Comment string `json:"comment" validate:"omitempty"`
	})
	if !h.Request.Validate(w, r, request) {
		return
	}

	if request.UserID == "" {
		current, ok := FromContext(r.Context())
		if ok {
			request.UserID = current.GetModel().UserID
		}
	}

	session, err := h.sessions.CreateBearer(request.UserID, request.Comment)
	if err != nil {
		h.Response.Error(w, r, err)
		return
	}

	h.Response.Success(w, r, map[string]string{
		"token": session.GetModel().Token,
	})
}

func (h *Handler[T]) UpdateBearer(w http.ResponseWriter, r *http.Request) {
	request := new(struct {
		ID      string `json:"id" validate:"required"`
		Comment string `json:"comment" validate:"omitempty"`
	})
	if !h.Request.Validate(w, r, request) {
		return
	}

	session, err := h.sessions.UpdateBearer(request.ID, request.Comment)
	if err != nil {
		h.Response.Error(w, r, err)
		return
	}

	h.Response.Success(w, r, map[string]string{
		"token": session.GetModel().Token,
	})
}

func (h *Handler[T]) Delete(w http.ResponseWriter, r *http.Request) {
	id := h.Request.GetParam(r, "id")

	err := h.sessions.DeleteByID(id)
	if err != nil {
		h.Response.Error(w, r, err)
		return
	}

	h.Response.Success(w, r, nil)
}
