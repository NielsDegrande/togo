package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"example.com/todo/internal/todo"
)

// JSONRepository implements todo.Repository using JSON file storage.
type JSONRepository struct {
	filename string
}

// NewJSONRepository creates a new JSON repository.
func NewJSONRepository(filename string) *JSONRepository {
	return &JSONRepository{
		filename: filename,
	}
}

// Save writes todos to a JSON file.
func (r *JSONRepository) Save(todos []todo.Todo) error {
	data, err := json.MarshalIndent(todos, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal todos: %w", err)
	}

	if err := os.WriteFile(r.filename, data, 0o600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Load reads todos from a JSON file.
func (r *JSONRepository) Load() ([]todo.Todo, error) {
	if _, err := os.Stat(r.filename); os.IsNotExist(err) {
		// File does not exist, start with empty todos.
		return []todo.Todo{}, nil
	}

	data, err := os.ReadFile(r.filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if len(data) == 0 {
		// Empty file.
		return []todo.Todo{}, nil
	}

	var todos []todo.Todo
	if err := json.Unmarshal(data, &todos); err != nil {
		return nil, fmt.Errorf("failed to unmarshal todos: %w", err)
	}

	return todos, nil
}
