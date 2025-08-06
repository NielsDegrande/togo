package todo

import (
	"testing"
)

// MockRepository is a mock implementation of Repository for testing.
type MockRepository struct {
	todos []Todo
	err   error
}

func (m *MockRepository) Save(todos []Todo) error {
	if m.err != nil {
		return m.err
	}
	m.todos = make([]Todo, len(todos))
	copy(m.todos, todos)
	return nil
}

func (m *MockRepository) Load() ([]Todo, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make([]Todo, len(m.todos))
	copy(result, m.todos)
	return result, nil
}

func TestNewService(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)

	if service == nil {
		t.Fatal("NewService returned nil")
	}

	if service.repo != repo {
		t.Error("Service repository not set correctly")
	}

	if service.nextID != 1 {
		t.Errorf("Expected nextID to be 1, got %d", service.nextID)
	}

	if len(service.todos) != 0 {
		t.Errorf("Expected empty todos slice, got %d items", len(service.todos))
	}
}

func TestService_Add(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)

	tests := []struct {
		name        string
		description string
		wantErr     bool
	}{
		{
			name:        "valid todo",
			description: "Test todo description",
			wantErr:     false,
		},
		{
			name:        "empty description",
			description: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todoItem, err := service.Add(tt.description)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if todoItem == nil {
				t.Error("Expected todo but got nil")
				return
			}

			if todoItem.Description != tt.description {
				t.Errorf("Expected description %s, got %s", tt.description, todoItem.Description)
			}

			if todoItem.Completed {
				t.Error("New todo should not be completed")
			}

			if todoItem.ID != 1 {
				t.Errorf("Expected ID 1, got %d", todoItem.ID)
			}
		})
	}
}

func TestService_GetByID(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)

	// Add a todo first.
	addedTodo, err := service.Add("Test todo")
	if err != nil {
		t.Fatalf("Failed to add todo: %v", err)
	}

	// Test getting the todo by ID.
	todoItem, err := service.GetByID(addedTodo.ID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if todoItem.ID != addedTodo.ID {
		t.Errorf("Expected ID %d, got %d", addedTodo.ID, todoItem.ID)
	}

	// Test getting non-existent todo.
	_, err = service.GetByID(999)
	if err == nil {
		t.Error("Expected error for non-existent todo")
	}
}

func TestService_Complete(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)

	// Add a todo first.
	addedTodo, err := service.Add("Test todo")
	if err != nil {
		t.Fatalf("Failed to add todo: %v", err)
	}

	// Complete the todo.
	err = service.Complete(addedTodo.ID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check if the todo is completed.
	todoItem, err := service.GetByID(addedTodo.ID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !todoItem.Completed {
		t.Error("Todo should be completed")
	}

	if todoItem.CompletedAt == nil {
		t.Error("CompletedAt should be set")
	}

	// Test completing already completed todo.
	err = service.Complete(addedTodo.ID)
	if err == nil {
		t.Error("Expected error when completing already completed todo")
	}
}

func TestService_Incomplete(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)

	// Add and complete a todo first.
	addedTodo, err := service.Add("Test todo")
	if err != nil {
		t.Fatalf("Failed to add todo: %v", err)
	}

	err = service.Complete(addedTodo.ID)
	if err != nil {
		t.Fatalf("Failed to complete todo: %v", err)
	}

	// Incomplete the todo.
	err = service.Incomplete(addedTodo.ID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check if the todo is incomplete.
	todoItem, err := service.GetByID(addedTodo.ID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if todoItem.Completed {
		t.Error("Todo should be incomplete")
	}

	if todoItem.CompletedAt != nil {
		t.Error("CompletedAt should be nil")
	}

	// Test marking already incomplete todo as incomplete.
	err = service.Incomplete(addedTodo.ID)
	if err == nil {
		t.Error("Expected error when marking already incomplete todo as incomplete")
	}
}

func TestService_Delete(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)

	// Add a todo first.
	addedTodo, err := service.Add("Test todo")
	if err != nil {
		t.Fatalf("Failed to add todo: %v", err)
	}

	// Delete the todo.
	err = service.Delete(addedTodo.ID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check if the todo is deleted.
	_, err = service.GetByID(addedTodo.ID)
	if err == nil {
		t.Error("Expected error for deleted todo")
	}

	// Test deleting non-existent todo.
	err = service.Delete(999)
	if err == nil {
		t.Error("Expected error for non-existent todo")
	}
}

func TestService_GetStats(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)

	// Test with no todos.
	stats := service.GetStats()
	if stats.Total != 0 || stats.Completed != 0 || stats.Pending != 0 {
		t.Error("Expected all stats to be 0 for empty service")
	}

	// Add some todos.
	_, err := service.Add("Todo 1")
	if err != nil {
		t.Fatalf("Failed to add todo: %v", err)
	}

	todo2, err := service.Add("Todo 2")
	if err != nil {
		t.Fatalf("Failed to add todo: %v", err)
	}

	_, err = service.Add("Todo 3")
	if err != nil {
		t.Fatalf("Failed to add todo: %v", err)
	}

	// Complete one todo.
	err = service.Complete(todo2.ID)
	if err != nil {
		t.Fatalf("Failed to complete todo: %v", err)
	}

	// Check stats.
	stats = service.GetStats()
	if stats.Total != 3 {
		t.Errorf("Expected total 3, got %d", stats.Total)
	}
	if stats.Completed != 1 {
		t.Errorf("Expected completed 1, got %d", stats.Completed)
	}
	if stats.Pending != 2 {
		t.Errorf("Expected pending 2, got %d", stats.Pending)
	}
}

func TestStats_CompletionRate(t *testing.T) {
	tests := []struct {
		name     string
		stats    Stats
		expected float64
	}{
		{
			name:     "no todos",
			stats:    Stats{Total: 0, Completed: 0, Pending: 0},
			expected: 0,
		},
		{
			name:     "all completed",
			stats:    Stats{Total: 5, Completed: 5, Pending: 0},
			expected: 100,
		},
		{
			name:     "half completed",
			stats:    Stats{Total: 4, Completed: 2, Pending: 2},
			expected: 50,
		},
		{
			name:     "none completed",
			stats:    Stats{Total: 3, Completed: 0, Pending: 3},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.stats.CompletionRate()
			if result != tt.expected {
				t.Errorf("Expected completion rate %.1f, got %.1f", tt.expected, result)
			}
		})
	}
}

func TestService_GetByStatus(t *testing.T) {
	repo := &MockRepository{}
	service := NewService(repo)

	// Add some todos.
	todo1, err := service.Add("Todo 1")
	if err != nil {
		t.Fatalf("Failed to add todo: %v", err)
	}

	todo2, err := service.Add("Todo 2")
	if err != nil {
		t.Fatalf("Failed to add todo: %v", err)
	}

	todo3, err := service.Add("Todo 3")
	if err != nil {
		t.Fatalf("Failed to add todo: %v", err)
	}

	// Complete some todos.
	err = service.Complete(todo1.ID)
	if err != nil {
		t.Fatalf("Failed to complete todo: %v", err)
	}

	err = service.Complete(todo3.ID)
	if err != nil {
		t.Fatalf("Failed to complete todo: %v", err)
	}

	// Test getting completed todos.
	completed := service.GetByStatus(true)
	if len(completed) != 2 {
		t.Errorf("Expected 2 completed todos, got %d", len(completed))
	}

	// Test getting pending todos.
	pending := service.GetByStatus(false)
	if len(pending) != 1 {
		t.Errorf("Expected 1 pending todo, got %d", len(pending))
	}

	if pending[0].ID != todo2.ID {
		t.Errorf("Expected pending todo ID %d, got %d", todo2.ID, pending[0].ID)
	}
}
