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
		return zero, apperrors.New(ErrorTokenNotFound, "токен не найден", err)
	}

	sessionItem, err := s.GetByToken(token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return zero, apperrors.New(ErrorSessionNotFound, "сессия не найдена", err)
		}
		return zero, apperrors.New(ErrorSessionNotFound, "не удалось проверить сессию", err)
	}

	session := sessionItem.GetModel()
	now := time.Now()
	if !session.ExpiresAt.IsZero() && session.ExpiresAt.Before(now) {
		s.RemoveByToken(token)
		return zero, apperrors.New(ErrorSessionExpired, "сессия истекла")
	}

	if !session.ExpiresAt.IsZero() && session.ExpiresAt.Sub(now) < s.sessionTTL-time.Hour {
		expiresAt := now.Add(s.sessionTTL)
		session.ExpiresAt = expiresAt
		err = s.Db().Model(sessionItem).Where("id = ?", session.ID).Update("expires_at", expiresAt).Error
		if err != nil {
			return zero, apperrors.New(ErrorSessionExtend, "не удалось продлить сессию", err)
		}
		setCookieToken(w, token, expiresAt)
	}

	return sessionItem, nil
}

func (s *Service[T]) Login(w http.ResponseWriter, userID string) (T, error) {
	var zero T

	token, err := generateToken(32)
	if err != nil {
		return zero, apperrors.New(ErrorSessionCreate, "не удалось создать сессию", err)
	}

	expiresAt := time.Now().Add(s.sessionTTL)
	sessionItem := s.factory()
	session := sessionItem.GetModel()
	session.UserID = userID
	session.Token = hashToken(token)
	session.ExpiresAt = expiresAt
	if err := s.Db().Create(sessionItem).Error; err != nil {
		return zero, apperrors.New(ErrorSessionCreate, "не удалось создать сессию", err)
	}

	setCookieToken(w, token, expiresAt)

	return sessionItem, nil
}

func (s *Service[T]) Bearer(userID string) (string, error) {
	token, err := generateToken(64)
	if err != nil {
		return "", apperrors.New(ErrorBearerCreate, "не удалось создать токен", err)
	}

	hash := hashToken(token)
	sessionItem := s.factory()
	err = s.Db().Where("user_id = ? AND expires_at = ?", userID, time.Time{}).First(sessionItem).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		session := sessionItem.GetModel()
		session.UserID = userID
		session.Token = hash
		if err = s.Db().Create(sessionItem).Error; err != nil {
			return "", apperrors.New(ErrorBearerCreate, "не удалось создать токен", err)
		}
		return token, nil
	}
	if err != nil {
		return "", apperrors.New(ErrorBearerCreate, "не удалось создать токен", err)
	}

	session := sessionItem.GetModel()
	if err = s.Db().Model(sessionItem).Update("token", hash).Error; err != nil {
		return "", apperrors.New(ErrorBearerUpdate, "не удалось обновить токен", err)
	}

	session.Token = hash
	return token, nil
}

func (s *Service[T]) Logout(w http.ResponseWriter, r *http.Request) error {
	cookie, err := getCookieToken(r)
	if err != nil {
		return apperrors.New(ErrorLogoutNotFound, "сессия не найдена", err)
	}

	err = s.Db().Where("token = ?", hashToken(cookie.Value)).Delete(s.factory()).Error
	if err != nil {
		return apperrors.New(ErrorLogoutComplete, "не удалось завершить сессию", err)
	}

	deleteCookieToken(w)
	return nil
}

func (s *Service[T]) Create(w http.ResponseWriter, userID string) (*Session, error) {
	token, err := generateToken(64)
	if err != nil {
		return nil, apperrors.New(ErrorCreate, "ошибка генерации токена", err)
	}

	session := &Session{
		UserID:    userID,
		Token:     hashToken(token),
		ExpiresAt: time.Now().Add(s.sessionTTL),
	}
	if err = s.Db().Create(session).Error; err != nil {
		return nil, apperrors.New(ErrorCreateComplete, "ошибка сохранения сессии", err)
	}

	setCookieToken(w, token, session.ExpiresAt)
	return session, nil
}

func (s *Service[T]) RemoveExpired() error {
	err := s.Db().Where("expires_at > ? AND expires_at < ?", time.Time{}, time.Now()).Delete(s.factory()).Error
	return s.Error(err)
}

func (s *Service[T]) GetByToken(token string) (T, error) {
	sessionItem := s.factory()
	err := s.Db().Preload("User").First(sessionItem, "token = ?", hashToken(token)).Error
	return sessionItem, s.Error(err)
}

func (s *Service[T]) RemoveByToken(token string) {
	s.Db().Where("token = ?", hashToken(token)).Delete(s.factory())
}
