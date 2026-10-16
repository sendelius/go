package database

import (
	"log"
	"time"

	"github.com/sendelius/go/env"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Module struct {
	service *Service
}

func New() *Module {
	service := newService()
	var err error
	var dsn string

	dsn, err = genDbDsn()
	if err != nil {
		panic(err)
	}

	logLevel := logger.Silent
	if env.Bool("DEV_MODE") {
		logLevel = logger.Info
	}

	for i := 0; i < 30; i++ {
		service.db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logLevel),
		})
		if err == nil {
			break
		}

		log.Printf("Ожидание PostgreSQL... (%d/30): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		panic("не удалось подключится к базе данных")
	}

	return &Module{
		service: service,
	}
}

func (m *Module) Service() *Service {
	return m.service
}
