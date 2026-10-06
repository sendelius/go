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

type Service struct {
	*infrastructure.BaseService[Session]
	sessionTTL time.Duration
	cookieKey  string
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		infrastructure.NewBaseService[Session](db),
		time.Duration(env.Int("SESSION_TTL", 172800)) * time.Second,
		"__Secure-snd_session_token",
	}
}

func (s *Service) Check(w http.ResponseWriter, r *http.Request) (*Session, error) {
	if !env.Bool("SESSION_ALLOW") {
		return nil, nil
	}

	var token string

	if bearer, err := s.getBearerToken(r); err == nil {
		token = strings.TrimSpace(bearer)
	} else if cookie, err := r.Cookie(s.cookieKey); err == nil {
		token = strings.TrimSpace(cookie.Value)
	}
	if token == "" {
		return nil, nil
	}
	sessionItem, err := s.getByToken(token)
	if err != nil {
		return nil, errors.New("сессия не найдена")
	}
	now := time.Now()
	if !sessionItem.ExpiresAt.IsZero() && sessionItem.ExpiresAt.Before(now) {
		s.removeByToken(token)
		return nil, errors.New("сессия истекла")
	}
	if !sessionItem.ExpiresAt.IsZero() && sessionItem.ExpiresAt.Sub(now) < s.sessionTTL-time.Hour {
		expiresAt := now.Add(s.sessionTTL)
		sessionItem.ExpiresAt = expiresAt
		err = s.Db.
			Model(&Session{}).
			Where("id = ?", sessionItem.ID).
			Update("expires_at", expiresAt).
			Error
		if err != nil {
			return nil, errors.New("не удалось продлить сессию")
		}
		s.setCookie(w, token, expiresAt)
	}
	return &sessionItem, nil
}

func (s *Service) Login(w http.ResponseWriter, userID string) (*Session, error) {
	if !env.Bool("SESSION_ALLOW") {
		return nil, nil
	}
	token, err := s.generateToken(32)
	if err != nil {
		return nil, errors.New("не удалось создать сессию")
	}
	expiresAt := time.Now().Add(s.sessionTTL)
	session := &Session{
		UserID:    userID,
		Token:     s.hashToken(token),
		ExpiresAt: expiresAt,
	}
	if err := s.Db.Create(session).Error; err != nil {
		return nil, errors.New("не удалось сохранить сессию")
	}
	s.setCookie(w, token, expiresAt)
	return session, nil
}

func (s *Service) Bearer(userID string) (string, error) {
	if !env.Bool("SESSION_ALLOW") {
		return "", nil
	}
	token, err := s.generateToken(64)
	if err != nil {
		return "", errors.New("не удалось создать токен")
	}
	hash := s.hashToken(token)
	var session Session
	err = s.Db.Where("user_id = ? AND expires_at = ?", userID, time.Time{}).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		session = Session{
			UserID: userID,
			Token:  hash,
		}
		if err := s.Db.Create(&session).Error; err != nil {
			return "", errors.New("не удалось сохранить токен")
		}
		return token, nil
	}
	if err != nil {
		return "", err
	}
	if err := s.Db.Model(&session).Update("token", hash).Error; err != nil {
		return "", errors.New("не удалось обновить токен")
	}
	return token, nil
}

func (s *Service) Logout(w http.ResponseWriter, r *http.Request) error {
	if !env.Bool("SESSION_ALLOW") {
		return nil
	}
	cookie, err := r.Cookie(s.cookieKey)
	if err != nil {
		return errors.New("сессия не найдена")
	}
	s.Db.Where("token = ?", s.hashToken(cookie.Value)).Delete(&Session{})
	s.deleteCookie(w)
	return nil
}

func (s *Service) CurrentUser(session *Session) (*users.Model, error) {
	if session == nil {
		return nil, errors.New("сессия не найдена")
	}
	return &session.User, nil
}

func (s *Service) RemoveExpired() error {
	err := s.Db.Where("expires_at < ?", time.Now()).Delete(&Session{}).Error
	return s.Error(err)
}

func (s *Service) getByToken(token string) (Session, error) {
	var item Session
	hash := s.hashToken(token)
	err := s.Db.Preload("User").First(&item, "token = ?", hash).Error
	return item, s.Error(err)
}

func (s *Service) removeByToken(token string) {
	hash := s.hashToken(token)
	s.Db.Where("token = ?", hash).Delete(&Session{})
}

func (s *Service) generateToken(size int) (string, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *Service) hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (s *Service) setCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
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

func (s *Service) deleteCookie(w http.ResponseWriter) {
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

func (s *Service) getBearerToken(r *http.Request) (string, error) {
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
