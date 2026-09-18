package database

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	DB *gorm.DB
}

func NewDatabase() (*Database, error) {
	database := &Database{}
	var err error
	var dsn string

	dsn, err = genDbDsn()
	if err != nil {
		return nil, err
	}

	logLevel := logger.Silent

	if os.Getenv("DEV_MODE") == "true" {
		logLevel = logger.Info
	}

	for i := 0; i < 30; i++ {
		database.DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logLevel),
		})
		if err == nil {
			break
		}

		log.Printf("Ожидание PostgreSQL... (%d/30): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return nil, errors.New("не удалось подключится к базе данных")
	}

	return database, nil
}

func (d *Database) InitExtension(ext string) {
	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(ext) {
		log.Printf("недопустимое имя расширения: %s", ext)
	}

	err := d.DB.Exec(
		fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS %s", ext),
	).Error

	if err != nil {
		log.Printf("не удалось подключить расширение: %s", ext)
	}
}

func (d *Database) InitTrigramIndex(table, column string) {
	err := d.DB.Exec("CREATE INDEX IF NOT EXISTS idx_" + table + "_" + column + "_trgm ON " + table + " USING gin (" + column + " gin_trgm_ops)").Error

	if err != nil {
		log.Print("не удалось создать индекс")
	}
}

func (d *Database) AutoMigrate(dst ...interface{}) {
	err := d.DB.AutoMigrate(dst...)
	if err != nil {
		log.Printf("ошибка авто миграции: %s", err.Error())
	}
}

func genDbDsn() (string, error) {
	requiredEnv := []string{
		"DATABASE_USER",
		"DATABASE_PASSWORD",
		"DATABASE_HOST",
		"DATABASE_PORT",
		"DATABASE_DB",
	}

	env := make(map[string]string, len(requiredEnv))

	for _, key := range requiredEnv {
		value, ok := os.LookupEnv(key)
		if !ok || value == "" {
			return "", fmt.Errorf("переменная среды %s не установлена", key)
		}
		env[key] = value
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		env["DATABASE_USER"],
		env["DATABASE_PASSWORD"],
		env["DATABASE_HOST"],
		env["DATABASE_PORT"],
		env["DATABASE_DB"],
	)

	return dsn, nil
}
