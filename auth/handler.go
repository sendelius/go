package auth

import (
	"net/http"

	"github.com/sendelius/go/captcha"
	"github.com/sendelius/go/infrastructure"
	"github.com/sendelius/go/users"
)

type Handler struct {
	infrastructure.Handler
	auth    *Service
	captcha *captcha.Service
}

func newHandler(
	auth *Service,
	captcha *captcha.Service,
) *Handler {
	return &Handler{
		auth:    auth,
		captcha: captcha,
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
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

	_, session, field, status, err := h.auth.Login(w, r, request.Login, request.Password)
	if err != nil {
		h.Response.ErrorWithField(w, r, err, field, status)
		return
	}

	h.Response.Success(w, r, map[string]string{
		"token": session.Token,
	})
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
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

	field, err := h.auth.Register(&users.User{
		Name:  request.Name,
		Email: request.Email,
		Phone: request.Phone,
	})
	if err != nil {
		h.Response.ErrorWithField(w, r, err, field)
		return
	}

	h.Response.Success(w, r, nil, http.StatusCreated)
}

func (h *Handler) Forgot(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	err := h.auth.Logout(w, r)
	if err != nil {
		h.Response.Error(w, r, err)
		return
	}
	h.Response.Success(w, r, nil)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user, err := h.auth.Me(r)
	if err != nil {
		h.Response.Error(w, r, err, http.StatusNotFound)
		return
	}
	h.Response.Success(w, r, map[string]any{
		"user": user,
	})
}
