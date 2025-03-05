package model

import (
	"time"
)

// Settings represents the application settings
type Settings struct {
	// Pomodoro session duration in minutes
	PomodoroDuration int `json:"pomodoro_duration"`
	// Short break duration in minutes
	ShortBreakDuration int `json:"short_break_duration"`
	// Long break duration in minutes
	LongBreakDuration int `json:"long_break_duration"`
	// Automatically start breaks after pomodoro completes
	AutoStartBreaks bool `json:"auto_start_breaks"`
}

// DefaultSettings creates and returns default settings
func DefaultSettings() Settings {
	return Settings{
		PomodoroDuration:   25,    // Default: 25 minutes
		ShortBreakDuration: 5,     // Default: 5 minutes
		LongBreakDuration:  30,    // Default: 30 minutes
		AutoStartBreaks:    false, // Default: don't auto-start breaks
	}
}

// GetPomodoroDuration returns the pomodoro duration as time.Duration
func (s Settings) GetPomodoroDuration() time.Duration {
	return time.Duration(s.PomodoroDuration) * time.Minute
}

// GetShortBreakDuration returns the short break duration as time.Duration
func (s Settings) GetShortBreakDuration() time.Duration {
	return time.Duration(s.ShortBreakDuration) * time.Minute
}

// GetLongBreakDuration returns the long break duration as time.Duration
func (s Settings) GetLongBreakDuration() time.Duration {
	return time.Duration(s.LongBreakDuration) * time.Minute
}
