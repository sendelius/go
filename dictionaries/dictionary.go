package dictionaries

type Dictionary[T any] struct {
	Title   string
	Version string
	Data    []T
}

func newDictionary[T any](title string, version string, data []T) *Dictionary[T] {
	return &Dictionary[T]{
		Title:   title,
		Version: version,
		Data:    data,
	}
}
