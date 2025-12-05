package commands

import "github.com/siaji/chauffeur/cli/lib"

// RunHelloWorld handles `chauf hello-world` command invocations.
func RunHelloWorld(args []string) error {
	logger := lib.NewCommandLogger("hello-world")

	for _, arg := range args {
		switch arg {
		case "--help", "-h":
			printHelloWorldUsage()
			return nil
		default:
			return logger.Error("unknown flag for hello-world", arg)
		}
	}

	logger.Success("Hello, World!", "Chauffeur greeting")
	logger.Info("Welcome to Chauffeur - your Linux PHP development environment")

	return nil
}

func printHelloWorldUsage() {
	logger := lib.NewCommandLogger("hello-world")
	logger.PrintBlock(`Usage: chauf hello-world

Prints a friendly greeting message.

Options:
  -h, --help   Show this help message`)
}
