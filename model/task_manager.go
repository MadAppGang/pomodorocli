package model

import (
	"time"
)

// TaskManager handles the collection of tasks
type TaskManager struct {
	Tasks         []Task
	ShowCompleted bool
}

// NewTaskManager creates a new task manager
func NewTaskManager() *TaskManager {
	return &TaskManager{
		Tasks:         make([]Task, 0),
		ShowCompleted: true,
	}
}

// LoadTasks loads tasks into the TaskManager
func (tm *TaskManager) LoadTasks(tasks []Task) {
	tm.Tasks = tasks
}

// GetTasks returns all tasks for saving
func (tm *TaskManager) GetTasks() []Task {
	return tm.Tasks
}

// AddTask adds a new task to the manager
func (tm *TaskManager) AddTask(description string, plannedPomodoros int) Task {
	task := NewTask(description, plannedPomodoros)
	tm.Tasks = append(tm.Tasks, task)
	return task
}

// GetTask retrieves a task by ID
func (tm *TaskManager) GetTask(id string) (Task, bool) {
	for i, task := range tm.Tasks {
		if task.ID == id {
			return tm.Tasks[i], true
		}
	}
	return Task{}, false
}

// UpdateTask updates an existing task in the task list
func (tm *TaskManager) UpdateTask(task Task) bool {
	for i, t := range tm.Tasks {
		if t.ID == task.ID {
			tm.Tasks[i] = task
			return true
		}
	}
	return false
}

// DeleteTask removes a task by ID
func (tm *TaskManager) DeleteTask(id string) bool {
	for i, task := range tm.Tasks {
		if task.ID == id {
			// Remove the task by appending everything before and after it
			tm.Tasks = append(tm.Tasks[:i], tm.Tasks[i+1:]...)
			return true
		}
	}
	return false
}

// ToggleShowCompleted toggles whether completed tasks are shown
func (tm *TaskManager) ToggleShowCompleted() {
	tm.ShowCompleted = !tm.ShowCompleted
}

// FilteredTasks returns tasks filtered according to current settings
func (tm *TaskManager) FilteredTasks() []Task {
	if tm.ShowCompleted {
		return tm.Tasks
	}

	filtered := make([]Task, 0)
	for _, task := range tm.Tasks {
		if !task.Completed {
			filtered = append(filtered, task)
		}
	}
	return filtered
}

// IncompleteTasks returns only incomplete tasks
func (tm *TaskManager) IncompleteTasks() []Task {
	filtered := make([]Task, 0)
	for _, task := range tm.Tasks {
		if !task.Completed {
			filtered = append(filtered, task)
		}
	}
	return filtered
}

// CompletedTasks returns only completed tasks
func (tm *TaskManager) CompletedTasks() []Task {
	filtered := make([]Task, 0)
	for _, task := range tm.Tasks {
		if task.Completed {
			filtered = append(filtered, task)
		}
	}
	return filtered
}

// ToggleTaskComplete toggles a task's completion status by ID and returns the updated task
func (tm *TaskManager) ToggleTaskComplete(id string) (Task, bool) {
	for i, task := range tm.Tasks {
		if task.ID == id {
			tm.Tasks[i].Completed = !task.Completed
			return tm.Tasks[i], true
		}
	}
	return Task{}, false
}

// AddCompletedPomodoro increments a task's completed pomodoro count by ID and returns the updated task
func (tm *TaskManager) AddCompletedPomodoro(id string) (Task, bool) {
	for i, task := range tm.Tasks {
		if task.ID == id {
			tm.Tasks[i].CompletedPomodoros++
			if tm.Tasks[i].CompletedPomodoros >= task.PlannedPomodoros {
				tm.Tasks[i].Completed = true
			}
			return tm.Tasks[i], true
		}
	}
	return Task{}, false
}

// AddTimeSpent adds time to a task by ID and returns the updated task
func (tm *TaskManager) AddTimeSpent(id string, duration time.Duration) (Task, bool) {
	for i, task := range tm.Tasks {
		if task.ID == id {
			tm.Tasks[i].TimeSpent += duration
			return tm.Tasks[i], true
		}
	}
	return Task{}, false
}
