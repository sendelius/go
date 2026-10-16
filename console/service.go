package console

type Service struct {
	items   map[string]*Command
	appName string
}

func newService(appName string) *Service {
	return &Service{
		items:   make(map[string]*Command),
		appName: appName,
	}
}
