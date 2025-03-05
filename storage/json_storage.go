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
	Tasks    []model.Task   `json:"tasks"`
	Settings model.Settings `json:"settings"`
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

	// Read existing data to preserve settings
	existingData, err := j.readData()
	if err != nil {
		// If we couldn't read existing data, create new data with default settings
		existingData = TaskData{
			Settings: model.DefaultSettings(),
		}
	}

	// Update tasks
	existingData.Tasks = tasks

	// Save to file
	return j.writeData(existingData)
}

// Load retrieves tasks from a JSON file
func (j *JSONTaskStorage) Load() ([]model.Task, error) {
	// Read the data
	data, err := j.readData()
	if err != nil {
		// Return empty task list if there was an error reading the file
		return make([]model.Task, 0), nil
	}

	// Return the loaded data
	return data.Tasks, nil
}

// SaveSettings persists settings to the JSON file
func (j *JSONTaskStorage) SaveSettings(settings model.Settings) error {
	// Read existing data to preserve tasks
	existingData, err := j.readData()
	if err != nil {
		// If we couldn't read existing data, create new data with empty tasks
		existingData = TaskData{
			Tasks: make([]model.Task, 0),
		}
	}

	// Update settings
	existingData.Settings = settings

	// Save to file
	return j.writeData(existingData)
}

// LoadSettings retrieves settings from the JSON file
func (j *JSONTaskStorage) LoadSettings() (model.Settings, error) {
	// Read the data
	data, err := j.readData()
	if err != nil {
		// Return default settings if there was an error reading the file
		return model.DefaultSettings(), nil
	}

	// If settings is empty (old file format), return default settings
	if (data.Settings == model.Settings{}) {
		return model.DefaultSettings(), nil
	}

	// Return the loaded settings
	return data.Settings, nil
}

// readData reads the JSON file and returns the parsed data
func (j *JSONTaskStorage) readData() (TaskData, error) {
	// Check if file exists
	if _, err := os.Stat(j.filePath); os.IsNotExist(err) {
		// Return empty data if file doesn't exist yet
		return TaskData{
			Tasks:    make([]model.Task, 0),
			Settings: model.DefaultSettings(),
		}, nil
	}

	// Read file
	fileData, err := os.ReadFile(j.filePath)
	if err != nil {
		return TaskData{}, err
	}

	// Unmarshal JSON
	var data TaskData
	if err := json.Unmarshal(fileData, &data); err != nil {
		return TaskData{}, err
	}

	return data, nil
}

// writeData writes data to the JSON file
func (j *JSONTaskStorage) writeData(data TaskData) error {
	// Marshal to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(j.filePath, jsonData, 0o644)
}
