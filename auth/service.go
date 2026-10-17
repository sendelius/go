package auth

import (
	"net/http"

	"github.com/sendelius/go/apperrors"
	"github.com/sendelius/go/sessions"
	"github.com/sendelius/go/users"
)

type Service struct {
	users    *users.Service[*users.User]
	sessions *sessions.Service[*sessions.Session]
}

func newService(
	users *users.Service[*users.User],
	sessions *sessions.Service[*sessions.Session],
) *Service {
	return &Service{
		users:    users,
		sessions: sessions,
	}
}

func (s *Service) Login(
	w http.ResponseWriter,
	r *http.Request,
	login string,
	password string,
) (*users.User, *sessions.Session, string, int, error) {
	_, err := s.sessions.Check(w, r)
	if err == nil {
		return nil, nil, "login", http.StatusBadRequest, apperrors.New(ErrorLoginAlready, "вы уже авторизованы")
	}

	user, field, err := s.users.Check(login, password)
	if err != nil {
		return nil, nil, field, http.StatusUnauthorized, err
	}

	sessionItem, err := s.sessions.Login(w, user.ID)
	if err != nil {
		return nil, nil, "login", http.StatusUnauthorized, err
	}

	return user, sessionItem, "", http.StatusOK, nil
}

func (s *Service) Register(user *users.User) (string, error) {
	_, field, err := s.users.Create(user)
	if err != nil {
		return field, err
	}

	// TODO: отправка письма о регистрации
	return "", nil
}

func (s *Service) Forgot() error {
	// TODO: отправка письма восстановления
	return nil
}

func (s *Service) Logout(w http.ResponseWriter, r *http.Request) error {
	return s.sessions.Logout(w, r)
}

func (s *Service) Me(r *http.Request) (*users.User, error) {
	current, ok := sessions.FromContext(r.Context())
	if !ok {
		return nil, apperrors.New(ErrorSessionNotFound, "сессия не найдена")
	}

	currentSession := current.GetModel()
	user, err := s.users.GetByID(currentSession.UserID)
	if err != nil {
		return nil, apperrors.New(ErrorUserNotFound, "пользователь не найден", err)
	}

	return user, nil
}
