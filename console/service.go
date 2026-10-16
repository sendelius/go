package console

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

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

func (s *Service) Register(name string, description string, handler CommandHandler) {
	s.items[name] = &Command{
		Name:        name,
		Description: description,
		Handler:     handler,
	}
}

func (s *Service) Execute() {
	if len(os.Args) < 2 {
		s.printHelp()
		os.Exit(1)
	}

	cmd := os.Args[1]

	item, ok := s.items[cmd]
	if !ok {
		fmt.Printf("неизвестная команда: %s\n", cmd)
		os.Exit(1)
	}

	if err := item.Handler(os.Args[2:]); err != nil {
		fmt.Printf("ошибка: %v\n", err)
		os.Exit(1)
	}
}

func (s *Service) printHelp() {
	var names []string
	for name := range s.items {
		names = append(names, name)
	}
	sort.Strings(names)

	var b strings.Builder
	b.WriteString("Доступные команды " + s.appName + "' <command>':\n")

	for _, name := range names {
		cmd := s.items[name]
		b.WriteString(
			fmt.Sprintf("  %-15s %s\n", cmd.Name, cmd.Description),
		)
	}
	fmt.Print(b.String())
}
