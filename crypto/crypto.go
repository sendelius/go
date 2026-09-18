package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	passwordLength = 16
	passwordChars  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
)

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func CheckPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GeneratePassword() string {
	result := make([]byte, passwordLength)

	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(passwordChars))))
		if err != nil {
			return ""
		}

		result[i] = passwordChars[n.Int64()]
	}

	return string(result)
}

func GenerateLogin(email string) string {
	email = strings.TrimSpace(strings.ToLower(email))

	login := email

	if index := strings.IndexByte(email, '@'); index >= 0 {
		login = email[:index]
	}

	hash := sha256.Sum256([]byte(email))
	suffix := hex.EncodeToString(hash[:])[:6]

	return login + "-" + suffix
}
