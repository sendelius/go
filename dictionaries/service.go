package dictionaries

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sendelius/go/apperrors"
)

type Service[T any] struct {
	root         string
	dictionaries []*Dictionary[T]
}

func newService[T any](root string) *Service[T] {
	return &Service[T]{
		root: root,
	}
}

func (s *Service[T]) Load(names ...string) ([]*Dictionary[T], error) {
	var dictionaries []*Dictionary[T]

	for _, name := range names {
		parts := strings.Split(strings.Trim(name, `/\`), "/")
		if len(parts) == 0 || parts[0] == "" {
			return nil, apperrors.New(ErrorEmpty, "название словаря пустое")
		}

		dir := filepath.Join(s.root, filepath.FromSlash(filepath.Join(parts[:len(parts)-1]...)))
		var files []string

		if len(parts) == 1 {
			dir = filepath.Join(s.root, parts[0])
			var err error
			files, err = filepath.Glob(filepath.Join(dir, "*.json"))
			if err != nil {
				return nil, apperrors.New(ErrorSearchFiles, fmt.Sprintf("ошибка поиска файлов словаря %s", name), err)
			}
		} else {
			file := filepath.Join(dir, parts[len(parts)-1]+".json")
			files = []string{file}
		}

		if len(files) == 0 {
			return nil, apperrors.New(ErrorJsonFiles, fmt.Sprintf("не найдены JSON-файлы словаря %s", name))
		}

		var (
			title   string
			version string
			data    []T
		)

		for _, path := range files {
			response, err := os.ReadFile(path)
			if err != nil {
				return nil, apperrors.New(ErrorReadFile, fmt.Sprintf("ошибка загрузки файла словаря %s", path), err)
			}

			var dictionary Dictionary[T]
			if err = json.Unmarshal(response, &dictionary); err != nil {
				return nil, apperrors.New(ErrorUnmarshalFile, fmt.Sprintf("ошибка разбора файла словаря %s", path), err)
			}

			if title == "" {
				title = dictionary.Title
				version = dictionary.Version
			} else {
				if dictionary.Title != title {
					return nil, apperrors.New(ErrorTitle, fmt.Sprintf("ошибка словаря %s: несоответствие заголовка в %s", name, path))
				}
				if dictionary.Version != version {
					return nil, apperrors.New(ErrorVersion, fmt.Sprintf("ошибка словаря %s: несоответствие версий в %s", name, path))
				}
			}

			data = append(data, dictionary.Data...)
		}

		dictionaries = append(dictionaries, &Dictionary[T]{
			Title:   title,
			Version: version,
			Data:    data,
		})
	}

	return dictionaries, nil
}
