package storage

import (
	"github.com/jackrudenko/pomodorocli/model"
)

// TaskStorage defines the interface for task persistence
type TaskStorage interface {
	// Save persists all tasks and returns any error
	Save(tasks []model.Task) error

	// Load retrieves all tasks
	Load() ([]model.Task, error)
}
