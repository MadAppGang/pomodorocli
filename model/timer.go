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

// Default timer durations
const (
	DefaultFocusDuration      = 25 * time.Minute
	DefaultShortBreakDuration = 5 * time.Minute
	DefaultLongBreakDuration  = 15 * time.Minute
	DefaultPomodorosPerCycle  = 4
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
}

// NewTimer creates a new timer with default settings
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

// Start starts the timer
func (t *Timer) Start() {
	t.State = TimerRunning
	t.StartTime = time.Now()
}

// Stop stops the timer and resets it
func (t *Timer) Stop() {
	t.State = TimerStopped
	t.Remaining = t.Duration

	// If we were tracking time for a task, add the elapsed time
	if t.CurrentTask != nil && t.Mode == FocusMode {
		elapsed := t.Duration - t.Remaining
		if elapsed > 0 {
			// Increment the time spent on the task
			task := *t.CurrentTask
			task.AddTimeSpent(elapsed)
			*t.CurrentTask = task
		}
	}
}

// Pause pauses the timer
func (t *Timer) Pause() {
	if t.State == TimerRunning {
		t.State = TimerPaused
		elapsed := time.Since(t.StartTime)
		t.Remaining -= elapsed

		// If we were tracking time for a task, add the elapsed time
		if t.CurrentTask != nil && t.Mode == FocusMode {
			if elapsed > 0 {
				// Increment the time spent on the task
				task := *t.CurrentTask
				task.AddTimeSpent(elapsed)
				*t.CurrentTask = task
			}
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
func (t *Timer) SetCurrentTask(task Task) {
	taskPtr := &task
	t.CurrentTask = taskPtr
}

// Update updates the timer state, should be called regularly
func (t *Timer) Update() bool {
	// Only update time if the timer is running
	if t.State == TimerRunning {
		elapsed := time.Since(t.StartTime)
		t.Remaining = t.Duration - elapsed

		// Check if timer has finished
		if t.Remaining <= 0 {
			t.Remaining = 0

			// Update task if in focus mode
			if t.Mode == FocusMode && t.CurrentTask != nil {
				// Increment completed pomodoros for the task
				task := *t.CurrentTask
				task.AddCompletedPomodoro()
				task.AddTimeSpent(t.Duration)
				*t.CurrentTask = task

				// Increment the timer's completed pomodoros
				t.CompletedPomodoros++
			}

			// Advance to the next timer mode
			t.advanceTimerMode()
			return true // Timer completed
		}
	}

	return false // Timer not completed
}

// advanceTimerMode advances to the next timer mode after completion
func (t *Timer) advanceTimerMode() {
	switch t.Mode {
	case FocusMode:
		// After focus time, check if we need a long or short break
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

	// Reset timer for the new mode
	t.Reset()
	t.State = TimerStopped
}

// FormatTime formats the remaining time as MM:SS
func (t *Timer) FormatTime() string {
	minutes := int(t.Remaining.Minutes())
	seconds := int(t.Remaining.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// ProgressPercentage returns the progress percentage (0.0-1.0)
func (t *Timer) ProgressPercentage() float64 {
	return 1.0 - float64(t.Remaining)/float64(t.Duration)
}
