package console

import (
	"fmt"
	"sort"
	"strings"
)

type Command struct {
	Name        string
	Description string
	Handler     CommandHandler
}

type CommandHandler func(args []string) error

func printHelp(appName string, items map[string]*Command) {
	var names []string
	for name := range items {
		names = append(names, name)
	}
	sort.Strings(names)

	var b strings.Builder
	b.WriteString("Доступные команды " + appName + "' <command>':\n")

	for _, name := range names {
		cmd := items[name]
		b.WriteString(
			fmt.Sprintf("  %-15s %s\n", cmd.Name, cmd.Description),
		)
	}
	fmt.Print(b.String())
}
