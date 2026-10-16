package dictionaries

type Module[T any] struct {
	service *Service[T]
}

func New[T any](root string) *Service[T] {
	service := newService[T](root)
	return service
}

func (m *Module[T]) Service() *Service[T] {
	return m.service
}
