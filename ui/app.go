package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jackrudenko/pomodorocli/model"
	"github.com/jackrudenko/pomodorocli/storage"
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
	timer          *model.Timer
	taskManager    *model.TaskManager
	storageManager *storage.StorageManager
	view           ViewState
	width          int
	height         int

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

	// Initialize storage
	jsonStorage, err := storage.NewJSONTaskStorage("./data/tasks.json")
	var storageManager *storage.StorageManager
	if err == nil {
		storageManager = storage.NewStorageManager(jsonStorage, taskManager)
		// Load tasks from storage
		if err := storageManager.LoadTasks(); err != nil {
			// If loading fails, we'll start with an empty task list
			fmt.Println("Error loading tasks:", err)
		}
	}

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
		storageManager: storageManager,
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
	// Only add sample tasks if we don't have any (i.e., no tasks were loaded from storage)
	if len(a.taskManager.GetTasks()) == 0 {
		// Add some sample tasks for demonstration
		a.taskManager.AddTask("Work on design concept", 4)
		a.taskManager.AddTask("Test the prototype with users", 3)
		a.taskManager.AddTask("Create a design concept for the Evergen App / Link", 3)

		// Save the initial tasks
		if a.storageManager != nil {
			if err := a.storageManager.SaveTasks(); err != nil {
				fmt.Println("Error saving initial tasks:", err)
			}
		}
	}

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
			model, cmd := a.updateMainView(msg)
			// Save tasks after certain actions in main view
			if msg.String() == " " { // Space toggles task completion
				if a.storageManager != nil {
					if err := a.storageManager.SaveTasks(); err != nil {
						fmt.Println("Error saving tasks:", err)
					}
				}
			}
			return model, cmd
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
			if selectedTaskPtr := a.taskListView.GetSelectedTaskPtr(); selectedTaskPtr != nil {
				a.timer.SetCurrentTask(*selectedTaskPtr)
				// Update task list view with current task
				a.taskListView.SetCurrentTask(selectedTaskPtr)
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
		if selectedTaskPtr := a.taskListView.GetSelectedTaskPtr(); selectedTaskPtr != nil {
			a.timer.SetCurrentTask(*selectedTaskPtr)
			// Update task list view with current task
			a.taskListView.SetCurrentTask(selectedTaskPtr)
			a.timer.Start()
		}

	case " ":
		// Toggle task complete
		a.taskListView.ToggleSelectedTaskComplete()

	case "d":
		// Delete the selected task
		if selectedTaskPtr := a.taskListView.GetSelectedTaskPtr(); selectedTaskPtr != nil {
			a.taskListView.DeleteSelectedTask()
			// Save tasks after deletion
			if a.storageManager != nil {
				if err := a.storageManager.SaveTasks(); err != nil {
					fmt.Println("Error saving tasks:", err)
				}
			}
		}

	case "f2":
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

			// Save tasks after adding a new one
			if a.storageManager != nil {
				if err := a.storageManager.SaveTasks(); err != nil {
					fmt.Println("Error saving tasks:", err)
				}
			}

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
		return a.debugView()
	}

	// Regular rendering for normal mode
	// Create main container with the background color
	mainContainerStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Width(a.width - 4)

	// Create inner box with rounded borders - using ColorBackground instead of ColorBoxBackground
	innerBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(a.width - 10).
		Height(a.height - 4)

	// Render the timer component
	timerSection := lipgloss.NewStyle().
		MarginTop(2). // 2 lines of margin at the top
		Render(a.timerView.Render())

	// Create divider with proper styling
	divider := lipgloss.NewStyle().
		Foreground(ColorGrayText).
		Padding(0, 0, 2, 0).
		AlignHorizontal(lipgloss.Center).
		Render(strings.Repeat("─", a.width-16))

	// Render the task list component but reduce its padding
	a.taskListView.SetWidth(a.width - 40) // Adjust width to account for inner box padding
	// taskListSection := lipgloss.NewStyle().
	// 	AlignHorizontal(lipgloss.Center).
	// 	Render(a.taskListView.Render())
	taskListSection := lipgloss.PlaceHorizontal(
		a.width-10,
		lipgloss.Center,
		a.taskListView.Render(),
	)
	// Assemble the inner content
	innerContent := lipgloss.JoinVertical(
		lipgloss.Center,
		timerSection,
		divider,
		taskListSection,
	)

	// Apply the inner box styling
	styledContent := innerBoxStyle.Render(innerContent)

	// Apply the main container styling
	debugModeText := ""
	if a.debugMode != NoDebug {
		debugModeText = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Align(lipgloss.Center).
			Render(fmt.Sprintf("\nPress [F2] to cycle debug modes"))
	}

	return mainContainerStyle.Render(styledContent + debugModeText)
}

func (a *App) debugView() string {
	// Create a base style with the correct background color for the entire debug view
	baseStyle := lipgloss.NewStyle()

	var builder strings.Builder
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

		// Add vertical padding with background color
		paddingStyle := lipgloss.NewStyle()
		paddedTimer := strings.Repeat(paddingStyle.Render("\n"), verticalPadding) + timerSection

		// Center the timer horizontally with background
		debugContent = lipgloss.PlaceHorizontal(a.width, lipgloss.Center, paddedTimer)

	} else if a.debugMode == TaskListDebug {
		debugTitle = "Debug Mode: Task List Component Only"

		// Render task list view with background
		tasksContent := a.taskListView.Render()
		// tasksContent := "a.taskListView.Render()\n\n\n\n\\n debugView()"

		// Make it half the width for better readability but don't add background
		containerStyle := lipgloss.NewStyle().
			Padding(1, 2).
			Width(a.width - 20)

		// Apply styling and center
		styledTaskList := containerStyle.Render(tasksContent)

		// Center the task list horizontally and add some vertical padding with background
		verticalPadding := (a.height - 30) / 4 // Less padding than timer as task list is taller
		if verticalPadding < 0 {
			verticalPadding = 0
		}

		paddingStyle := lipgloss.NewStyle()
		paddedTaskList := strings.Repeat(paddingStyle.Render("\n"), verticalPadding) + styledTaskList
		debugContent = lipgloss.PlaceHorizontal(a.width, lipgloss.Center, paddedTaskList)
	}

	// Render debug title and content with background
	debugTitleStyle := debugStyle()
	builder.WriteString(debugTitleStyle.Render(debugTitle))
	builder.WriteString("\n")
	builder.WriteString(debugContent)
	builder.WriteString("\n\n")

	exitMessageStyle := debugStyle()
	exitMessage = exitMessageStyle.Render("Press [F2] to cycle debug modes")
	builder.WriteString(exitMessage)

	// Apply the background to the entire view
	return baseStyle.Render(builder.String())
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

// SetDebugMode sets the debug mode for the app
func (a *App) SetDebugMode(mode DebugMode) {
	a.debugMode = mode
}

// SetTimerOnlyMode sets the app to show only the timer component
func (a *App) SetTimerOnlyMode(timerOnly bool) {
	if timerOnly {
		a.debugMode = TimerDebug
	} else {
		a.debugMode = NoDebug
	}
}

// SetTaskListOnlyMode sets the app to show only the task list component
func (a *App) SetTaskListOnlyMode(tasksOnly bool) {
	if tasksOnly {
		a.debugMode = TaskListDebug
	} else {
		a.debugMode = NoDebug
	}
}

// RenderMainView renders the main view and returns it as a string
// Used in print mode to output the view and exit
func (a *App) RenderMainView() string {
	// Initialize with sample tasks if needed
	if len(a.taskManager.Tasks) == 0 {
		a.taskManager.AddTask("Work on design concept", 4)
		a.taskManager.AddTask("Test the prototype with users", 3)
		a.taskManager.AddTask("Create a design concept for the Evergen App / Link", 3)
	}

	// Use the existing mainView method to render
	return a.mainView()
}

// RenderTimerView renders only the timer view and returns it as a string
// Used in print mode to output the view and exit
func (a *App) RenderTimerView() string {
	// Initialize with sample tasks if needed
	if len(a.taskManager.Tasks) == 0 {
		task := a.taskManager.AddTask("Work on design concept", 4)
		a.timer.SetCurrentTask(task)
	}

	// Force timer debug mode
	a.debugMode = TimerDebug

	// Use the existing debugView method to render
	return a.debugView()
}

// RenderTaskListView renders only the task list view and returns it as a string
// Used in print mode to output the view and exit
func (a *App) RenderTaskListView() string {
	// Initialize with sample tasks if needed
	if len(a.taskManager.Tasks) == 0 {
		a.taskManager.AddTask("Work on design concept", 4)
		a.taskManager.AddTask("Test the prototype with users", 3)
		a.taskManager.AddTask("Create a design concept for the Evergen App / Link", 3)
	}

	// Force task list debug mode
	a.debugMode = TaskListDebug

	// Use the existing debugView method to render
	return a.debugView()
}
