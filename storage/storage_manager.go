package storage

import (
	"github.com/jackrudenko/pomodorocli/model"
)

// StorageManager handles loading and saving tasks and settings using the provided storage
type StorageManager struct {
	storage         TaskStorage
	settingsStorage SettingsStorage
	taskManager     *model.TaskManager
	settings        *model.Settings
}

// NewStorageManager creates a new StorageManager
func NewStorageManager(storage TaskStorage, settingsStorage SettingsStorage, taskManager *model.TaskManager, settings *model.Settings) *StorageManager {
	return &StorageManager{
		storage:         storage,
		settingsStorage: settingsStorage,
		taskManager:     taskManager,
		settings:        settings,
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

// LoadSettings loads settings from storage
func (sm *StorageManager) LoadSettings() error {
	settings, err := sm.settingsStorage.LoadSettings()
	if err != nil {
		return err
	}

	*sm.settings = settings
	return nil
}

// SaveSettings saves settings to storage
func (sm *StorageManager) SaveSettings() error {
	return sm.settingsStorage.SaveSettings(*sm.settings)
}

// AutoSave returns a function that can be called after task operations to autosave
func (sm *StorageManager) AutoSave() func() {
	return func() {
		// Ignore errors during autosave
		_ = sm.SaveTasks()
	}
}

// AutoSaveSettings returns a function that can be called after settings operations to autosave
func (sm *StorageManager) AutoSaveSettings() func() {
	return func() {
		// Ignore errors during autosave
		_ = sm.SaveSettings()
	}
}
