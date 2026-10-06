package users

import (
	"github.com/sendelius/go/infrastructure"
	"gorm.io/gorm"
)

type Model struct {
	infrastructure.BaseModel
	Login    string `json:"login" gorm:"uniqueIndex"` // логин
	Password string `json:"-"`                        // пароль
	Status   string `json:"status"`                   // статус
}

func (m *Model) BeforeSave(*gorm.DB) error {
	if m.Status == "" {
		m.Status = "pending"
	}
	return nil
}

func (m *Model) BeforeCreate(tx *gorm.DB) error {
	if err := m.BaseModel.BeforeCreate(tx); err != nil {
		return err
	}
	return nil
}
