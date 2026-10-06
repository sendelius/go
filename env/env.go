package env

import (
	"os"
	"strconv"
	"strings"
)

func String(key string, fallback ...string) string {
	value := os.Getenv(key)
	if value == "" && len(fallback) > 0 {
		return fallback[0]
	}
	return value
}

func Bool(key string, fallback ...bool) bool {
	value := os.Getenv(key)
	if value == "" {
		if len(fallback) > 0 {
			return fallback[0]
		}
		return false
	}
	result, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}
	return result
}

func Int(key string, fallback ...int) int {
	value := os.Getenv(key)
	if value == "" {
		if len(fallback) > 0 {
			return fallback[0]
		}
		return 0
	}
	result, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return result
}

func Float(key string, fallback ...float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		if len(fallback) > 0 {
			return fallback[0]
		}
		return 0
	}
	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return result
}

func Array(key string, separator string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return nil
	}
	items := strings.Split(value, separator)
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}
