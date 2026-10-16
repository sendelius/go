package infrastructure

import (
	"regexp"

	"github.com/jackc/pgx/v5/pgconn"
)

var keyRe = regexp.MustCompile(`Key \((.*?)\)=`)

func getFieldName(pgErr *pgconn.PgError) string {
	if pgErr.ColumnName != "" {
		return pgErr.ColumnName
	}

	if m := keyRe.FindStringSubmatch(pgErr.Detail); len(m) == 2 {
		return m[1]
	}

	return ""
}
