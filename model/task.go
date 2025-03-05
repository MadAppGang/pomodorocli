package model

import (
	"fmt"
	"time"

	"github.com/segmentio/ksuid"
)

// Task represents a single task in the Pomodoro timer
type Task struct {
	ID          string    `json:"id"` // Now a string to store the KSUID
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	Completed   bool      `json:"completed"`
	// Total number of pomodoros planned for this task
	PlannedPomodoros int `json:"planned_pomodoros"`
	// Number of pomodoros completed for this task
	CompletedPomodoros int `json:"completed_pomodoros"`
	// Total time spent on this task
	TimeSpent time.Duration `json:"time_spent"`
}

// NewTask creates a new task with default values
func NewTask(description string, plannedPomodoros int) Task {
	// Generate a new KSUID for the task
	id := ksuid.New().String()

	return Task{
		ID:                 id,
		Description:        description,
		CreatedAt:          time.Now(),
		Completed:          false,
		PlannedPomodoros:   plannedPomodoros,
		CompletedPomodoros: 0,
		TimeSpent:          0,
	}
}

// ToggleComplete toggles the completed status of the task
func (t *Task) ToggleComplete() {
	t.Completed = !t.Completed
}

// AddCompletedPomodoro increments the completed pomodoro count
func (t *Task) AddCompletedPomodoro() {
	t.CompletedPomodoros++
	if t.CompletedPomodoros >= t.PlannedPomodoros {
		t.Completed = true
	}
}

// AddTimeSpent adds duration to the time spent on this task
func (t *Task) AddTimeSpent(duration time.Duration) {
	t.TimeSpent += duration
}

// FormattedTimeSpent returns the formatted time spent on the task
func (t Task) FormattedTimeSpent() string {
	hours := int(t.TimeSpent.Hours())
	minutes := int(t.TimeSpent.Minutes()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// PomodoroProgress returns a string representation of pomodoro progress
func (t Task) PomodoroProgress() string {
	return fmt.Sprintf("[%d/%d]", t.CompletedPomodoros, t.PlannedPomodoros)
}
