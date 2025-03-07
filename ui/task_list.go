package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jackrudenko/pomodorocli/model"
)

// TaskListView represents the task list component
type TaskListView struct {
	taskManager    *model.TaskManager
	selectedIndex  int
	width          int
	currentTask    *model.Task // We'll keep this as a pointer since it's just a reference
	hasCurrentTask bool        // Flag to track if we have a current task
	currentTaskID  string      // Store the ID of the current task
}

// NewTaskListView creates a new task list view
func NewTaskListView(taskManager *model.TaskManager, width int) *TaskListView {
	return &TaskListView{
		taskManager:    taskManager,
		selectedIndex:  0,
		width:          width,
		currentTask:    nil,
		hasCurrentTask: false,
		currentTaskID:  "",
	}
}

// SetWidth updates the width of the task list view
func (t *TaskListView) SetWidth(width int) {
	t.width = width
}

// GetSelectedTask returns the currently selected task, or empty task if no tasks
func (t *TaskListView) GetSelectedTask() model.Task {
	tasks := t.taskManager.FilteredTasks()
	if len(tasks) == 0 {
		return model.Task{} // Return an empty task
	}
	return tasks[t.selectedIndex%len(tasks)]
}

// GetSelectedTaskPtr returns a pointer to the currently selected task, or nil if no tasks
// This is needed for compatibility with code that expects pointers
func (t *TaskListView) GetSelectedTaskPtr() *model.Task {
	tasks := t.taskManager.FilteredTasks()
	if len(tasks) == 0 {
		return nil
	}

	// Get the selected task
	selectedTask := tasks[t.selectedIndex%len(tasks)]

	// Find and return a pointer to the actual task in the task manager
	for i := range t.taskManager.Tasks {
		if t.taskManager.Tasks[i].ID == selectedTask.ID {
			return &t.taskManager.Tasks[i]
		}
	}

	return nil
}

// MoveSelectionDown moves the selection down in the task list
func (t *TaskListView) MoveSelectionDown() {
	t.selectedIndex++
}

// MoveSelectionUp moves the selection up in the task list
func (t *TaskListView) MoveSelectionUp() {
	if t.selectedIndex > 0 {
		t.selectedIndex--
	}
}

// ToggleSelectedTaskComplete toggles the completion status of the selected task
func (t *TaskListView) ToggleSelectedTaskComplete() {
	tasks := t.taskManager.FilteredTasks()
	if len(tasks) == 0 {
		return
	}

	// Get the index of the selected task
	index := t.selectedIndex % len(tasks)
	// Get the task
	task := tasks[index]
	// Toggle completion
	task.ToggleComplete()
	// Update the task in the task manager
	t.taskManager.UpdateTask(task)
}

// ToggleShowCompleted toggles showing or hiding completed tasks
func (t *TaskListView) ToggleShowCompleted() {
	t.taskManager.ToggleShowCompleted()
}

// SetCurrentTask sets the current active task
func (t *TaskListView) SetCurrentTask(task *model.Task) {
	t.currentTask = task
	if task == nil {
		t.hasCurrentTask = false
		t.currentTaskID = ""
	} else {
		t.hasCurrentTask = true
		t.currentTaskID = task.ID
	}
}

// DeleteSelectedTask deletes the currently selected task
func (t *TaskListView) DeleteSelectedTask() {
	tasks := t.taskManager.FilteredTasks()
	if len(tasks) == 0 {
		return
	}

	// Get the index of the selected task
	index := t.selectedIndex % len(tasks)
	// Get the task
	task := tasks[index]
	// Delete the task from the task manager
	t.taskManager.DeleteTask(task.ID)

	// Adjust selection index to prevent out of bounds
	if len(t.taskManager.FilteredTasks()) > 0 && t.selectedIndex >= len(t.taskManager.FilteredTasks()) {
		t.selectedIndex = len(t.taskManager.FilteredTasks()) - 1
	}

	// If the deleted task was the current task, clear it
	if t.hasCurrentTask && t.currentTaskID == task.ID {
		t.currentTask = nil
		t.hasCurrentTask = false
		t.currentTaskID = ""
	}
}

// Render renders the task list component
func (t *TaskListView) Render() string {
	// Get content
	taskControls := t.renderTaskControls()
	tasksContent := t.renderTaskList()

	// Combine the content
	combined := lipgloss.JoinVertical(
		lipgloss.Left,
		taskControls,
		tasksContent,
	)

	// Apply a consistent background color to the entire component
	return lipgloss.NewStyle().
		Padding(0, 2). // Horizontal padding only, no vertical padding
		Width(t.width).
		Render(combined)
}

// renderTaskList returns the rendered task list
func (t *TaskListView) renderTaskList() string {
	var tasks []string

	// Add padding for tasks
	for i, task := range t.taskManager.FilteredTasks() {
		isSelected := i == (t.selectedIndex % len(t.taskManager.FilteredTasks()))
		isCurrentTask := t.hasCurrentTask && task.ID == t.currentTaskID

		// Task number and selection indicator
		var taskNumber string

		// Add selection indicator if selected
		var prefix string
		if isSelected {
			prefix = lipgloss.NewStyle().
				Foreground(ColorTaskTag).
				Bold(true).
				Render("👉 ")
		} else {
			prefix = "   "
		}

		// Prepare task number for rendering, use digits for consistent width
		taskNumber = fmt.Sprintf("%s%d", prefix, i+1)

		// Task progress
		taskProgress := task.PomodoroProgress()

		// Task time spent
		taskTimeSpent := task.FormattedTimeSpent()

		// Check if the description contains "Link" to highlight it
		description := task.Description
		taskDescription := description
		linkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#1E90FF")) // Bright blue for Link

		// Handle Link highlighting if present
		if strings.Contains(description, "Link") {
			parts := strings.Split(description, "Link")
			beforeLink := parts[0]
			afterLink := ""
			if len(parts) > 1 {
				afterLink = parts[1]
			}

			// Create the combined description with highlighted Link
			plainDesc := beforeLink
			highlightedLink := linkStyle.Render("Link")
			taskDescription = fmt.Sprintf("%s%s%s", plainDesc, highlightedLink, afterLink)
		}

		// Apply styling based on state
		var taskNumberStyle, taskDescStyle, taskProgressStyle, taskTimeStyle lipgloss.Style

		// Base styles - no explicit background
		taskNumberStyle = TaskStyle
		taskDescStyle = TaskStyle
		taskProgressStyle = TaskProgressStyle
		taskTimeStyle = TaskTimeStyle

		// Selected task styling
		if isSelected {
			taskNumberStyle = taskNumberStyle.Bold(true)
			taskDescStyle = taskDescStyle.Bold(true)
			taskProgressStyle = taskProgressStyle.Bold(true)
		}

		// Current task styling
		if isCurrentTask {
			taskDescStyle = CurrentTaskStyle
			taskDescStyle = taskDescStyle.Bold(true)
		}

		// Completed task styling (lighter color)
		if task.Completed {
			taskDescStyle = taskDescStyle.Foreground(ColorProgressBar) // Use gray for completed tasks
			taskTimeStyle = taskTimeStyle.Foreground(ColorProgressBar)
		}

		// Render the task components
		renderedNumber := taskNumberStyle.Render(taskNumber)
		renderedProgress := taskProgressStyle.Render(taskProgress)
		renderedTime := taskTimeStyle.Render(taskTimeSpent)

		// For consistency, first get the regular number rendering for spacing calculation
		regularNumber := taskNumberStyle.Render(taskNumber)

		// Handle special rendering for current task
		if isCurrentTask {
			clockEmoji := "⏰"
			// Keep the same exact width as the regular number by using spaces
			numberWidth := lipgloss.Width(regularNumber)
			prefixWidth := lipgloss.Width(prefix)

			// Calculate padding needed (ensure it's not negative)
			paddingSize := numberWidth - prefixWidth - lipgloss.Width(clockEmoji)
			if paddingSize < 0 {
				paddingSize = 0
			}

			// Create the display with exact spacing: prefix + clock emoji + padding if needed
			renderedNumber = taskNumberStyle.Render(fmt.Sprintf("%s%s%s",
				prefix,
				clockEmoji,
				strings.Repeat(" ", paddingSize)))
		}

		// Add +task prefix for the task description
		renderedDesc := fmt.Sprintf("%s %s",
			taskProgressStyle.Render("+task"),
			taskDescStyle.Render(taskDescription))

		// Adjust the layout based on reference screenshot
		// Based on the screenshot, we need specific ordering and spacing:
		// 1. Number
		// 2. Progress [X/Y]
		// 3. Time (50m, 1h 15m)
		// 4. Description (+task...)

		// Create strings for layout
		var fullTaskLine string

		// Create layout with fixed spacing that matches the reference
		progressWidth := lipgloss.Width(renderedProgress)
		timeWidth := lipgloss.Width(renderedTime)

		// Position the elements with fixed spacing as in the screenshot
		fullTaskLine = renderedNumber
		// Adjusted spacing to account for the selection indicator
		spacingAfterNumber := 1
		if spacingAfterNumber > 0 && !isCurrentTask {
			fullTaskLine += strings.Repeat(" ", spacingAfterNumber)
		}
		fullTaskLine += " "
		fullTaskLine += renderedProgress
		// Ensure spacing is never negative
		spacingAfterProgress := 8 - progressWidth
		if spacingAfterProgress < 0 {
			spacingAfterProgress = 0
		}
		fullTaskLine += strings.Repeat(" ", spacingAfterProgress) // Fixed spacing after progress
		fullTaskLine += renderedTime
		// Ensure spacing is never negative
		spacingAfterTime := 4 - timeWidth
		if spacingAfterTime < 0 {
			spacingAfterTime = 0
		}
		fullTaskLine += strings.Repeat(" ", spacingAfterTime) // Fixed spacing after time
		fullTaskLine += renderedDesc

		tasks = append(tasks, fullTaskLine)
	}

	if len(t.taskManager.FilteredTasks()) == 0 {
		tasks = append(tasks, TaskStyle.Render("No tasks. Add a new task with [N]."))
	}

	// Add the "Add new task" control at the bottom with consistent styling
	// Use padding instead of empty lines for spacing
	addNewTaskStyle := AddNewTaskStyle.PaddingTop(1)
	tasks = append(tasks, addNewTaskStyle.Render("Add new task [N]"))

	// Just join the tasks vertically without additional wrapping
	return lipgloss.JoinVertical(lipgloss.Left, tasks...)
}

// renderTaskControls returns the rendered task controls
func (t *TaskListView) renderTaskControls() string {
	// Use a header without explicit background
	tasksHeader := TasksHeaderStyle.Render("Tasks")

	// Match the Figma design styling for controls
	var hideCompletedText string
	if t.taskManager.ShowCompleted {
		hideCompletedText = "[H] Hide completed tasks"
	} else {
		hideCompletedText = "[H] Show completed tasks"
	}

	// Render hide completed control without margin or explicit background
	hideCompleted := HideCompletedStyle.
		MarginTop(0).
		MarginBottom(0).
		Render(hideCompletedText)

	// Add delete task control
	deleteTask := HideCompletedStyle.
		MarginTop(0).
		MarginBottom(0).
		Render("[D] Delete task")

	// Simple spacer without explicit background
	spacer := "       "

	// Join horizontally without explicit background wrapping
	return lipgloss.JoinHorizontal(lipgloss.Left, tasksHeader, spacer, hideCompleted, spacer, deleteTask)
}
