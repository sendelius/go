package captcha

import (
	"net/http"

	"github.com/sendelius/go/infrastructure"
)

type Handler struct {
	infrastructure.Handler
	captcha *Service
}

func newHandler(captcha *Service) *Handler {
	return &Handler{
		captcha: captcha,
	}
}

func (h *Handler) Challenge(w http.ResponseWriter, r *http.Request) {
	result, err := h.captcha.Create()
	if err != nil {
		h.Response.Error(w, r, err)
		return
	}
	h.Response.Success(w, r, result)
}

func (h *Handler) Verify(w http.ResponseWriter, r *http.Request) {
	request := new(struct {
		ID    string `json:"id" validate:"required"`
		Nonce uint64 `json:"nonce" validate:"required"`
	})

	if !h.Request.Validate(w, r, request) {
		return
	}

	token, err := h.captcha.Verify(request.ID, request.Nonce)
	if err != nil {
		h.Response.Error(w, r, err)
		return
	}

	h.Response.Success(w, r, map[string]any{
		"token": token,
	})
}
