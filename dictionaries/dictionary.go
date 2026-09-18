package dictionaries

type Dictionary[T any] struct {
	Title   string
	Version string
	Data    []T
}

func NewDictionary[T any](
	title string,
	version string,
	data []T,
) *Dictionary[T] {
	return &Dictionary[T]{
		Title:   title,
		Version: version,
		Data:    data,
	}
}
