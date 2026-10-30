package infrastructure

import (
	"reflect"
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

func resetCloneFields(value reflect.Value) {
	for _, name := range []string{"ID", "CreatedAt", "UpdatedAt"} {
		field := value.FieldByName(name)
		if field.IsValid() && field.CanSet() {
			field.Set(reflect.Zero(field.Type()))
		}
	}
}

func uniqueIDs(ids []string) []string {
	unique := make([]string, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	return unique
}

func getData(data []map[string]any) map[string]any {
	if len(data) == 0 {
		return nil
	}

	return data[0]
}
