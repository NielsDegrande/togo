package cli

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"example.com/todo/internal/todo"
)

// Command represents a CLI command.
type Command struct {
	Name        string
	Description string
	Execute     func(service *todo.Service, args []string) error
}

// GetCommands returns all available commands.
func GetCommands() map[string]Command {
	return map[string]Command{
		"add": {
			Name:        "add",
			Description: "Add a new todo",
			Execute:     AddCommand,
		},
		"list": {
			Name:        "list",
			Description: "List todos",
			Execute:     ListCommand,
		},
		"complete": {
			Name:        "complete",
			Description: "Mark a todo as completed",
			Execute:     CompleteCommand,
		},
		"incomplete": {
			Name:        "incomplete",
			Description: "Mark a todo as not completed",
			Execute:     IncompleteCommand,
		},
		"delete": {
			Name:        "delete",
			Description: "Delete a todo",
			Execute:     DeleteCommand,
		},
		"stats": {
			Name:        "stats",
			Description: "Show todo statistics",
			Execute:     StatsCommand,
		},
		"help": {
			Name:        "help",
			Description: "Show help information",
			Execute:     func(_ *todo.Service, _ []string) error { PrintUsage(); return nil },
		},
		"version": {
			Name:        "version",
			Description: "Show version information",
			Execute:     VersionCommand,
		},
	}
}

// AddCommand handles the add command.
func AddCommand(service *todo.Service, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("description is required")
	}

	description := strings.Join(args, " ")
	todoItem, err := service.Add(description)
	if err != nil {
		return err
	}

	fmt.Printf("Added todo #%d: %s\n", todoItem.ID, todoItem.Description)
	return nil
}

// ListCommand handles the list command.
func ListCommand(service *todo.Service, args []string) error {
	var filterCompleted *bool

	flagSet := flag.NewFlagSet("list", flag.ExitOnError)
	flagSet.Usage = func() {
		_, _ = fmt.Fprintf(flagSet.Output(), "Usage: todo list [OPTIONS]\n")
		_, _ = fmt.Fprintf(flagSet.Output(), "Options:\n")
		flagSet.PrintDefaults()
	}

	showAll := flagSet.Bool("all", false, "Show all todos")
	completed := flagSet.Bool("completed", false, "Show only completed todos")
	pending := flagSet.Bool("pending", false, "Show only pending todos")

	if err := flagSet.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Determine filter.
	if *completed && *pending {
		return fmt.Errorf("cannot use both -completed and -pending flags")
	}
	if *completed {
		filterCompleted = completed
	} else if *pending {
		falseVal := false
		filterCompleted = &falseVal
	}

	var todos []todo.Todo
	if filterCompleted != nil && !*showAll {
		todos = service.GetByStatus(*filterCompleted)
	} else {
		todos = service.GetAll()
	}

	if len(todos) == 0 {
		fmt.Println("No todos found.")
		return nil
	}

	// Create tabwriter for aligned output.
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "ID\tStatus\tDescription\tCreated"); err != nil {
		return err
	}

	for _, todoItem := range todos {
		status := "[ ]"
		if todoItem.Completed {
			status = "[✓]"
		}

		created := todoItem.CreatedAt.Format("2006-01-02 15:04")
		if _, err := fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", todoItem.ID, status, todoItem.Description, created); err != nil {
			return err
		}
	}

	return w.Flush()
}

// CompleteCommand handles the complete command.
func CompleteCommand(service *todo.Service, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("todo ID is required")
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid todo ID: %s", args[0])
	}

	if err := service.Complete(id); err != nil {
		return err
	}

	fmt.Printf("Marked todo #%d as completed\n", id)
	return nil
}

// IncompleteCommand handles the incomplete command.
func IncompleteCommand(service *todo.Service, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("todo ID is required")
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid todo ID: %s", args[0])
	}

	if err := service.Incomplete(id); err != nil {
		return err
	}

	fmt.Printf("Marked todo #%d as incomplete\n", id)
	return nil
}

// DeleteCommand handles the delete command.
func DeleteCommand(service *todo.Service, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("todo ID is required")
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid todo ID: %s", args[0])
	}

	// Get todo details before deletion for confirmation
	todoItem, err := service.GetByID(id)
	if err != nil {
		return err
	}

	if err := service.Delete(id); err != nil {
		return err
	}

	fmt.Printf("Deleted todo #%d: %s\n", id, todoItem.Description)
	return nil
}

// StatsCommand handles the stats command.
func StatsCommand(service *todo.Service, args []string) error {
	stats := service.GetStats()

	fmt.Printf("Todo Statistics:\n")
	fmt.Printf("  Total: %d\n", stats.Total)
	fmt.Printf("  Completed: %d\n", stats.Completed)
	fmt.Printf("  Pending: %d\n", stats.Pending)

	if stats.Total > 0 {
		fmt.Printf("  Completion Rate: %.1f%%\n", stats.CompletionRate())
	}

	return nil
}

// VersionCommand handles the version command.
func VersionCommand(_ *todo.Service, _ []string) error {
	fmt.Printf("ToDo Manager v1.0.0\n")
	return nil
}

// PrintUsage prints the usage information.
func PrintUsage() {
	fmt.Printf(`ToDo Manager v1.0.0

USAGE:
    todo [GLOBAL OPTIONS] <COMMAND> [COMMAND OPTIONS] [ARGUMENTS...]

GLOBAL OPTIONS:
    -file <filename>    Todo storage file (default: data/todos.json)

COMMANDS:
    add <description>   Add a new todo
    list [OPTIONS]      List todos
        -all            Show all todos (default)
        -completed      Show only completed todos
        -pending        Show only pending todos
    complete <id>       Mark a todo as completed
    incomplete <id>     Mark a todo as not completed
    delete <id>         Delete a todo
    stats               Show todo statistics
    help                Show this help message
    version             Show version information

EXAMPLES:
    todo add "Buy groceries"
    todo list
    todo list -pending
    todo complete 1
    todo delete 2
    todo stats

`)
}
