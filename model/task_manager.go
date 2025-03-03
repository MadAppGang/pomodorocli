package model

// TaskManager handles the collection of tasks
type TaskManager struct {
	Tasks         []*Task
	nextID        int
	ShowCompleted bool
}

// NewTaskManager creates a new task manager
func NewTaskManager() *TaskManager {
	return &TaskManager{
		Tasks:         make([]*Task, 0),
		nextID:        1,
		ShowCompleted: true,
	}
}

// AddTask adds a new task to the manager
func (tm *TaskManager) AddTask(description string, plannedPomodoros int) *Task {
	task := NewTask(tm.nextID, description, plannedPomodoros)
	tm.Tasks = append(tm.Tasks, task)
	tm.nextID++
	return task
}

// GetTask retrieves a task by ID
func (tm *TaskManager) GetTask(id int) *Task {
	for _, task := range tm.Tasks {
		if task.ID == id {
			return task
		}
	}
	return nil
}

// DeleteTask removes a task by ID
func (tm *TaskManager) DeleteTask(id int) bool {
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
func (tm *TaskManager) FilteredTasks() []*Task {
	if tm.ShowCompleted {
		return tm.Tasks
	}

	filtered := make([]*Task, 0)
	for _, task := range tm.Tasks {
		if !task.Completed {
			filtered = append(filtered, task)
		}
	}
	return filtered
}

// IncompleteTasks returns only incomplete tasks
func (tm *TaskManager) IncompleteTasks() []*Task {
	filtered := make([]*Task, 0)
	for _, task := range tm.Tasks {
		if !task.Completed {
			filtered = append(filtered, task)
		}
	}
	return filtered
}

// CompletedTasks returns only completed tasks
func (tm *TaskManager) CompletedTasks() []*Task {
	filtered := make([]*Task, 0)
	for _, task := range tm.Tasks {
		if task.Completed {
			filtered = append(filtered, task)
		}
	}
	return filtered
}
