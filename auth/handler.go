package auth

import (
	"net/http"

	"github.com/sendelius/go/captcha"
	"github.com/sendelius/go/infrastructure"
	"github.com/sendelius/go/sessions"
	"github.com/sendelius/go/users"
)

type Handler[
	TUser users.Accessor,
	TSession sessions.Accessor,
] struct {
	infrastructure.Handler
	auth    *Service[TUser, TSession]
	captcha *captcha.Service
}

func newHandler[
	TUser users.Accessor,
	TSession sessions.Accessor,
](
	auth *Service[TUser, TSession],
	captcha *captcha.Service,
) *Handler[TUser, TSession] {
	return &Handler[TUser, TSession]{
		auth:    auth,
		captcha: captcha,
	}
}

func (h *Handler[TUser, TSession]) Login(w http.ResponseWriter, r *http.Request) {
	request := new(struct {
		Login    string `json:"login" validate:"required"`
		Password string `json:"password" validate:"required"`
		Captcha  string `json:"captcha"`
	})
	if !h.Request.Validate(w, r, request) {
		return
	}

	if err := h.captcha.Consume(request.Captcha); err != nil {
		h.Response.Error(w, r, err)
		return
	}

	_, session, err := h.auth.Login(w, r, request.Login, request.Password)
	if err != nil {
		h.Response.Error(w, r, err)
		return
	}

	h.Response.Success(w, r, map[string]string{
		"token": session.GetModel().Token,
	})
}

func (h *Handler[TUser, TSession]) Register(w http.ResponseWriter, r *http.Request) {
	request := new(struct {
		Name    string `json:"name" validate:"required"`
		Email   string `json:"email" validate:"required"`
		Phone   string `json:"phone"`
		Captcha string `json:"captcha"`
	})
	if !h.Request.Validate(w, r, request) {
		return
	}

	if err := h.captcha.Consume(request.Captcha); err != nil {
		h.Response.Error(w, r, err)
		return
	}

	user := h.auth.NewUserModel()
	model := user.GetModel()

	model.Name = request.Name
	model.Email = request.Email
	model.Phone = request.Phone

	err := h.auth.Register(user)
	if err != nil {
		h.Response.Error(w, r, err)
		return
	}

	h.Response.Success(w, r, nil, http.StatusCreated)
}

func (h *Handler[TUser, TSession]) Forgot(w http.ResponseWriter, r *http.Request) {
	request := new(struct {
		Email   string `json:"email" validate:"required"`
		Captcha string `json:"captcha"`
	})
	if !h.Request.Validate(w, r, request) {
		return
	}

	if err := h.captcha.Consume(request.Captcha); err != nil {
		h.Response.Error(w, r, err)
		return
	}

	if err := h.auth.Forgot(); err != nil {
		h.Response.Error(w, r, err)
		return
	}

	h.Response.Success(w, r, nil)
}

func (h *Handler[TUser, TSession]) Logout(w http.ResponseWriter, r *http.Request) {
	err := h.auth.Logout(w, r)
	if err != nil {
		h.Response.Error(w, r, err)
		return
	}
	h.Response.Success(w, r, nil)
}

func (h *Handler[TUser, TSession]) Me(w http.ResponseWriter, r *http.Request) {
	user, err := h.auth.Me(r)
	if err != nil {
		h.Response.Error(w, r, err, http.StatusNotFound)
		return
	}
	h.Response.Success(w, r, map[string]any{
		"user": user,
	})
}
