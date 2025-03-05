package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/jackrudenko/pomodorocli/model"
)

// TaskData represents the data structure stored in the JSON file
type TaskData struct {
	Tasks []model.Task `json:"tasks"`
}

// JSONTaskStorage implements TaskStorage using a local JSON file
type JSONTaskStorage struct {
	filePath string
}

// NewJSONTaskStorage creates a new JSONTaskStorage instance
func NewJSONTaskStorage(filePath string) (*JSONTaskStorage, error) {
	// Ensure the directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	return &JSONTaskStorage{
		filePath: filePath,
	}, nil
}

// Save persists tasks to a JSON file
func (j *JSONTaskStorage) Save(tasks []model.Task) error {
	if tasks == nil {
		return errors.New("tasks cannot be nil")
	}

	// Create the data structure
	data := TaskData{
		Tasks: tasks,
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(j.filePath, jsonData, 0o644)
}

// Load retrieves tasks from a JSON file
func (j *JSONTaskStorage) Load() ([]model.Task, error) {
	// Check if file exists
	if _, err := os.Stat(j.filePath); os.IsNotExist(err) {
		// Return empty task list if file doesn't exist yet
		return make([]model.Task, 0), nil
	}

	// Read file
	fileData, err := os.ReadFile(j.filePath)
	if err != nil {
		return nil, err
	}

	// Unmarshal JSON
	var data TaskData
	if err := json.Unmarshal(fileData, &data); err != nil {
		return nil, err
	}

	// Return the loaded data
	return data.Tasks, nil
}
