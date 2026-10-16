package console

func New(appName string) *Service {
	cmd := newService(appName)
	cmd.Register("help", "показать список команд", func(args []string) error {
		cmd.printHelp()
		return nil
	})
	return cmd
}
