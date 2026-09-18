package session

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"
)

type Session interface {
}

type Service interface {
	Check(w http.ResponseWriter, r *http.Request) (Session, error)
}

func GenerateToken() (string, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func SetCookie(
	cookieKey string,
	w http.ResponseWriter,
	token string,
	expiresAt time.Time,
) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieKey,
		Value:    token,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		Domain:   "." + os.Getenv("BASE_DOMAIN"),
	})
}

func DeleteCookie(
	cookieKey string,
	w http.ResponseWriter,
) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieKey,
		Value:    "",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		Domain:   "." + os.Getenv("BASE_DOMAIN"),
	})

}

func GetBearerToken(r *http.Request) (string, error) {
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
