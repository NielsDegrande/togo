package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"example.com/todo/internal/todo"
)

func TestNewJSONRepository(t *testing.T) {
	filename := "test.json"
	repo := NewJSONRepository(filename)

	if repo == nil {
		t.Fatal("NewJSONRepository returned nil")
	}

	if repo.filename != filename {
		t.Errorf("Expected filename %s, got %s", filename, repo.filename)
	}
}

func TestJSONRepository_Save(t *testing.T) {
	tests := []struct {
		name    string
		todos   []todo.Todo
		wantErr bool
	}{
		{
			name:    "empty todos",
			todos:   []todo.Todo{},
			wantErr: false,
		},
		{
			name: "single todo",
			todos: []todo.Todo{
				{
					ID:          1,
					Description: "Test todo",
					Completed:   false,
					CreatedAt:   time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				},
			},
			wantErr: false,
		},
		{
			name: "multiple todos",
			todos: []todo.Todo{
				{
					ID:          1,
					Description: "First todo",
					Completed:   false,
					CreatedAt:   time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				},
				{
					ID:          2,
					Description: "Second todo",
					Completed:   true,
					CreatedAt:   time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
					CompletedAt: func() *time.Time {
						t := time.Date(2023, 1, 2, 13, 0, 0, 0, time.UTC)
						return &t
					}(),
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filename := filepath.Join(tmpDir, "test.json")
			repo := NewJSONRepository(filename)

			err := repo.Save(tt.todos)

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

			// Verify file was created and contains expected data.
			data, err := os.ReadFile(filename)
			if err != nil {
				t.Errorf("Failed to read saved file: %v", err)
				return
			}

			var savedTodos []todo.Todo
			if err := json.Unmarshal(data, &savedTodos); err != nil {
				t.Errorf("Failed to unmarshal saved data: %v", err)
				return
			}

			if len(savedTodos) != len(tt.todos) {
				t.Errorf("Expected %d todos, got %d", len(tt.todos), len(savedTodos))
				return
			}

			for i, expected := range tt.todos {
				if savedTodos[i].ID != expected.ID {
					t.Errorf("Todo %d: expected ID %d, got %d", i, expected.ID, savedTodos[i].ID)
				}
				if savedTodos[i].Description != expected.Description {
					t.Errorf("Todo %d: expected description %s, got %s", i, expected.Description, savedTodos[i].Description)
				}
				if savedTodos[i].Completed != expected.Completed {
					t.Errorf("Todo %d: expected completed %t, got %t", i, expected.Completed, savedTodos[i].Completed)
				}
			}
		})
	}
}

func TestJSONRepository_Save_FilePermissionError(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "readonly_dir")

	if err := os.Mkdir(filename, 0o444); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	repo := NewJSONRepository(filename)
	todos := []todo.Todo{
		{
			ID:          1,
			Description: "Test todo",
			Completed:   false,
			CreatedAt:   time.Now(),
		},
	}

	err := repo.Save(todos)
	if err == nil {
		t.Error("Expected error when writing to directory, but got none")
	}
}

func TestJSONRepository_Load(t *testing.T) {
	tests := []struct {
		name        string
		setupFile   func(filename string) error
		expectedLen int
		wantErr     bool
	}{
		{
			name: "non-existent file",
			setupFile: func(filename string) error {
				// Do not create file.
				return nil
			},
			expectedLen: 0,
			wantErr:     false,
		},
		{
			name: "empty file",
			setupFile: func(filename string) error {
				return os.WriteFile(filename, []byte{}, 0o600)
			},
			expectedLen: 0,
			wantErr:     false,
		},
		{
			name: "valid empty json array",
			setupFile: func(filename string) error {
				return os.WriteFile(filename, []byte("[]"), 0o600)
			},
			expectedLen: 0,
			wantErr:     false,
		},
		{
			name: "valid json with one todo",
			setupFile: func(filename string) error {
				todos := []todo.Todo{
					{
						ID:          1,
						Description: "Test todo",
						Completed:   false,
						CreatedAt:   time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					},
				}
				data, err := json.MarshalIndent(todos, "", "  ")
				if err != nil {
					return err
				}
				return os.WriteFile(filename, data, 0o600)
			},
			expectedLen: 1,
			wantErr:     false,
		},
		{
			name: "valid json with multiple todos",
			setupFile: func(filename string) error {
				todos := []todo.Todo{
					{
						ID:          1,
						Description: "First todo",
						Completed:   false,
						CreatedAt:   time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					},
					{
						ID:          2,
						Description: "Second todo",
						Completed:   true,
						CreatedAt:   time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						CompletedAt: func() *time.Time {
							t := time.Date(2023, 1, 2, 13, 0, 0, 0, time.UTC)
							return &t
						}(),
					},
				}
				data, err := json.MarshalIndent(todos, "", "  ")
				if err != nil {
					return err
				}
				return os.WriteFile(filename, data, 0o600)
			},
			expectedLen: 2,
			wantErr:     false,
		},
		{
			name: "invalid json",
			setupFile: func(filename string) error {
				return os.WriteFile(filename, []byte("invalid json"), 0o600)
			},
			expectedLen: 0,
			wantErr:     true,
		},
		{
			name: "malformed json object",
			setupFile: func(filename string) error {
				return os.WriteFile(filename, []byte(`{"invalid": "json"`), 0o600)
			},
			expectedLen: 0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filename := filepath.Join(tmpDir, "test.json")

			if err := tt.setupFile(filename); err != nil {
				t.Fatalf("Failed to setup test file: %v", err)
			}

			repo := NewJSONRepository(filename)
			todos, err := repo.Load()

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

			if len(todos) != tt.expectedLen {
				t.Errorf("Expected %d todos, got %d", tt.expectedLen, len(todos))
			}

			// Verify todos are properly structured if we expect any.
			if tt.expectedLen > 0 {
				for i, todoItem := range todos {
					if todoItem.ID == 0 {
						t.Errorf("Todo %d: ID should not be zero", i)
					}
					if todoItem.Description == "" {
						t.Errorf("Todo %d: Description should not be empty", i)
					}
					if todoItem.CreatedAt.IsZero() {
						t.Errorf("Todo %d: CreatedAt should not be zero", i)
					}
				}
			}
		})
	}
}

func TestJSONRepository_SaveAndLoad_Integration(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "integration_test.json")
	repo := NewJSONRepository(filename)

	// Test data.
	originalTodos := []todo.Todo{
		{
			ID:          1,
			Description: "First todo",
			Completed:   false,
			CreatedAt:   time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			ID:          2,
			Description: "Second todo",
			Completed:   true,
			CreatedAt:   time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
			CompletedAt: func() *time.Time {
				t := time.Date(2023, 1, 2, 13, 0, 0, 0, time.UTC)
				return &t
			}(),
		},
	}

	// Save todos.
	err := repo.Save(originalTodos)
	if err != nil {
		t.Fatalf("Failed to save todos: %v", err)
	}

	// Load todos.
	loadedTodos, err := repo.Load()
	if err != nil {
		t.Fatalf("Failed to load todos: %v", err)
	}

	// Verify loaded todos match original.
	if len(loadedTodos) != len(originalTodos) {
		t.Errorf("Expected %d todos, got %d", len(originalTodos), len(loadedTodos))
	}

	for i, original := range originalTodos {
		loaded := loadedTodos[i]

		if loaded.ID != original.ID {
			t.Errorf("Todo %d: expected ID %d, got %d", i, original.ID, loaded.ID)
		}
		if loaded.Description != original.Description {
			t.Errorf("Todo %d: expected description %s, got %s", i, original.Description, loaded.Description)
		}
		if loaded.Completed != original.Completed {
			t.Errorf("Todo %d: expected completed %t, got %t", i, original.Completed, loaded.Completed)
		}
		if !loaded.CreatedAt.Equal(original.CreatedAt) {
			t.Errorf("Todo %d: expected CreatedAt %v, got %v", i, original.CreatedAt, loaded.CreatedAt)
		}

		// Check CompletedAt.
		if original.CompletedAt == nil && loaded.CompletedAt != nil {
			t.Errorf("Todo %d: expected CompletedAt to be nil, got %v", i, loaded.CompletedAt)
		}
		if original.CompletedAt != nil && loaded.CompletedAt == nil {
			t.Errorf("Todo %d: expected CompletedAt to be %v, got nil", i, original.CompletedAt)
		}
		if original.CompletedAt != nil && loaded.CompletedAt != nil {
			if !loaded.CompletedAt.Equal(*original.CompletedAt) {
				t.Errorf("Todo %d: expected CompletedAt %v, got %v", i, original.CompletedAt, loaded.CompletedAt)
			}
		}
	}
}
