package server

import (
	"encoding"
	"fmt"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	timeType            = reflect.TypeOf(time.Time{})
	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
)

type queryPart struct {
	name    string
	bracket bool
}

// DecodeQuery заполняет структуру значениями из URL query.
func DecodeQuery(dst any, query url.Values) error {
	if dst == nil {
		return fmt.Errorf("destination cannot be nil")
	}

	v := reflect.ValueOf(dst)

	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("destination must be a non-nil pointer")
	}

	v = v.Elem()

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("destination must point to a struct")
	}

	keys := make([]string, 0, len(query))

	for key := range query {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		parts, err := parseQueryPath(key)
		if err != nil {
			return err
		}

		if len(parts) == 0 {
			return fmt.Errorf("empty query parameter: %q", key)
		}

		if err := assignQueryValue(v, parts, query[key]); err != nil {
			return fmt.Errorf("query parameter %q: %w", key, err)
		}
	}

	return nil
}

// parseQueryPath разбирает filter.status, sort[field], items[0].name.
func parseQueryPath(path string) ([]queryPart, error) {
	var parts []queryPart
	var name strings.Builder

	for i := 0; i < len(path); i++ {
		switch path[i] {
		case '.':
			if name.Len() == 0 {
				return nil, fmt.Errorf("invalid query path %q", path)
			}

			parts = append(parts, queryPart{name: name.String()})
			name.Reset()

		case '[':
			if name.Len() > 0 {
				parts = append(parts, queryPart{name: name.String()})
				name.Reset()
			}

			end := strings.IndexByte(path[i+1:], ']')
			if end < 0 {
				return nil, fmt.Errorf("invalid query path %q", path)
			}

			parts = append(parts, queryPart{
				name:    path[i+1 : i+1+end],
				bracket: true,
			})

			i += end + 1

		case ']':
			return nil, fmt.Errorf("invalid query path %q", path)

		default:
			name.WriteByte(path[i])
		}
	}

	if name.Len() > 0 {
		parts = append(parts, queryPart{name: name.String()})
	}

	return parts, nil
}

// assignQueryValue рекурсивно заполняет поля структуры.
func assignQueryValue(v reflect.Value, parts []queryPart, values []string) error {
	if len(parts) == 0 {
		return setQueryScalar(v, values)
	}

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}

		return assignQueryValue(v.Elem(), parts, values)
	}

	part := parts[0]

	switch v.Kind() {
	case reflect.Struct:
		field, ok := findQueryField(v, part.name)
		if !ok {
			return fmt.Errorf("unknown field %q", part.name)
		}

		return assignQueryValue(field, parts[1:], values)

	case reflect.Map:
		return assignQueryMap(v, parts, values)

	case reflect.Slice, reflect.Array:
		return assignQueryCollection(v, parts, values)

	default:
		return fmt.Errorf("cannot access field %q on %s", part.name, v.Kind())
	}
}

// findQueryField ищет поле по json-тегу или имени Go в snake_case.
func findQueryField(v reflect.Value, name string) (reflect.Value, bool) {
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if !field.IsExported() {
			continue
		}

		tag := strings.Split(field.Tag.Get("json"), ",")[0]

		if tag == "-" {
			continue
		}

		if tag == "" {
			tag = toSnakeCase(field.Name)
		}

		if tag == name || field.Name == name {
			return v.Field(i), true
		}
	}

	return reflect.Value{}, false
}

func toSnakeCase(s string) string {
	var result strings.Builder

	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}

		if r >= 'A' && r <= 'Z' {
			r += 'a' - 'A'
		}

		result.WriteRune(r)
	}

	return result.String()
}

// assignQueryMap заполняет map по ключу из квадратных скобок.
func assignQueryMap(v reflect.Value, parts []queryPart, values []string) error {
	if v.IsNil() {
		v.Set(reflect.MakeMap(v.Type()))
	}

	key, err := parseMapKey(parts[0].name, v.Type().Key())
	if err != nil {
		return err
	}

	item := v.MapIndex(key)

	if !item.IsValid() {
		item = reflect.New(v.Type().Elem()).Elem()
	} else {
		copy := reflect.New(v.Type().Elem()).Elem()
		copy.Set(item)
		item = copy
	}

	if err := assignQueryValue(item, parts[1:], values); err != nil {
		return err
	}

	v.SetMapIndex(key, item)

	return nil
}

// parseMapKey преобразует строковый ключ в тип ключа map.
func parseMapKey(value string, typ reflect.Type) (reflect.Value, error) {
	key := reflect.New(typ).Elem()

	if err := setQueryScalar(key, []string{value}); err != nil {
		return reflect.Value{}, fmt.Errorf("invalid map key %q: %w", value, err)
	}

	return key, nil
}

// assignQueryCollection обрабатывает индексы срезов и массивов.
func assignQueryCollection(v reflect.Value, parts []queryPart, values []string) error {
	index := -1

	if parts[0].name != "" {
		var err error

		index, err = strconv.Atoi(parts[0].name)
		if err != nil || index < 0 {
			return fmt.Errorf("invalid collection index %q", parts[0].name)
		}
	}

	if v.Kind() == reflect.Slice {
		if index < 0 {
			index = v.Len()
		}

		if index >= v.Len() {
			if index > 10000 {
				return fmt.Errorf("collection index is too large")
			}

			newSlice := reflect.MakeSlice(v.Type(), index+1, index+1)
			reflect.Copy(newSlice, v)
			v.Set(newSlice)
		}
	} else if index < 0 || index >= v.Len() {
		return fmt.Errorf("array index %d out of range", index)
	}

	return assignQueryValue(v.Index(index), parts[1:], values)
}

// setQueryScalar устанавливает значение конкретного поля.
func setQueryScalar(v reflect.Value, values []string) error {
	if !v.CanSet() {
		return fmt.Errorf("field cannot be set")
	}

	if len(values) == 0 {
		return nil
	}

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}

		return setQueryScalar(v.Elem(), values)
	}

	// Поддержка пользовательских типов и time.Time.
	if v.CanAddr() && v.Addr().Type().Implements(textUnmarshalerType) {
		return v.Addr().Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(values[0]))
	}

	// Срез принимает несколько одинаковых параметров.
	if v.Kind() == reflect.Slice {
		result := reflect.MakeSlice(v.Type(), 0, len(values))

		for _, value := range values {
			item := reflect.New(v.Type().Elem()).Elem()

			if err := setQueryScalar(item, []string{value}); err != nil {
				return err
			}

			result = reflect.Append(result, item)
		}

		v.Set(result)
		return nil
	}

	if len(values) > 1 {
		return fmt.Errorf("multiple values for scalar field")
	}

	value := values[0]

	switch v.Kind() {
	case reflect.String:
		v.SetString(value)

	case reflect.Bool:
		n, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		v.SetBool(n)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(value, 10, v.Type().Bits())
		if err != nil {
			return err
		}
		v.SetInt(n)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := strconv.ParseUint(value, 10, v.Type().Bits())
		if err != nil {
			return err
		}
		v.SetUint(n)

	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(value, v.Type().Bits())
		if err != nil {
			return err
		}
		v.SetFloat(n)

	default:
		return fmt.Errorf("unsupported type %s", v.Type())
	}

	return nil
}
