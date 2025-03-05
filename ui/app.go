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
	timer           *model.Timer
	taskManager     *model.TaskManager
	settingsManager *model.SettingsManager
	storageManager  *storage.StorageManager
	view            ViewState
	width           int
	height          int

	// Input fields for adding tasks
	taskInput      textinput.Model
	pomodorosInput textinput.Model
	inputting      bool

	// Input fields for settings
	pomodoroDurationInput   textinput.Model
	shortBreakDurationInput textinput.Model
	longBreakDurationInput  textinput.Model

	// Components
	timerView    *TimerView
	taskListView *TaskListView

	// Debug mode
	debugMode DebugMode

	// Font manager for rendering big digits
	fontManager *FontManager

	// UI control flags
	showHelpText bool
}

// NewApp creates a new application model
func NewApp() *App {
	// Initialize task inputs
	taskInput := textinput.New()
	taskInput.Placeholder = "Task description"
	taskInput.Width = 60
	taskInput.Focus()

	pomodorosInput := textinput.New()
	pomodorosInput.Placeholder = "Number of pomodoros (default: 4)"
	pomodorosInput.Width = 10

	// Initialize settings inputs
	pomodoroDurationInput := textinput.New()
	pomodoroDurationInput.Placeholder = "Pomodoro duration (minutes)"
	pomodoroDurationInput.Width = 10

	shortBreakDurationInput := textinput.New()
	shortBreakDurationInput.Placeholder = "Short break duration (minutes)"
	shortBreakDurationInput.Width = 10

	longBreakDurationInput := textinput.New()
	longBreakDurationInput.Placeholder = "Long break duration (minutes)"
	longBreakDurationInput.Width = 10

	width := GetTerminalWidth()
	height := GetTerminalHeight()

	// Initialize model objects
	settingsManager := model.NewSettingsManager()
	timer := model.NewTimer()
	taskManager := model.NewTaskManager()

	// Set the timer to use the settings
	timer.SetSettings(&settingsManager.Settings)

	// Initialize storage
	jsonStorage, err := storage.NewJSONTaskStorage("./data/tasks.json")
	var storageManager *storage.StorageManager
	if err == nil {
		// Now jsonStorage implements both TaskStorage and SettingsStorage
		storageManager = storage.NewStorageManager(jsonStorage, jsonStorage, taskManager, &settingsManager.Settings)

		// Load tasks from storage
		if err := storageManager.LoadTasks(); err != nil {
			// If loading fails, we'll start with an empty task list
			fmt.Println("Error loading tasks:", err)
		}

		// Load settings from storage
		if err := storageManager.LoadSettings(); err != nil {
			// If loading fails, we'll use default settings
			fmt.Println("Error loading settings:", err)
		} else {
			// Debug: Print loaded settings
			fmt.Printf("Loaded settings - Pomodoro: %d, Short break: %d, Long break: %d\n",
				settingsManager.Settings.PomodoroDuration,
				settingsManager.Settings.ShortBreakDuration,
				settingsManager.Settings.LongBreakDuration)

			// Make sure the timer is updated with the loaded settings
			timer.SetSettings(&settingsManager.Settings)

			// Explicitly reset the timer to ensure it uses the loaded duration
			timer.Reset()
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
		timer:                   timer,
		taskManager:             taskManager,
		settingsManager:         settingsManager,
		storageManager:          storageManager,
		view:                    MainView,
		width:                   width,
		height:                  height,
		taskInput:               taskInput,
		pomodorosInput:          pomodorosInput,
		pomodoroDurationInput:   pomodoroDurationInput,
		shortBreakDurationInput: shortBreakDurationInput,
		longBreakDurationInput:  longBreakDurationInput,
		inputting:               false,
		debugMode:               NoDebug,
		fontManager:             fontManager,
		showHelpText:            false, // Show help text by default
	}

	// Initialize components
	app.timerView = NewTimerView(timer, width)
	app.taskListView = NewTaskListView(taskManager, width)

	// Set the font manager in the timer view
	if fontManager != nil {
		app.timerView.SetFontManager(fontManager)
	}

	// Register settings change handler to update timer
	settingsManager.RegisterChangeHandler(func() {
		// Update timer with new settings
		timer.SetSettings(&settingsManager.Settings)

		// Explicitly reset the timer when settings change
		timer.Reset()

		// Debug: Print settings after change
		fmt.Printf("Settings changed - Pomodoro: %d, Short break: %d, Long break: %d\n",
			settingsManager.Settings.PomodoroDuration,
			settingsManager.Settings.ShortBreakDuration,
			settingsManager.Settings.LongBreakDuration)

		// Save settings on change
		if storageManager != nil {
			_ = storageManager.SaveSettings()
		}
	})

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

	case "S", "s":
		// Toggle between start and pause without resetting
		if a.timer.State == model.TimerRunning {
			a.timer.Pause()
		} else {
			// Will resume if paused, or start if stopped
			a.timer.Start()
		}

	case "R", "r":
		// Reset timer to full duration
		a.timer.Reset()

	case "N", "n":
		// Add new task
		a.view = AddTaskView
		a.taskInput.Focus()
		a.inputting = true

	case "H", "h":
		// Toggle hiding completed tasks
		a.taskListView.ToggleShowCompleted()

	case "J", "j", "down":
		// Move down in task list
		a.taskListView.MoveSelectionDown()

	case "K", "k", "up":
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

	case "D", "d":
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

	case "O", "o":
		// Open settings
		a.view = SettingsView

	case "M", "m":
		// Cycle through debug modes: NoDebug -> TimerDebug -> TaskListDebug -> NoDebug
		a.debugMode = (a.debugMode + 1) % 3

	case "F", "f":
		// Toggle through fonts if font manager is available
		if a.fontManager != nil {
			a.fontManager.NextFont()
		}

	case "?":
		// Toggle help text visibility
		a.showHelpText = !a.showHelpText
	}

	return a, nil
}

// updateAddTaskView handles input for the add task view
func (a *App) updateAddTaskView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "ctrl+c", "q":
		return a, tea.Quit

	case "esc":
		// Cancel and return to main view
		a.view = MainView
		a.taskInput.Blur()
		a.pomodorosInput.Blur()
		a.taskInput.SetValue("")
		a.pomodorosInput.SetValue("")
		return a, nil

	case "?": // Toggle help text visibility
		a.showHelpText = !a.showHelpText
		return a, nil

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
	var cmd tea.Cmd

	switch msg.String() {
	case "ctrl+c", "q":
		return a, tea.Quit

	case "esc", "o":
		// Save settings on exit
		a.saveSettings()
		// Return to main view
		a.view = MainView
		return a, nil

	case "tab", "shift+tab":
		// Switch between inputs
		if a.pomodoroDurationInput.Focused() {
			a.pomodoroDurationInput.Blur()
			a.shortBreakDurationInput.Focus()
		} else if a.shortBreakDurationInput.Focused() {
			a.shortBreakDurationInput.Blur()
			a.longBreakDurationInput.Focus()
		} else {
			a.longBreakDurationInput.Blur()
			a.pomodoroDurationInput.Focus()
		}

	case "enter":
		// Save settings using the saveSettings method
		a.saveSettings()

		// Return to main view after saving
		a.view = MainView
		a.pomodoroDurationInput.Blur()
		a.shortBreakDurationInput.Blur()
		a.longBreakDurationInput.Blur()

	case "?": // Toggle help text visibility
		a.showHelpText = !a.showHelpText
		return a, nil
	}

	// Handle text input updates
	if a.pomodoroDurationInput.Focused() {
		a.pomodoroDurationInput, cmd = a.pomodoroDurationInput.Update(msg)
		return a, cmd
	} else if a.shortBreakDurationInput.Focused() {
		a.shortBreakDurationInput, cmd = a.shortBreakDurationInput.Update(msg)
		return a, cmd
	} else if a.longBreakDurationInput.Focused() {
		a.longBreakDurationInput, cmd = a.longBreakDurationInput.Update(msg)
		return a, cmd
	}

	return a, nil
}

// updateSettingsInputs updates the input fields with current settings values
func (a *App) updateSettingsInputs() {
	a.pomodoroDurationInput.SetValue(fmt.Sprintf("%d", a.settingsManager.Settings.PomodoroDuration))
	a.shortBreakDurationInput.SetValue(fmt.Sprintf("%d", a.settingsManager.Settings.ShortBreakDuration))
	a.longBreakDurationInput.SetValue(fmt.Sprintf("%d", a.settingsManager.Settings.LongBreakDuration))
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

	// Create inner box with rounded borders - adjust height based on help text visibility
	innerBoxHeight := a.height - 6
	if !a.showHelpText {
		// Expand the inner box when help text is hidden to use that space
		innerBoxHeight += 3 // Add space that would have been used by help text
	}

	innerBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(a.width - 10).
		Height(innerBoxHeight)

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

	// Help text section
	helpStyle := lipgloss.NewStyle().
		Foreground(ColorGrayText).
		Align(lipgloss.Center).
		PaddingTop(1)

	helpTextContent := ""
	if a.showHelpText {
		helpTextContent = helpStyle.Render(
			"\n[S] Start/Pause  [s] Stop  [r] Reset  [n] New Task  [o] Settings  [h] Toggle Completed  [Space] Toggle Selected  [Enter] Run Task  [Ctrl+C/q] Quit  [?] Hide Help")
	}

	return mainContainerStyle.Render(styledContent + helpTextContent + debugModeText)
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

// addTaskView renders the add task view
func (a *App) addTaskView() string {
	var builder strings.Builder

	builder.WriteString(TitleStyle.Render("Add New Task"))
	builder.WriteString("\n\n")

	builder.WriteString("Task Name:\n")
	builder.WriteString(a.taskInput.View())
	builder.WriteString("\n\n")

	builder.WriteString("Number of Pomodoros:\n")
	builder.WriteString(a.pomodorosInput.View())
	builder.WriteString("\n\n")

	// Instructions with help toggle
	if a.showHelpText {
		builder.WriteString("Press Enter to add, Esc to cancel, Tab to switch fields, ? to hide help")
	} else {
		builder.WriteString("Press ? to show help")
	}

	return BoxStyle.Render(builder.String())
}

// settingsView renders the settings view
func (a *App) settingsView() string {
	var builder strings.Builder

	builder.WriteString(TitleStyle.Render("Settings"))
	builder.WriteString("\n\n")

	// Initialize input values when opening the settings view
	if !a.pomodoroDurationInput.Focused() &&
		!a.shortBreakDurationInput.Focused() &&
		!a.longBreakDurationInput.Focused() {
		a.updateSettingsInputs()
		a.pomodoroDurationInput.Focus()
	}

	// Pomodoro Duration
	builder.WriteString(lipgloss.NewStyle().Bold(true).Render("Pomodoro Duration (minutes):"))
	builder.WriteString("\n")
	builder.WriteString(a.pomodoroDurationInput.View())
	builder.WriteString("\n\n")

	// Short Break Duration
	builder.WriteString(lipgloss.NewStyle().Bold(true).Render("Short Break Duration (minutes):"))
	builder.WriteString("\n")
	builder.WriteString(a.shortBreakDurationInput.View())
	builder.WriteString("\n\n")

	// Long Break Duration
	builder.WriteString(lipgloss.NewStyle().Bold(true).Render("Long Break Duration (minutes):"))
	builder.WriteString("\n")
	builder.WriteString(a.longBreakDurationInput.View())
	builder.WriteString("\n\n")

	// Instructions with help toggle
	if a.showHelpText {
		builder.WriteString("Press Enter to save, Esc to cancel, Tab to switch fields, ? to hide help")
	} else {
		builder.WriteString("Press ? to show help")
	}

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

// saveSettings saves the current settings via the storage manager
func (a *App) saveSettings() {
	// Apply current input values to settings
	if a.pomodoroDurationInput.Value() != "" {
		var minutes int
		fmt.Sscanf(a.pomodoroDurationInput.Value(), "%d", &minutes)
		if minutes > 0 {
			a.settingsManager.SetPomodoroDuration(minutes)
		}
	}

	if a.shortBreakDurationInput.Value() != "" {
		var minutes int
		fmt.Sscanf(a.shortBreakDurationInput.Value(), "%d", &minutes)
		if minutes > 0 {
			a.settingsManager.SetShortBreakDuration(minutes)
		}
	}

	if a.longBreakDurationInput.Value() != "" {
		var minutes int
		fmt.Sscanf(a.longBreakDurationInput.Value(), "%d", &minutes)
		if minutes > 0 {
			a.settingsManager.SetLongBreakDuration(minutes)
		}
	}

	// Save to storage
	if a.storageManager != nil {
		_ = a.storageManager.SaveSettings()
	}
}
