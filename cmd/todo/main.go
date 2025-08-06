package main

import (
	"fmt"
	"os"

	"example.com/todo/internal/cli"
	"example.com/todo/internal/storage"
	"example.com/todo/internal/todo"
)

const defaultFilename = "data/todos.json"

func main() {
	if len(os.Args) < 2 {
		cli.PrintUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	// Parse global flags - only look for -file flag.
	filename := defaultFilename

	// Look for -file flag in args.
	for i := 0; i < len(args); i++ {
		if args[i] == "-file" && i+1 < len(args) {
			filename = args[i+1]
			// Remove -file and its value from args.
			args = append(args[:i], args[i+2:]...)
			break
		}
	}

	// Initialize dependencies.
	repo := storage.NewJSONRepository(filename)
	service := todo.NewService(repo)

	// Get available commands.
	commands := cli.GetCommands()

	cmd, exists := commands[command]
	if !exists {
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		cli.PrintUsage()
		os.Exit(1)
	}

	if err := cmd.Execute(service, args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
