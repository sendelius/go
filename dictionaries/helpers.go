package dictionaries

import (
	"reflect"
	"strings"
)

func getField[T any](item T, field string) string {
	value := reflect.ValueOf(item)

	if value.Kind() == reflect.Pointer {
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return ""
	}

	fieldValue := value.FieldByNameFunc(func(name string) bool {
		return strings.EqualFold(name, field)
	})

	if !fieldValue.IsValid() {
		return ""
	}

	if fieldValue.Kind() != reflect.String {
		return ""
	}

	return fieldValue.String()
}

func score(value, query string) float64 {
	value = normalize(value)
	query = normalize(query)

	if value == "" || query == "" {
		return 0
	}

	if value == query {
		return 1.0
	}

	if strings.HasPrefix(value, query) {
		return 0.9
	}

	if strings.Contains(value, query) {
		return 0.75
	}

	return 0
}

func similarityScore(value, query string) float64 {
	value = normalize(value)
	query = normalize(query)

	if value == "" || query == "" {
		return 0
	}

	if value == query {
		return 1
	}

	if strings.HasPrefix(value, query) {
		return 0.9
	}

	if strings.Contains(value, query) {
		return 0.75
	}

	distance := levenshtein(value, query)

	maxLength := max(len([]rune(value)), len([]rune(query)))

	if maxLength == 0 {
		return 0
	}

	result := 1.0 - float64(distance)/float64(maxLength)

	if result < 0 {
		return 0
	}

	return result
}

func levenshtein(a, b string) int {
	ar := []rune(a)
	br := []rune(b)

	if len(ar) == 0 {
		return len(br)
	}

	if len(br) == 0 {
		return len(ar)
	}

	prev := make([]int, len(br)+1)
	curr := make([]int, len(br)+1)

	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= len(ar); i++ {
		curr[0] = i

		for j := 1; j <= len(br); j++ {
			cost := 0

			if ar[i-1] != br[j-1] {
				cost = 1
			}

			curr[j] = min(
				curr[j-1]+1,
				prev[j]+1,
				prev[j-1]+cost,
			)
		}

		prev, curr = curr, prev
	}

	return prev[len(br)]
}

func normalize(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ToLower(value)

	return strings.ReplaceAll(value, "ё", "е")
}
