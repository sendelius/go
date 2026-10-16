package database

import (
	"fmt"
	"log"
	"regexp"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func newService() *Service {
	return &Service{}
}

func (s *Service) InitExtension(ext string) {
	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(ext) {
		log.Printf("недопустимое имя расширения: %s", ext)
	}

	err := s.db.Exec(
		fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS %s", ext),
	).Error

	if err != nil {
		log.Printf("не удалось подключить расширение: %s", ext)
	}
}

func (s *Service) InitTrigramIndex(table, column string) {
	err := s.db.Exec("CREATE INDEX IF NOT EXISTS idx_" + table + "_" + column + "_trgm ON " + table + " USING gin (" + column + " gin_trgm_ops)").Error

	if err != nil {
		log.Print("не удалось создать индекс")
	}
}

func (s *Service) AutoMigrate(dst ...interface{}) {
	err := s.db.AutoMigrate(dst...)
	if err != nil {
		log.Printf("ошибка авто миграции: %s", err.Error())
	}
}
