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

	// Selected index for task list
	selectedIndex int

	// Components
	timerView *TimerView

	// Debug mode
	debugMode bool

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

	// Initialize the font manager
	fontManager, err := NewFontManager()
	if err != nil {
		// If we can't load fonts, continue with a nil fontManager
		// The timer component will fall back to the hardcoded digits
		fontManager = nil
	}

	app := &App{
		timer:          timer,
		taskManager:    model.NewTaskManager(),
		view:           MainView,
		width:          width,
		height:         height,
		taskInput:      taskInput,
		pomodorosInput: pomodorosInput,
		inputting:      false,
		selectedIndex:  0,
		debugMode:      false,
		fontManager:    fontManager,
	}

	app.timerView = NewTimerView(timer, width)

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

		return a, nil

	case TickMsg:
		// Update the timer
		a.timer.Update()

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
			// If we have tasks, set the current task and start
			tasks := a.taskManager.IncompleteTasks()
			if len(tasks) > 0 {
				a.timer.SetCurrentTask(tasks[a.selectedIndex%len(tasks)])
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
		a.taskManager.ToggleShowCompleted()

	case "j", "down":
		// Move down in task list
		a.selectedIndex++

	case "k", "up":
		// Move up in task list
		if a.selectedIndex > 0 {
			a.selectedIndex--
		}

	case "enter":
		// Select current task
		tasks := a.taskManager.FilteredTasks()
		if len(tasks) > 0 {
			selectedTask := tasks[a.selectedIndex%len(tasks)]
			a.timer.SetCurrentTask(selectedTask)
			a.timer.Start()
		}

	case " ":
		// Toggle task complete
		tasks := a.taskManager.FilteredTasks()
		if len(tasks) > 0 {
			selectedTask := tasks[a.selectedIndex%len(tasks)]
			selectedTask.ToggleComplete()
		}

	case "d":
		// Toggle debug mode
		a.debugMode = !a.debugMode

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

	// If in debug mode, only render the timer component
	if a.debugMode {
		builder.WriteString(debugStyle().Render("Debug Mode: Timer Component Only"))
		builder.WriteString("\n\n")

		// Render timer view using the extracted component
		timerSection := a.timerView.Render()

		// Center the timer horizontally using lipgloss.PlaceHorizontal
		centeredTimer := lipgloss.PlaceHorizontal(a.width, lipgloss.Center, timerSection)
		builder.WriteString(centeredTimer)

		builder.WriteString("\n\n")

		// Center the exit message as well
		exitMessage := debugStyle().Render("Press [D] to exit debug mode")
		centeredExitMessage := lipgloss.PlaceHorizontal(a.width, lipgloss.Center, exitMessage)
		builder.WriteString(centeredExitMessage)

		return builder.String() // No need for AppStyle here, as we want full width
	}

	// Regular rendering for normal mode
	// Render timer view using the extracted component
	timerSection := a.timerView.Render()

	// Create divider without background
	divider := strings.Repeat("─", a.width-20)

	// Create task content
	tasksContent := a.renderTaskList()
	taskControls := a.renderTaskControls()

	// Apply background to the remaining content
	tasksSection := lipgloss.NewStyle().
		Background(ColorBoxBackground).
		Width(a.width - 12).
		Render(tasksContent + taskControls)

	// Apply background to the divider
	styledDivider := lipgloss.NewStyle().
		Background(ColorBoxBackground).
		Width(a.width - 12).
		Align(lipgloss.Center).
		Render(DividerStyle.Render(divider))

	// Now apply the background to the entire box content at once
	boxContent := timerSection + "\n" +
		styledDivider + "\n" +
		tasksSection

	// Wrap all content in a single background style
	renderedBox := BoxStyle.
		Padding(1, 2).
		Width(a.width - 8).
		Render(boxContent)

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

	// Debug mode info
	if !a.debugMode {
		builder.WriteString("\n")
		builder.WriteString(debugStyle().Render("Press [D] to enter debug mode"))
	}

	return AppStyle.Render(builder.String())
}

// renderTaskList renders the task list
func (a *App) renderTaskList() string {
	var builder strings.Builder

	builder.WriteString(TasksHeaderStyle.Render("Tasks"))
	builder.WriteString("\n")

	tasks := a.taskManager.FilteredTasks()
	if len(tasks) == 0 {
		builder.WriteString("No tasks. Add a new task with [N].")
		return builder.String()
	}

	for i, task := range tasks {
		isSelected := i == (a.selectedIndex % len(tasks))
		taskPrefix := fmt.Sprintf("%d", i+1)

		// Format the task line
		var taskLine string
		if isSelected {
			taskLine = "> "
		} else {
			taskLine = "  "
		}

		taskLine += taskPrefix + " "
		taskLine += "+task " + task.Description + " "

		// Join the main task text with the progress and time
		taskText := lipgloss.JoinHorizontal(lipgloss.Left,
			TaskStyle.Render(taskLine),
			TaskProgressStyle.Render(task.PomodoroProgress()),
			TaskTimeStyle.Render(" "+task.FormattedTimeSpent()),
		)

		builder.WriteString(taskText)
		builder.WriteString("\n")
	}

	return builder.String()
}

// renderTaskControls renders the task control buttons
func (a *App) renderTaskControls() string {
	var showHideText string
	if a.taskManager.ShowCompleted {
		showHideText = "Hide completed tasks"
	} else {
		showHideText = "Show completed tasks"
	}

	controls := lipgloss.JoinHorizontal(lipgloss.Left,
		HideCompletedStyle.Render(showHideText+" [H]"),
		lipgloss.NewStyle().Width(20).String(),
		AddNewTaskStyle.Render("Add new task [N]"),
	)

	return lipgloss.NewStyle().Margin(1, 0, 0, 0).Render(controls)
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
	a.debugMode = enabled
}
