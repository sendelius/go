package redis

type Module struct {
	service *Service
}

func New() *Module {
	service := newService()

	return &Module{
		service: service,
	}
}

func (m *Module) Service() *Service {
	return m.service
}
