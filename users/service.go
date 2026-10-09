package users

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/sendelius/go/infrastructure"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	PasswordChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
)

type Accessor interface {
	GetModel() *User
}

type Service[T Accessor] struct {
	*infrastructure.BaseService[T]
	factory func() T
}

func NewService(db *gorm.DB) *Service[*User] {
	return NewServiceModel[*User](db, func() *User {
		return &User{}
	})
}

func NewServiceModel[T Accessor](db *gorm.DB, factory func() T) *Service[T] {
	return &Service[T]{
		BaseService: infrastructure.NewBaseService[T](db),
		factory:     factory,
	}
}

func (s *Service[T]) HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func (s *Service[T]) CheckPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s *Service[T]) GeneratePassword(length ...int) string {
	size := 16
	if len(length) > 0 {
		size = length[0]
	}
	result := make([]byte, size)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(PasswordChars))))
		if err != nil {
			return ""
		}
		result[i] = PasswordChars[n.Int64()]
	}
	return string(result)
}

func (s *Service[T]) GenerateLogin(email string) string {
	email = strings.TrimSpace(strings.ToLower(email))
	login := email
	if index := strings.IndexByte(email, '@'); index >= 0 {
		login = email[:index]
	}
	hash := sha256.Sum256([]byte(email))
	suffix := hex.EncodeToString(hash[:])[:6]
	return login + "-" + suffix
}
