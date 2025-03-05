package model

import (
	"fmt"
	"time"
)

// TimerState represents the current state of the timer
type TimerState int

const (
	// TimerStopped means the timer is not running
	TimerStopped TimerState = iota
	// TimerRunning means the timer is actively counting down
	TimerRunning
	// TimerPaused means the timer has been temporarily paused
	TimerPaused
)

// TimerMode represents different timer modes
type TimerMode int

const (
	// FocusMode is for concentrated work
	FocusMode TimerMode = iota
	// ShortBreakMode is for short breaks between pomodoros
	ShortBreakMode
	// LongBreakMode is for longer breaks after a cycle of pomodoros
	LongBreakMode
)

// Default pomodoros per cycle
const (
	DefaultPomodorosPerCycle = 4
)

// Timer represents a pomodoro timer
type Timer struct {
	// Current state of the timer (running, paused, stopped)
	State TimerState
	// Current mode (focus, short break, long break)
	Mode TimerMode
	// Time remaining in the current timer
	Remaining time.Duration
	// When the timer was started
	StartTime time.Time
	// The original duration for this timer
	Duration time.Duration
	// Number of completed pomodoros in the current cycle
	CompletedPomodoros int
	// Maximum number of pomodoros before a long break
	PomodorosPerCycle int
	// The current active task (nil if none)
	CurrentTask *Task
	// Settings for timer durations
	Settings *Settings
}

// NewTimer creates a new timer with default settings
func NewTimer() *Timer {
	// Create with default settings
	settings := DefaultSettings()

	return &Timer{
		State:              TimerStopped,
		Mode:               FocusMode,
		Remaining:          settings.GetPomodoroDuration(),
		Duration:           settings.GetPomodoroDuration(),
		CompletedPomodoros: 0,
		PomodorosPerCycle:  DefaultPomodorosPerCycle,
		CurrentTask:        nil,
		Settings:           &settings,
	}
}

// SetSettings updates the timer's settings
func (t *Timer) SetSettings(settings *Settings) {
	t.Settings = settings
	// Always update the duration to reflect the new settings
	t.updateDurationFromSettings()
}

// updateDurationFromSettings updates the timer duration based on current mode and settings
func (t *Timer) updateDurationFromSettings() {
	// Update the mode duration
	switch t.Mode {
	case FocusMode:
		t.Duration = t.Settings.GetPomodoroDuration()
	case ShortBreakMode:
		t.Duration = t.Settings.GetShortBreakDuration()
	case LongBreakMode:
		t.Duration = t.Settings.GetLongBreakDuration()
	}

	// Only reset the remaining time if the timer is stopped
	if t.State == TimerStopped {
		t.Remaining = t.Duration
	}
}

// Start starts the timer
func (t *Timer) Start() {
	// If the timer is already paused, resume it instead of resetting
	if t.State == TimerPaused {
		// Use Resume logic
		t.Resume()
		return
	}

	// Otherwise, start a fresh timer
	t.State = TimerRunning
	t.StartTime = time.Now()
	t.updateDurationFromSettings()
}

// Stop stops the timer
func (t *Timer) Stop() {
	t.State = TimerStopped
	// Reset to initial duration based on current mode
	t.updateDurationFromSettings()
}

// Reset resets the timer to its initial state for the current mode
func (t *Timer) Reset() {
	t.State = TimerStopped
	t.updateDurationFromSettings()
}

// Pause pauses the timer
func (t *Timer) Pause() {
	if t.State == TimerRunning {
		t.State = TimerPaused
		// Calculate remaining time
		elapsed := time.Since(t.StartTime)
		t.Remaining = t.Duration - elapsed
		if t.Remaining < 0 {
			t.Remaining = 0
		}
	}
}

// Resume resumes the timer from a paused state
func (t *Timer) Resume() {
	if t.State == TimerPaused {
		t.State = TimerRunning
		t.StartTime = time.Now().Add(-t.Duration + t.Remaining)
	}
}

// SetCurrentTask sets the current task
func (t *Timer) SetCurrentTask(task Task) {
	t.CurrentTask = &task
}

// Update updates the timer's state and returns true if the timer completed
func (t *Timer) Update() bool {
	if t.State != TimerRunning {
		return false
	}

	// Calculate remaining time
	elapsed := time.Since(t.StartTime)
	t.Remaining = t.Duration - elapsed

	// Check if timer has finished
	if t.Remaining <= 0 {
		t.Remaining = 0
		t.State = TimerStopped

		// If we were in focus mode, increment completed pomodoros
		if t.Mode == FocusMode {
			t.CompletedPomodoros++

			// Update current task if one is set
			if t.CurrentTask != nil {
				t.CurrentTask.AddCompletedPomodoro()
				t.CurrentTask.AddTimeSpent(t.Duration)
			}
		}

		// Advance to the next timer mode
		t.advanceTimerMode()

		return true // Timer completed
	}

	return false // Timer still running
}

// advanceTimerMode moves to the next timer mode based on the completed pomodoros
func (t *Timer) advanceTimerMode() {
	switch t.Mode {
	case FocusMode:
		// After focus mode, decide if we need a short or long break
		if t.CompletedPomodoros%t.PomodorosPerCycle == 0 {
			// Long break after completing a cycle
			t.Mode = LongBreakMode
			t.Duration = t.Settings.GetLongBreakDuration()
		} else {
			// Short break after individual pomodoro
			t.Mode = ShortBreakMode
			t.Duration = t.Settings.GetShortBreakDuration()
		}
	case ShortBreakMode, LongBreakMode:
		// After any break, go back to focus mode
		t.Mode = FocusMode
		t.Duration = t.Settings.GetPomodoroDuration()
	}

	t.Remaining = t.Duration
}

// FormatTime formats the remaining time as mm:ss
func (t *Timer) FormatTime() string {
	minutes := int(t.Remaining.Minutes())
	seconds := int(t.Remaining.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// ProgressPercentage returns the percentage of time elapsed
func (t *Timer) ProgressPercentage() float64 {
	if t.Duration <= 0 {
		return 0
	}
	return 100.0 * (1.0 - float64(t.Remaining)/float64(t.Duration))
}
