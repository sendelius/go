package session

import (
	"time"

	"github.com/sendelius/go/infrastructure"
	"github.com/sendelius/go/users"
)

type Session struct {
	infrastructure.BaseModel
	UserID    string      `json:"user_id" gorm:"index;not null"`               // идентификатор пользователя
	User      users.Model `json:"user" gorm:"foreignKey:UserID;references:ID"` // объект с пользователем
	Token     string      `json:"-" gorm:"type:char(64);uniqueIndex"`          // токен
	ExpiresAt time.Time   `json:"expires_at"`                                  // дата окончания
}

func (s *Session) GetSession() *Session {
	return s
}
