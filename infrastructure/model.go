package infrastructure

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Model struct {
	CreatedAt time.Time `json:"created_at"`           // дата создания
	UpdatedAt time.Time `json:"updated_at"`           // дата обновления
	ID        string    `json:"id" gorm:"primaryKey"` // идентификатор
}

func (f *Model) BeforeCreate(*gorm.DB) error {
	f.ID = uuid.NewString()
	return nil
}
