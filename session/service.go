package session

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/sendelius/go/env"
	"github.com/sendelius/go/infrastructure"
	"github.com/sendelius/go/users"
	"gorm.io/gorm"
)

type Accessor interface {
	GetSession() *Session
}

type Service[T Accessor] struct {
	*infrastructure.BaseService[T]
	sessionTTL time.Duration
	cookieKey  string
	factory    func() T
}

func NewService(db *gorm.DB) *Service[*Session] {
	return NewServiceModel[*Session](db, func() *Session {
		return &Session{}
	})
}

func NewServiceModel[T Accessor](
	db *gorm.DB,
	factory func() T,
) *Service[T] {
	return &Service[T]{
		BaseService: infrastructure.NewBaseService[T](db),
		sessionTTL:  time.Duration(env.Int("SESSION_TTL", 172800)) * time.Second,
		cookieKey:   "__Secure-snd_session_token",
		factory:     factory,
	}
}

func (s *Service[T]) Check(w http.ResponseWriter, r *http.Request) (T, error) {
	var zero T
	if !env.Bool("SESSION_ALLOW") {
		return zero, nil
	}
	var token string
	if bearer, err := s.getBearerToken(r); err == nil {
		token = strings.TrimSpace(bearer)
	} else if cookie, err := r.Cookie(s.cookieKey); err == nil {
		token = strings.TrimSpace(cookie.Value)
	}
	if token == "" {
		return zero, nil
	}
	sessionItem, err := s.getByToken(token)
	if err != nil {
		return zero, errors.New("сессия не найдена")
	}
	session := sessionItem.GetSession()
	now := time.Now()
	if !session.ExpiresAt.IsZero() && session.ExpiresAt.Before(now) {
		s.removeByToken(token)
		return zero, errors.New("сессия истекла")
	}
	if !session.ExpiresAt.IsZero() && session.ExpiresAt.Sub(now) < s.sessionTTL-time.Hour {
		expiresAt := now.Add(s.sessionTTL)
		session.ExpiresAt = expiresAt
		err = s.Db.Model(sessionItem).Where("id = ?", session.ID).Update("expires_at", expiresAt).Error
		if err != nil {
			return zero, errors.New("не удалось продлить сессию")
		}
		s.setCookie(w, token, expiresAt)
	}
	return sessionItem, nil
}

func (s *Service[T]) Login(w http.ResponseWriter, userID string) (T, error) {
	var zero T
	if !env.Bool("SESSION_ALLOW") {
		return zero, nil
	}
	token, err := s.generateToken(32)
	if err != nil {
		return zero, errors.New("не удалось создать сессию")
	}
	expiresAt := time.Now().Add(s.sessionTTL)
	sessionItem := s.factory()
	session := sessionItem.GetSession()
	session.UserID = userID
	session.Token = s.hashToken(token)
	session.ExpiresAt = expiresAt
	if err := s.Db.Create(sessionItem).Error; err != nil {
		return zero, errors.New("не удалось сохранить сессию")
	}

	s.setCookie(w, token, expiresAt)
	return sessionItem, nil
}

func (s *Service[T]) Bearer(userID string) (string, error) {
	if !env.Bool("SESSION_ALLOW") {
		return "", nil
	}
	token, err := s.generateToken(64)
	if err != nil {
		return "", errors.New("не удалось создать токен")
	}
	hash := s.hashToken(token)
	sessionItem := s.factory()
	err = s.Db.Where("user_id = ? AND expires_at = ?", userID, time.Time{}).First(sessionItem).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		session := sessionItem.GetSession()
		session.UserID = userID
		session.Token = hash
		if err := s.Db.Create(sessionItem).Error; err != nil {
			return "", errors.New("не удалось сохранить токен")
		}
		return token, nil
	}
	if err != nil {
		return "", err
	}
	session := sessionItem.GetSession()
	if err := s.Db.
		Model(sessionItem).
		Update("token", hash).
		Error; err != nil {
		return "", errors.New("не удалось обновить токен")
	}
	session.Token = hash
	return token, nil
}

func (s *Service[T]) Logout(w http.ResponseWriter, r *http.Request) error {
	if !env.Bool("SESSION_ALLOW") {
		return nil
	}
	cookie, err := r.Cookie(s.cookieKey)
	if err != nil {
		return errors.New("сессия не найдена")
	}
	s.Db.Where("token = ?", s.hashToken(cookie.Value)).Delete(s.factory())
	s.deleteCookie(w)
	return nil
}

func (s *Service[T]) CurrentUser(sessionItem T) (*users.Model, error) {
	session := sessionItem.GetSession()
	return &session.User, nil
}

func (s *Service[T]) RemoveExpired() error {
	err := s.Db.Where("expires_at < ?", time.Now()).Delete(s.factory()).Error
	return s.Error(err)
}

func (s *Service[T]) getByToken(token string) (T, error) {
	sessionItem := s.factory()
	hash := s.hashToken(token)
	err := s.Db.Preload("User").First(sessionItem, "token = ?", hash).Error
	return sessionItem, s.Error(err)
}

func (s *Service[T]) removeByToken(token string) {
	hash := s.hashToken(token)
	s.Db.Where("token = ?", hash).Delete(s.factory())
}

func (s *Service[T]) generateToken(size int) (string, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *Service[T]) hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (s *Service[T]) setCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.cookieKey,
		Value:    token,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		Domain:   "." + env.String("BASE_DOMAIN"),
	})
}

func (s *Service[T]) deleteCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.cookieKey,
		Value:    "",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		Domain:   "." + env.String("BASE_DOMAIN"),
	})

}

func (s *Service[T]) getBearerToken(r *http.Request) (string, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", errors.New("заголовок Authorization отсутствует")
	}

	if !strings.HasPrefix(auth, "Bearer ") {
		return "", errors.New("неверный формат Authorization")
	}

	token := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	if token == "" {
		return "", errors.New("пустой токен")
	}

	return token, nil
}
