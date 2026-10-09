package users

import (
	"github.com/sendelius/go/infrastructure"
	"gorm.io/gorm"
)

type User struct {
	infrastructure.BaseModel
	Login    string `json:"login" gorm:"uniqueIndex"` // логин
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
	if err := m.BaseModel.BeforeCreate(tx); err != nil {
		return err
	}
	return nil
}
