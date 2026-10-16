package database

import (
	"fmt"
	"os"
)

func genDbDsn() (string, error) {
	requiredEnv := []string{
		"POSTGRES_USER",
		"POSTGRES_PASSWORD",
		"POSTGRES_HOST",
		"POSTGRES_PORT",
		"POSTGRES_DB",
	}

	e := make(map[string]string, len(requiredEnv))

	for _, key := range requiredEnv {
		value, ok := os.LookupEnv(key)
		if !ok || value == "" {
			return "", fmt.Errorf("переменная среды %s не установлена", key)
		}
		e[key] = value
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		e["POSTGRES_USER"],
		e["POSTGRES_PASSWORD"],
		e["POSTGRES_HOST"],
		e["POSTGRES_PORT"],
		e["POSTGRES_DB"],
	)

	return dsn, nil
}
