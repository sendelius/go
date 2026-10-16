package console

type Command struct {
	Name        string
	Description string
	Handler     CommandHandler
}

type CommandHandler func(args []string) error
