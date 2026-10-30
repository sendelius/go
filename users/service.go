package users

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"net/http"
	"strings"

	"github.com/sendelius/go/apperrors"
	"github.com/sendelius/go/database"
	"github.com/sendelius/go/infrastructure"
	"golang.org/x/crypto/bcrypt"
)

const (
	PasswordChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
)

type Accessor interface {
	GetModel() *User
}

type ServiceInterface[T Accessor] interface {
	NewModel() T
	GetByLogin(login string) (T, error)
	Check(login, password string) (T, error)
	Create(item T) (T, error)
	GetByID(id string) (T, error)
	HashPassword(password string) string
	CheckPassword(hash, password string) bool
	GeneratePassword(length ...int) string
	GenerateLogin(email string) string
}

type Service[T Accessor] struct {
	*infrastructure.Service[T]
	factory func() T
}

func newService[T Accessor](db *database.Service, factory func() T) *Service[T] {
	return &Service[T]{
		Service: infrastructure.NewService[T](db),
		factory: factory,
	}
}

func (s *Service[T]) NewModel() T {
	return s.factory()
}

func (s *Service[T]) GetByLogin(login string) (T, error) {
	item := s.factory()
	err := s.Db().First(item, "login = ?", login).Error
	return item, s.Error(err)
}

func (s *Service[T]) Check(login string, password string) (T, error) {
	userItem, err := s.GetByLogin(login)
	if err != nil {
		var zero T
		return zero, apperrors.New(ErrorNotFound, "пользователь не найден").
			Field("login").
			Status(http.StatusUnauthorized).
			Build()
	}

	user := userItem.GetModel()
	if user.Status != "public" {
		var zero T
		return zero, apperrors.New(ErrorNotActive, "пользователь не активирован, ожидайте решения администрации").
			Field("login").
			Status(http.StatusUnauthorized).
			Build()
	}

	if !s.CheckPassword(user.Password, password) {
		var zero T
		return zero, apperrors.New(ErrorPasswordInvalid, "неверный пароль").
			Field("password").
			Status(http.StatusUnauthorized).
			Build()
	}

	return userItem, nil
}

func (s *Service[T]) Create(item T) (T, error) {
	var zero T
	model := item.GetModel()

	if len(model.Login) == 0 && len(model.Email) > 0 {
		model.Login = s.GenerateLogin(model.Email)
	}

	if len(model.Password) == 0 {
		model.Password = s.HashPassword(s.GeneratePassword())
	}

	if len(model.Email) > 0 {
		err := s.Db().First(&model, "email = ?", model.Email).Error
		if err == nil {
			if model.Status != "public" {
				return zero, apperrors.New(ErrorCreateAlreadySubmitted, "по указанному email уже оформлена заявка на регистрацию").
					Cause(err).
					Field("email").
					Build()
			}
			return zero, apperrors.New(ErrorCreateAlreadyRegistration, "по указанному email уже зарегистрирован пользователь").
				Cause(err).
				Field("email").
				Build()
		}
	} else {
		return zero, apperrors.New(ErrorCreateEmailRequired, "поле email обязательно для заполнения").
			Field("email").
			Build()
	}

	if len(model.Phone) > 0 {
		err := s.Db().First(&model, "phone = ?", model.Phone).Error
		if err == nil {
			if model.Status != "public" {
				return zero, apperrors.New(ErrorCreateAlreadySubmitted, "по указанному телефону уже оформлена заявка на регистрацию").
					Cause(err).
					Field("phone").
					Build()
			}
			return zero, apperrors.New(ErrorCreateAlreadyRegistration, "по указанному телефону уже зарегистрирован пользователь").
				Cause(err).
				Field("phone").
				Build()
		}
	}

	if err := s.Db().Create(item).Error; err != nil {
		return zero, apperrors.New(ErrorCreate, "ошибка сохранения пользователя").
			Cause(err).
			Field("name").
			Build()
	}
	return item, nil
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
