package model

import (
	"fmt"
	"time"
)

// TimerState represents the current state of the timer
type TimerState int

const (
	// TimerStopped indicates the timer is not running
	TimerStopped TimerState = iota
	// TimerRunning indicates the timer is active
	TimerRunning
	// TimerPaused indicates the timer is paused
	TimerPaused
)

// TimerMode represents the current mode of the pomodoro timer
type TimerMode int

const (
	// FocusMode is the standard pomodoro work period
	FocusMode TimerMode = iota
	// ShortBreakMode is the short break between pomodoros
	ShortBreakMode
	// LongBreakMode is the longer break after several pomodoros
	LongBreakMode
)

// Default durations for different timer modes (in minutes)
const (
	DefaultFocusDuration      = 25 * time.Minute
	DefaultShortBreakDuration = 5 * time.Minute
	DefaultLongBreakDuration  = 15 * time.Minute
	// Number of pomodoros before a long break
	DefaultPomodorosPerCycle = 4
)

// Timer represents the Pomodoro timer
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
}

// NewTimer creates a new Pomodoro timer
func NewTimer() *Timer {
	return &Timer{
		State:              TimerStopped,
		Mode:               FocusMode,
		Remaining:          DefaultFocusDuration,
		Duration:           DefaultFocusDuration,
		CompletedPomodoros: 0,
		PomodorosPerCycle:  DefaultPomodorosPerCycle,
		CurrentTask:        nil,
	}
}

// Start begins the timer
func (t *Timer) Start() {
	if t.State != TimerRunning {
		t.State = TimerRunning
		t.StartTime = time.Now()
	}
}

// Stop stops the timer
func (t *Timer) Stop() {
	if t.State == TimerRunning {
		t.State = TimerStopped
		// If the timer was for a task, add the elapsed time to the task
		if t.CurrentTask != nil {
			elapsed := t.Duration - t.Remaining
			t.CurrentTask.AddTimeSpent(elapsed)
		}
	}
}

// Pause pauses the timer
func (t *Timer) Pause() {
	if t.State == TimerRunning {
		t.State = TimerPaused
		// Update remaining time
		elapsed := time.Since(t.StartTime)
		t.Remaining -= elapsed
		if t.Remaining < 0 {
			t.Remaining = 0
		}
	}
}

// Resume resumes a paused timer
func (t *Timer) Resume() {
	if t.State == TimerPaused {
		t.State = TimerRunning
		t.StartTime = time.Now()
	}
}

// Reset resets the timer to its initial state
func (t *Timer) Reset() {
	t.State = TimerStopped
	t.Remaining = t.Duration
}

// SetCurrentTask sets the current active task
func (t *Timer) SetCurrentTask(task *Task) {
	t.CurrentTask = task
}

// Update updates the timer state, should be called regularly
func (t *Timer) Update() bool {
	// Only update time if the timer is running
	if t.State == TimerRunning {
		elapsed := time.Since(t.StartTime)
		t.Remaining = t.Duration - elapsed

		// Check if timer has completed
		if t.Remaining <= 0 {
			t.Remaining = 0
			t.State = TimerStopped

			// Handle timer completion based on mode
			if t.Mode == FocusMode && t.CurrentTask != nil {
				t.CurrentTask.AddCompletedPomodoro()
				t.CompletedPomodoros++
			}

			// Determine next timer mode
			t.advanceTimerMode()
			return true // Timer completed
		}
	}
	return false // Timer not completed
}

// advanceTimerMode advances to the next timer mode
func (t *Timer) advanceTimerMode() {
	switch t.Mode {
	case FocusMode:
		// After focus period, determine if we need short or long break
		if t.CompletedPomodoros%t.PomodorosPerCycle == 0 {
			t.Mode = LongBreakMode
			t.Duration = DefaultLongBreakDuration
		} else {
			t.Mode = ShortBreakMode
			t.Duration = DefaultShortBreakDuration
		}
	case ShortBreakMode, LongBreakMode:
		// After any break, go back to focus mode
		t.Mode = FocusMode
		t.Duration = DefaultFocusDuration
	}
	t.Remaining = t.Duration
}

// FormatTime returns a formatted string of the remaining time
func (t *Timer) FormatTime() string {
	minutes := int(t.Remaining.Minutes())
	seconds := int(t.Remaining.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// ProgressPercentage returns the percentage of timer completed (0-100)
func (t *Timer) ProgressPercentage() float64 {
	if t.Duration == 0 {
		return 0
	}
	return 100.0 * (1.0 - (float64(t.Remaining) / float64(t.Duration)))
}
