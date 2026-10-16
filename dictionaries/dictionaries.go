package dictionaries

func New[T any](root string) *Service[T] {
	service := newService[T](root)
	return service
}
