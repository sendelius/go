package users

import (
	"github.com/sendelius/go/infrastructure"
	"gorm.io/gorm"
)

type User struct {
	infrastructure.Model
	Login    string `json:"login" gorm:"uniqueIndex"` // логин
	Name     string `json:"name"`                     // имя
	Email    string `json:"email"`                    // email
	Phone    string `json:"phone"`                    // телефон
	Password string `json:"-"`                        // пароль
	Status   string `json:"status"`                   // статус
}

func (m *User) GetModel() *User {
	return m
}

func (m *User) BeforeSave(*gorm.DB) error {
	if m.Status == "" {
		m.Status = "pending"
	}
	return nil
}

func (m *User) BeforeCreate(tx *gorm.DB) error {
	if err := m.Model.BeforeCreate(tx); err != nil {
		return err
	}
	return nil
}
