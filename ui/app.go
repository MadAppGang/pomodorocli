package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jackrudenko/pomodorocli/model"
)

// ViewState represents the different views in the application
type ViewState int

const (
	// MainView is the main timer and task list view
	MainView ViewState = iota
	// AddTaskView is the view for adding a new task
	AddTaskView
	// SettingsView is the view for configuring settings
	SettingsView
)

// TickMsg is sent when the timer should update
type TickMsg time.Time

// WindowSizeMsg is sent when the terminal window size changes
type WindowSizeMsg struct {
	Width  int
	Height int
}

// DebugMode represents different debug view modes
type DebugMode int

const (
	// NoDebug is normal mode with all components
	NoDebug DebugMode = iota
	// TimerDebug shows only the timer component
	TimerDebug
	// TaskListDebug shows only the task list component
	TaskListDebug
)

// App is the main application model
type App struct {
	// Application state
	timer       *model.Timer
	taskManager *model.TaskManager
	view        ViewState
	width       int
	height      int

	// Input fields for adding tasks
	taskInput      textinput.Model
	pomodorosInput textinput.Model
	inputting      bool

	// Components
	timerView    *TimerView
	taskListView *TaskListView

	// Debug mode
	debugMode DebugMode

	// Font manager for rendering big digits
	fontManager *FontManager
}

// NewApp creates a new application model
func NewApp() *App {
	taskInput := textinput.New()
	taskInput.Placeholder = "Task description"
	taskInput.Width = 60
	taskInput.Focus()

	pomodorosInput := textinput.New()
	pomodorosInput.Placeholder = "Number of pomodoros (default: 4)"
	pomodorosInput.Width = 10

	width := GetTerminalWidth()
	height := GetTerminalHeight()

	timer := model.NewTimer()
	taskManager := model.NewTaskManager()

	// Initialize the font manager
	fontManager, err := NewFontManager()
	if err != nil {
		// If we can't load fonts, continue with a nil fontManager
		// The timer component will fall back to the hardcoded digits
		fontManager = nil
	}

	app := &App{
		timer:          timer,
		taskManager:    taskManager,
		view:           MainView,
		width:          width,
		height:         height,
		taskInput:      taskInput,
		pomodorosInput: pomodorosInput,
		inputting:      false,
		debugMode:      NoDebug,
		fontManager:    fontManager,
	}

	// Initialize components
	app.timerView = NewTimerView(timer, width)
	app.taskListView = NewTaskListView(taskManager, width)

	// Set the font manager in the timer view
	if fontManager != nil {
		app.timerView.SetFontManager(fontManager)
	}

	return app
}

// Init initializes the Bubble Tea program
func (a *App) Init() tea.Cmd {
	// Add some sample tasks for demonstration
	a.taskManager.AddTask("Work on design concept", 4)
	a.taskManager.AddTask("Test the prototype with users", 3)
	a.taskManager.AddTask("Create a design concept for the Evergen App / Link", 3)

	// Start the timer ticker and request initial window size
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// Update handles messages and user input
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Handle window resize
		a.width = msg.Width
		a.height = msg.Height

		// Update styles with new dimensions
		UpdateStyles()

		// Update component dimensions
		a.timerView.SetWidth(a.width)
		a.taskListView.SetWidth(a.width)

		return a, nil

	case TickMsg:
		// Update the timer
		a.timer.Update()

		// Sync the current task to the task list view
		a.taskListView.SetCurrentTask(a.timer.CurrentTask)

		// Continue ticking
		return a, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return TickMsg(t)
		})

	case tea.KeyMsg:
		switch a.view {
		case MainView:
			return a.updateMainView(msg)
		case AddTaskView:
			return a.updateAddTaskView(msg)
		case SettingsView:
			return a.updateSettingsView(msg)
		}
	}

	return a, cmd
}

// updateMainView handles input for the main view
func (a *App) updateMainView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return a, tea.Quit

	case "s":
		// Stop the timer
		if a.timer.State == model.TimerRunning {
			a.timer.Stop()
		} else if a.timer.State == model.TimerStopped {
			// If we have selected task, set it as current and start
			if selectedTask := a.taskListView.GetSelectedTask(); selectedTask != nil {
				a.timer.SetCurrentTask(selectedTask)
				// Update task list view with current task
				a.taskListView.SetCurrentTask(selectedTask)
				a.timer.Start()
			}
		}

	case "p":
		// Pause/resume timer
		if a.timer.State == model.TimerRunning {
			a.timer.Pause()
		} else if a.timer.State == model.TimerPaused {
			a.timer.Resume()
		}

	case "n":
		// Add new task
		a.view = AddTaskView
		a.taskInput.Focus()
		a.inputting = true

	case "h":
		// Toggle hiding completed tasks
		a.taskListView.ToggleShowCompleted()

	case "j", "down":
		// Move down in task list
		a.taskListView.MoveSelectionDown()

	case "k", "up":
		// Move up in task list
		a.taskListView.MoveSelectionUp()

	case "enter":
		// Select current task
		if selectedTask := a.taskListView.GetSelectedTask(); selectedTask != nil {
			a.timer.SetCurrentTask(selectedTask)
			// Update task list view with current task
			a.taskListView.SetCurrentTask(selectedTask)
			a.timer.Start()
		}

	case " ":
		// Toggle task complete
		a.taskListView.ToggleSelectedTaskComplete()

	case "d":
		// Cycle through debug modes: NoDebug -> TimerDebug -> TaskListDebug -> NoDebug
		a.debugMode = (a.debugMode + 1) % 3

	case "f":
		// Toggle through fonts if font manager is available
		if a.fontManager != nil {
			a.fontManager.NextFont()
		}
	}

	return a, nil
}

// updateAddTaskView handles input for the add task view
func (a *App) updateAddTaskView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "ctrl+c", "esc":
		// Cancel adding task
		a.view = MainView
		a.inputting = false
		a.taskInput.Reset()
		a.pomodorosInput.Reset()

	case "tab":
		// Switch between inputs
		if a.taskInput.Focused() {
			a.taskInput.Blur()
			a.pomodorosInput.Focus()
		} else {
			a.pomodorosInput.Blur()
			a.taskInput.Focus()
		}

	case "enter":
		// Submit new task
		if a.taskInput.Value() != "" {
			description := strings.TrimSpace(a.taskInput.Value())
			pomodoros := 4 // Default

			// Try to parse the pomodoros input
			if a.pomodorosInput.Value() != "" {
				fmt.Sscanf(a.pomodorosInput.Value(), "%d", &pomodoros)
				if pomodoros <= 0 {
					pomodoros = 1
				}
			}

			a.taskManager.AddTask(description, pomodoros)
			a.view = MainView
			a.inputting = false
			a.taskInput.Reset()
			a.pomodorosInput.Reset()
		}
	}

	// Handle text input updates
	if a.taskInput.Focused() {
		a.taskInput, cmd = a.taskInput.Update(msg)
		return a, cmd
	} else if a.pomodorosInput.Focused() {
		a.pomodorosInput, cmd = a.pomodorosInput.Update(msg)
		return a, cmd
	}

	return a, nil
}

// updateSettingsView handles input for the settings view
func (a *App) updateSettingsView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc", "q":
		a.view = MainView
	}

	return a, nil
}

// View renders the current UI
func (a *App) View() string {
	switch a.view {
	case MainView:
		return a.mainView()
	case AddTaskView:
		return a.addTaskView()
	case SettingsView:
		return a.settingsView()
	default:
		return "Unknown view"
	}
}

// mainView renders the main application view
func (a *App) mainView() string {
	var builder strings.Builder

	// App title
	builder.WriteString(AppNameStyle.Render("~ pomodoro tracker"))
	builder.WriteString("\n\n")

	// Debug mode: Timer Only or TaskList Only
	if a.debugMode != NoDebug {
		var debugTitle, debugContent, exitMessage string

		if a.debugMode == TimerDebug {
			debugTitle = "Debug Mode: Timer Component Only"

			// Render timer view
			timerSection := a.timerView.Render()

			// Center the timer horizontally and vertically
			verticalPadding := (a.height - 20) / 2 // Estimate height of timer display
			if verticalPadding < 0 {
				verticalPadding = 0
			}

			// Add vertical padding
			paddedTimer := strings.Repeat("\n", verticalPadding) + timerSection

			// Center the timer horizontally
			debugContent = lipgloss.PlaceHorizontal(a.width, lipgloss.Center, paddedTimer)

		} else if a.debugMode == TaskListDebug {
			debugTitle = "Debug Mode: Task List Component Only"

			// Render task list view with background
			tasksContent := a.taskListView.Render()

			// Apply consistent background styling
			containerStyle := lipgloss.NewStyle().
				Background(ColorBoxBackground).
				Padding(1, 2).
				Width(a.width / 2) // Make it half the width for better readability

			// Apply styling and center
			styledTaskList := containerStyle.Render(tasksContent)

			// Center the task list horizontally and add some vertical padding
			verticalPadding := (a.height - 30) / 4 // Less padding than timer as task list is taller
			if verticalPadding < 0 {
				verticalPadding = 0
			}

			paddedTaskList := strings.Repeat("\n", verticalPadding) + styledTaskList
			debugContent = lipgloss.PlaceHorizontal(a.width, lipgloss.Center, paddedTaskList)
		}

		// Render debug title and content
		builder.WriteString(debugStyle().Render(debugTitle))
		builder.WriteString("\n\n")
		builder.WriteString(debugContent)
		builder.WriteString("\n\n")

		// Exit message
		exitMessage = debugStyle().Render("Press [D] to cycle debug modes")
		builder.WriteString(lipgloss.PlaceHorizontal(a.width, lipgloss.Center, exitMessage))

		return builder.String() // No need for AppStyle here
	}

	// Regular rendering for normal mode
	// Create container style for consistent background
	containerStyle := lipgloss.NewStyle().
		Background(ColorBoxBackground).
		Width(a.width - 12)

	// Render timer view using the extracted component and apply container background
	timerContent := a.timerView.Render()
	builder.WriteString("\n\n")
	timerSection := containerStyle.Align(lipgloss.Center).Render(timerContent)

	// Create divider without its own background
	divider := strings.Repeat("─", a.width-20)
	styledDivider := DividerStyle.Render(divider)

	// Render task list and apply container background
	tasksContent := a.taskListView.Render()
	tasksSection := containerStyle.Render(tasksContent)

	// Join everything vertically with consistent backgrounds
	content := lipgloss.JoinVertical(lipgloss.Left,
		timerSection,
		containerStyle.Align(lipgloss.Center).Render(styledDivider),
		tasksSection,
	)

	// Wrap in box style
	renderedBox := BoxStyle.
		Padding(1, 2).
		Width(a.width - 8).
		Render(content)

	builder.WriteString(renderedBox)

	// Bottom menu
	menuItems := []string{"settings", "statistics", "log in with google"}
	menuBar := lipgloss.JoinHorizontal(lipgloss.Center,
		MenuItemStyle.Render(menuItems[0]),
		MenuItemStyle.Render(menuItems[1]),
		MenuItemStyle.Render(menuItems[2]),
	)

	builder.WriteString("\n")
	builder.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(a.width - 4).Render(menuBar))

	builder.WriteString("\n")
	builder.WriteString(debugStyle().Render("Press [D] to enter debug mode"))

	return AppStyle.Render(builder.String())
}

// addTaskView renders the view for adding a new task
func (a *App) addTaskView() string {
	var builder strings.Builder

	builder.WriteString(TitleStyle.Render("Add New Task"))
	builder.WriteString("\n\n")

	builder.WriteString("Task Description:\n")
	builder.WriteString(a.taskInput.View())
	builder.WriteString("\n\n")

	builder.WriteString("Number of Pomodoros:\n")
	builder.WriteString(a.pomodorosInput.View())
	builder.WriteString("\n\n")

	builder.WriteString("Press Enter to add, Esc to cancel, Tab to switch fields")

	return BoxStyle.Render(builder.String())
}

// settingsView renders the settings view
func (a *App) settingsView() string {
	var builder strings.Builder

	builder.WriteString(TitleStyle.Render("Settings"))
	builder.WriteString("\n\n")

	builder.WriteString("Not implemented yet.")
	builder.WriteString("\n\n")

	builder.WriteString("Press Esc to go back")

	return BoxStyle.Render(builder.String())
}

// Add this helper function for debug styling at the end of the file
// debugStyle returns a style for debug messages
func debugStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")). // Bright pink color
		Bold(true)
}

// SetTimerOnlyMode sets the application to timer-only mode (debug mode)
func (a *App) SetTimerOnlyMode(enabled bool) {
	if enabled {
		a.debugMode = TimerDebug
	} else {
		a.debugMode = NoDebug
	}
}

// SetTaskListOnlyMode sets the application to task-list-only mode
func (a *App) SetTaskListOnlyMode(enabled bool) {
	if enabled {
		a.debugMode = TaskListDebug
	} else {
		a.debugMode = NoDebug
	}
}
