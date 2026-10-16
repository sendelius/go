package database

import (
	"log"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func newService() *Service {
	return &Service{}
}

func (s *Service) Db() *gorm.DB {
	return s.db
}

func (s *Service) AutoMigrate(dst ...interface{}) {
	err := s.db.AutoMigrate(dst...)
	if err != nil {
		log.Printf("ошибка авто миграции: %s", err.Error())
	}
}
