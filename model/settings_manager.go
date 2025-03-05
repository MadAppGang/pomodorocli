package model

// SettingsManager handles the application settings
type SettingsManager struct {
	Settings Settings
	OnChange func()
}

// NewSettingsManager creates a new settings manager with default settings
func NewSettingsManager() *SettingsManager {
	return &SettingsManager{
		Settings: DefaultSettings(),
		OnChange: nil,
	}
}

// SetPomodoroDuration sets the pomodoro duration in minutes
func (sm *SettingsManager) SetPomodoroDuration(minutes int) {
	if minutes < 1 {
		minutes = 1 // Minimum 1 minute
	}
	sm.Settings.PomodoroDuration = minutes
	sm.notifyChange()
}

// SetShortBreakDuration sets the short break duration in minutes
func (sm *SettingsManager) SetShortBreakDuration(minutes int) {
	if minutes < 1 {
		minutes = 1 // Minimum 1 minute
	}
	sm.Settings.ShortBreakDuration = minutes
	sm.notifyChange()
}

// SetLongBreakDuration sets the long break duration in minutes
func (sm *SettingsManager) SetLongBreakDuration(minutes int) {
	if minutes < 1 {
		minutes = 1 // Minimum 1 minute
	}
	sm.Settings.LongBreakDuration = minutes
	sm.notifyChange()
}

// RegisterChangeHandler sets a function to be called when settings change
func (sm *SettingsManager) RegisterChangeHandler(handler func()) {
	sm.OnChange = handler
}

// notifyChange calls the change handler if one is registered
func (sm *SettingsManager) notifyChange() {
	if sm.OnChange != nil {
		sm.OnChange()
	}
}
