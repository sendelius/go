package auth

import (
	"net/http"

	"github.com/sendelius/go/apperrors"
	"github.com/sendelius/go/sessions"
	"github.com/sendelius/go/users"
)

type Service[
	TUser users.Accessor,
	TSession sessions.Accessor,
] struct {
	users    users.ServiceInterface[TUser]
	sessions sessions.ServiceInterface[TSession]
}

func newService[
	TUser users.Accessor,
	TSession sessions.Accessor,
](
	users users.ServiceInterface[TUser],
	sessions sessions.ServiceInterface[TSession],
) *Service[TUser, TSession] {
	return &Service[TUser, TSession]{
		users:    users,
		sessions: sessions,
	}
}

func (s *Service[TUser, TSession]) NewUserModel() TUser {
	return s.users.NewModel()
}

func (s *Service[TUser, TSession]) NewSessionModel() TSession {
	return s.sessions.NewModel()
}

func (s *Service[TUser, TSession]) Login(
	w http.ResponseWriter,
	r *http.Request,
	login string,
	password string,
) (TUser, TSession, string, int, error) {
	var zeroUser TUser
	var zeroSession TSession

	_, err := s.sessions.Check(w, r)
	if err == nil {
		return zeroUser, zeroSession, "login", http.StatusBadRequest, apperrors.New(ErrorLoginAlready, "вы уже авторизованы")
	}

	user, field, err := s.users.Check(login, password)
	if err != nil {
		return zeroUser, zeroSession, field, http.StatusUnauthorized, err
	}

	sessionItem, err := s.sessions.Login(w, user.GetModel().ID)
	if err != nil {
		return zeroUser, zeroSession, "login", http.StatusUnauthorized, err
	}

	return user, sessionItem, "", http.StatusOK, nil
}

func (s *Service[TUser, TSession]) Register(user TUser) (string, error) {
	_, field, err := s.users.Create(user)
	if err != nil {
		return field, err
	}

	// TODO: отправка письма о регистрации
	return "", nil
}

func (s *Service[TUser, TSession]) Forgot() error {
	// TODO: отправка письма восстановления
	return nil
}

func (s *Service[TUser, TSession]) Logout(w http.ResponseWriter, r *http.Request) error {
	return s.sessions.Logout(w, r)
}

func (s *Service[TUser, TSession]) Me(r *http.Request) (TUser, error) {
	var zero TUser

	current, ok := sessions.FromContext(r.Context())
	if !ok {
		return zero, apperrors.New(ErrorSessionNotFound, "сессия не найдена")
	}

	currentSession := current.GetModel()
	user, err := s.users.GetByID(currentSession.UserID)
	if err != nil {
		return zero, apperrors.New(ErrorUserNotFound, "пользователь не найден", err)
	}

	return user, nil
}
