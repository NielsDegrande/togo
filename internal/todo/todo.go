package todo

import (
	"fmt"
	"time"
)

// Todo represents a single todo item.
type Todo struct {
	ID          int        `json:"id"`
	Description string     `json:"description"`
	Completed   bool       `json:"completed"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// Stats represents todo statistics.
type Stats struct {
	Total     int
	Completed int
	Pending   int
}

// CompletionRate calculates the completion rate as a percentage.
func (s Stats) CompletionRate() float64 {
	if s.Total == 0 {
		return 0
	}
	return float64(s.Completed) / float64(s.Total) * 100
}

// Repository defines the interface for todo storage operations.
type Repository interface {
	Save(todos []Todo) error
	Load() ([]Todo, error)
}

// Service handles business logic for todo operations.
type Service struct {
	repo   Repository
	todos  []Todo
	nextID int
}

// NewService creates a new todo service.
func NewService(repo Repository) *Service {
	service := &Service{
		repo:   repo,
		todos:  make([]Todo, 0),
		nextID: 1,
	}

	if err := service.loadTodos(); err != nil {
		// Log error but do not fail, as this might be the first run.
		fmt.Printf("Warning: could not load existing todos: %v\n", err)
	}

	return service
}

// Add creates a new todo item.
func (s *Service) Add(description string) (*Todo, error) {
	if description == "" {
		return nil, fmt.Errorf("description cannot be empty")
	}

	todo := Todo{
		ID:          s.nextID,
		Description: description,
		Completed:   false,
		CreatedAt:   time.Now(),
	}

	s.todos = append(s.todos, todo)
	s.nextID++

	if err := s.save(); err != nil {
		return nil, fmt.Errorf("failed to save todo: %w", err)
	}

	return &todo, nil
}

// GetAll returns all todos.
func (s *Service) GetAll() []Todo {
	return s.todos
}

// GetByStatus returns todos filtered by completion status.
func (s *Service) GetByStatus(completed bool) []Todo {
	var filtered []Todo
	for _, todo := range s.todos {
		if todo.Completed == completed {
			filtered = append(filtered, todo)
		}
	}
	return filtered
}

// GetByID returns a todo by its ID.
func (s *Service) GetByID(id int) (*Todo, error) {
	for i, todo := range s.todos {
		if todo.ID == id {
			return &s.todos[i], nil
		}
	}
	return nil, fmt.Errorf("todo with ID %d not found", id)
}

// Complete marks a todo as completed.
func (s *Service) Complete(id int) error {
	todo, err := s.GetByID(id)
	if err != nil {
		return err
	}

	if todo.Completed {
		return fmt.Errorf("todo with ID %d is already completed", id)
	}

	todo.Completed = true
	now := time.Now()
	todo.CompletedAt = &now

	return s.save()
}

// Incomplete marks a todo as not completed.
func (s *Service) Incomplete(id int) error {
	todo, err := s.GetByID(id)
	if err != nil {
		return err
	}

	if !todo.Completed {
		return fmt.Errorf("todo with ID %d is already incomplete", id)
	}

	todo.Completed = false
	todo.CompletedAt = nil

	return s.save()
}

// Delete removes a todo by ID.
func (s *Service) Delete(id int) error {
	for i, todo := range s.todos {
		if todo.ID == id {
			s.todos = append(s.todos[:i], s.todos[i+1:]...)
			return s.save()
		}
	}
	return fmt.Errorf("todo with ID %d not found", id)
}

// GetStats returns statistics about todos.
func (s *Service) GetStats() Stats {
	stats := Stats{
		Total: len(s.todos),
	}

	for _, todo := range s.todos {
		if todo.Completed {
			stats.Completed++
		} else {
			stats.Pending++
		}
	}

	return stats
}

func (s *Service) save() error {
	return s.repo.Save(s.todos)
}

func (s *Service) loadTodos() error {
	todos, err := s.repo.Load()
	if err != nil {
		return err
	}

	s.todos = todos

	// Set nextID to the highest ID + 1.
	for _, todo := range s.todos {
		if todo.ID >= s.nextID {
			s.nextID = todo.ID + 1
		}
	}

	return nil
}
