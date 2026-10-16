package dictionaries

type Module struct {
	service *Service
}

func New(root string) *Module {
	return &Module{
		service: newService(root),
	}
}

func (m *Module) Service() *Service {
	return m.service
}
