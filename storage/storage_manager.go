package storage

import (
	"github.com/jackrudenko/pomodorocli/model"
)

// StorageManager handles loading and saving tasks using the provided storage
type StorageManager struct {
	storage     TaskStorage
	taskManager *model.TaskManager
}

// NewStorageManager creates a new StorageManager
func NewStorageManager(storage TaskStorage, taskManager *model.TaskManager) *StorageManager {
	return &StorageManager{
		storage:     storage,
		taskManager: taskManager,
	}
}

// LoadTasks loads tasks from storage into the task manager
func (sm *StorageManager) LoadTasks() error {
	tasks, err := sm.storage.Load()
	if err != nil {
		return err
	}

	sm.taskManager.LoadTasks(tasks)
	return nil
}

// SaveTasks saves tasks from the task manager to storage
func (sm *StorageManager) SaveTasks() error {
	tasks := sm.taskManager.GetTasks()
	return sm.storage.Save(tasks)
}

// AutoSave returns a function that can be called after task operations to autosave
func (sm *StorageManager) AutoSave() func() {
	return func() {
		// Ignore errors during autosave
		_ = sm.SaveTasks()
	}
}
