package dictionaries

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Manager struct {
	root string
}

func NewManager(root string) *Manager {
	return &Manager{
		root: root,
	}
}

type dictionaryFile[T any] struct {
	Title   string `json:"title"`
	Version string `json:"version"`
	Data    []T    `json:"data"`
}

func Load[T any](
	manager *Manager,
	name string,
) (*Dictionary[T], error) {
	parts := strings.Split(strings.Trim(name, "/\\"), "/")
	if len(parts) == 0 {
		return nil, fmt.Errorf("название словаря пустое")
	}

	dir := filepath.Join(
		append([]string{manager.root}, parts[:len(parts)-1]...)...,
	)

	var files []string
	if len(parts) == 1 {
		dir = filepath.Join(manager.root, parts[0])
		_, err := filepath.Glob(
			filepath.Join(dir, "*.json"),
		)
		if err != nil {
			return nil, fmt.Errorf("не найдены файлы словаря %s: %w", name, err)
		}
	} else {
		file := filepath.Join(parts[len(parts)-1] + ".json")
		files = []string{filepath.Join(dir, file)}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("словарь %s не является типом json", name)
	}

	var (
		title   string
		version string
		data    []T
	)

	for _, path := range files {
		response, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("ошибка загрузки файла словаря %s: %w", path, err)
		}

		var dictionary dictionaryFile[T]
		if err := json.Unmarshal(response, &dictionary); err != nil {
			return nil, fmt.Errorf("ошибка разбора файла словаря %s: %w\n", path, err)
		}

		if title == "" {
			title = dictionary.Title
		}

		if version == "" {
			version = dictionary.Version
		}

		if dictionary.Title != title {
			return nil, fmt.Errorf("ошибка словаря %s: несоответствие заголовка в %s", name, path)
		}

		if dictionary.Version != version {
			return nil, fmt.Errorf("ошибка словаря %s: несоответствие версий в %s\n", name, path)
		}

		data = append(data, dictionary.Data...)
	}

	return NewDictionary(title, version, data), nil
}
