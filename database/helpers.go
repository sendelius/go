package database

import (
	"fmt"
	"strings"

	"github.com/sendelius/go/env"
)

func genDbDsn(name string) (string, error) {
	prefix := "POSTGRES"
	user := env.String("POSTGRES_USER", "postgres")
	password := env.String("POSTGRES_PASSWORD", "postgres")
	host := env.String("POSTGRES_HOST", "postgres")
	port := env.Int("POSTGRES_PORT", 5432)
	db := env.String("POSTGRES_DB", "snd")
	if name != "" {
		prefix += "_" + strings.ToUpper(name)
		user = env.String(prefix+"_USER", user)
		password = env.String(prefix+"_PASSWORD", password)
		host = env.String(prefix+"_HOST", host)
		port = env.Int(prefix+"_PORT", port)
		db = env.String(prefix+"_DB", strings.ToLower(name))
	}
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		user,
		password,
		host,
		port,
		db,
	)
	return dsn, nil
}
