package console

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

type CommandHandler func(args []string) error

type Command struct {
	Name        string
	Description string
	Handler     CommandHandler
}

type Commands struct {
	items   map[string]*Command
	appName string
}

func NewCommands(appName string) *Commands {
	cmd := &Commands{
		items:   make(map[string]*Command),
		appName: appName,
	}

	cmd.Register("help", "показать список команд", func(args []string) error {
		cmd.printHelp()
		return nil
	})

	return cmd
}

func (c *Commands) Register(name string, description string, handler CommandHandler) {
	c.items[name] = &Command{
		Name:        name,
		Description: description,
		Handler:     handler,
	}
}

func (c *Commands) Execute() {
	if len(os.Args) < 2 {
		c.printHelp()
		os.Exit(1)
	}

	cmd := os.Args[1]

	item, ok := c.items[cmd]
	if !ok {
		fmt.Printf("неизвестная команда: %s\n", cmd)
		os.Exit(1)
	}

	if err := item.Handler(os.Args[2:]); err != nil {
		fmt.Printf("ошибка: %v\n", err)
		os.Exit(1)
	}
}

func (c *Commands) printHelp() {
	var names []string
	for name := range c.items {
		names = append(names, name)
	}
	sort.Strings(names)

	var b strings.Builder
	b.WriteString("Доступные команды " + c.appName + "' <command>':\n")

	for _, name := range names {
		cmd := c.items[name]
		b.WriteString(
			fmt.Sprintf("  %-15s %s\n", cmd.Name, cmd.Description),
		)
	}
	fmt.Print(b.String())
}
