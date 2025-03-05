package storage

import (
	"github.com/jackrudenko/pomodorocli/model"
)

// SettingsStorage defines the interface for settings persistence
type SettingsStorage interface {
	// SaveSettings persists settings and returns any error
	SaveSettings(settings model.Settings) error

	// LoadSettings retrieves settings
	LoadSettings() (model.Settings, error)
}
