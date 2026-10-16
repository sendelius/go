package console

import (
	"fmt"
	"os"
)

type Module struct {
	service *Service
}

func New(appName string) *Module {
	service := newService(appName)
	module := &Module{
		service: service,
	}
	module.Register("help", "показать список команд", func(args []string) error {
		printHelp(appName, service.items)
		return nil
	})
	return module
}

func (m *Module) Start() {
	if len(os.Args) < 2 {
		printHelp(m.service.appName, m.service.items)
		os.Exit(1)
	}

	cmd := os.Args[1]

	item, ok := m.service.items[cmd]
	if !ok {
		fmt.Printf("неизвестная команда: %s\n", cmd)
		os.Exit(1)
	}

	if err := item.Handler(os.Args[2:]); err != nil {
		fmt.Printf("ошибка: %v\n", err)
		os.Exit(1)
	}
}

func (m *Module) Register(name string, description string, handler CommandHandler) {
	m.service.items[name] = &Command{
		Name:        name,
		Description: description,
		Handler:     handler,
	}
}

func (m *Module) Service() *Service {
	return m.service
}
