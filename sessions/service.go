package sessions

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/sendelius/go/apperrors"
	"github.com/sendelius/go/database"
	"github.com/sendelius/go/env"
	"github.com/sendelius/go/infrastructure"
	"gorm.io/gorm"
)

type Accessor interface {
	GetModel() *Session
}

type ServiceInterface[T Accessor] interface {
	NewModel() T
	Check(w http.ResponseWriter, r *http.Request) (T, error)
	GetByToken(token string) (T, error)
	Create(w http.ResponseWriter, userID string) (T, error)
	CreateBearer(userID string) (string, error)
	RemoveCurrent(w http.ResponseWriter, r *http.Request) error
	RemoveExpired() error
	RemoveByToken(token string)
}

type Service[T Accessor] struct {
	*infrastructure.Service[T]
	sessionTTL time.Duration
	factory    func() T
}

func newService[T Accessor](db *database.Service, factory func() T) *Service[T] {
	return &Service[T]{
		Service:    infrastructure.NewService[T](db),
		sessionTTL: time.Duration(env.Int("SESSION_TTL", 172800)) * time.Second,
		factory:    factory,
	}
}

func (s *Service[T]) NewModel() T {
	return s.factory()
}

func (s *Service[T]) Check(w http.ResponseWriter, r *http.Request) (T, error) {
	var zero T
	var token string
	var err error

	if bearer, err := getBearerToken(r); err == nil {
		token = strings.TrimSpace(bearer)
	} else if cookie, err := getCookieToken(r); err == nil {
		token = strings.TrimSpace(cookie.Value)
	}
	if token == "" {
		return zero, apperrors.New(ErrorTokenNotFound, "токен не найден").
			Status(http.StatusUnauthorized).
			Cause(err).
			Build()
	}

	sessionItem, err := s.GetByToken(token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return zero, apperrors.New(ErrorSessionNotFound, "сессия не найдена").
				Status(http.StatusUnauthorized).
				Cause(err).
				Build()
		}
		return zero, apperrors.New(ErrorSessionNotFound, "не удалось проверить сессию").
			Status(http.StatusUnauthorized).
			Cause(err).
			Build()
	}

	session := sessionItem.GetModel()
	now := time.Now()
	if !session.ExpiresAt.IsZero() && session.ExpiresAt.Before(now) {
		s.RemoveByToken(token)
		return zero, apperrors.New(ErrorSessionExpired, "сессия истекла").
			Status(http.StatusUnauthorized).
			Build()
	}

	if !session.ExpiresAt.IsZero() && session.ExpiresAt.Sub(now) < s.sessionTTL-time.Hour {
		expiresAt := now.Add(s.sessionTTL)
		session.ExpiresAt = expiresAt
		err = s.Db().Model(sessionItem).Where("id = ?", session.ID).Update("expires_at", expiresAt).Error
		if err != nil {
			return zero, apperrors.New(ErrorSessionExtend, "не удалось продлить сессию").
				Status(http.StatusUnauthorized).
				Cause(err).
				Build()
		}
		setCookieToken(w, token, expiresAt)
	}

	return sessionItem, nil
}

func (s *Service[T]) GetByToken(token string) (T, error) {
	item := s.factory()
	err := s.Db().Preload("User").First(item, "token = ?", hashToken(token)).Error
	return item, s.Error(err)
}

func (s *Service[T]) Create(w http.ResponseWriter, userID string) (T, error) {
	var zero T

	token, err := generateToken(64)
	if err != nil {
		return zero, apperrors.New(ErrorCreate, "ошибка генерации токена").
			Field("login").
			Status(http.StatusUnauthorized).
			Cause(err).
			Build()
	}

	expiresAt := time.Now().Add(s.sessionTTL)
	sessionItem := s.factory()
	session := sessionItem.GetModel()
	session.UserID = userID
	session.Token = hashToken(token)
	session.ExpiresAt = expiresAt

	if err = s.Db().Create(sessionItem).Error; err != nil {
		return zero, apperrors.New(ErrorCreateComplete, "ошибка сохранения сессии").
			Field("login").
			Status(http.StatusUnauthorized).
			Cause(err).
			Build()
	}

	setCookieToken(w, token, expiresAt)
	session.Token = token
	return sessionItem, nil
}

func (s *Service[T]) CreateBearer(userID string) (string, error) {
	token, err := generateToken(64)
	if err != nil {
		return "", apperrors.New(ErrorBearerCreate, "не удалось создать токен").Cause(err).Build()
	}

	hash := hashToken(token)
	sessionItem := s.factory()
	err = s.Db().Where("user_id = ? AND expires_at = ?", userID, time.Time{}).First(sessionItem).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		session := sessionItem.GetModel()
		session.UserID = userID
		session.Token = hash
		if err = s.Db().Create(sessionItem).Error; err != nil {
			return "", apperrors.New(ErrorBearerCreate, "не удалось создать токен").Cause(err).Build()
		}
		return token, nil
	}
	if err != nil {
		return "", apperrors.New(ErrorBearerCreate, "не удалось создать токен").Cause(err).Build()
	}

	session := sessionItem.GetModel()
	if err = s.Db().Model(sessionItem).Update("token", hash).Error; err != nil {
		return "", apperrors.New(ErrorBearerUpdate, "не удалось обновить токен").Cause(err).Build()
	}

	session.Token = hash
	return token, nil
}

func (s *Service[T]) RemoveCurrent(w http.ResponseWriter, r *http.Request) error {
	cookie, err := getCookieToken(r)
	if err != nil {
		return apperrors.New(ErrorLogoutNotFound, "сессия не найдена").Cause(err).Build()
	}

	err = s.Db().Where("token = ?", hashToken(cookie.Value)).Delete(s.factory()).Error
	if err != nil {
		return apperrors.New(ErrorLogoutComplete, "не удалось завершить сессию").Cause(err).Build()
	}

	deleteCookieToken(w)
	return nil
}

func (s *Service[T]) RemoveExpired() error {
	err := s.Db().Where("expires_at > ? AND expires_at < ?", time.Time{}, time.Now()).Delete(s.factory()).Error
	return s.Error(err)
}

func (s *Service[T]) RemoveByToken(token string) {
	s.Db().Where("token = ?", hashToken(token)).Delete(s.factory())
}
