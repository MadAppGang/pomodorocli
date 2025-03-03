package ui

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// GetTerminalWidth returns the current terminal width
func GetTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		// Fallback to a reasonable default if we can't detect
		return 100
	}
	return width
}

// GetTerminalHeight returns the current terminal height
func GetTerminalHeight() int {
	_, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || height <= 0 {
		// Fallback to a reasonable default if we can't detect
		return 30
	}
	return height
}

// Color scheme based on the Figma design
var (
	// Background colors
	ColorBackground    = lipgloss.Color("#121416")
	ColorBoxBackground = lipgloss.Color("#09090A")
	ColorBorder        = lipgloss.Color("#222528")

	// Text colors
	ColorText          = lipgloss.Color("#FFFFFF")
	ColorProgressBar   = lipgloss.Color("#808183")
	ColorStopButton    = lipgloss.Color("#BB566B")
	ColorTaskTag       = lipgloss.Color("#9485D7")
	ColorTasksHeader   = lipgloss.Color("#7BC0AB")
	ColorHideCompleted = lipgloss.Color("#C1B476")
	ColorAddNewTask    = lipgloss.Color("#474433")
)

// Styles for different UI components
var (
	// Get the terminal width
	termWidth = GetTerminalWidth()

	// Base app container - takes full width, minus a small margin
	AppStyle = lipgloss.NewStyle().
			Background(ColorBackground).
			Padding(1, 2).
			Width(termWidth - 4)

	// Main content container - 4 chars less than app width for borders/padding
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true).
			Background(ColorBoxBackground).
			Padding(1, 2).
			Width(termWidth - 8)

	// Title style
	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Bold(true).
			Align(lipgloss.Center)

	// App name style
	AppNameStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			MarginBottom(1)

	// Timer display - using large text style
	TimerStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Bold(true).
			Align(lipgloss.Center).
			Padding(0, 0, 1, 0)

	// Progress bar
	ProgressBarStyle = lipgloss.NewStyle().
				Foreground(ColorProgressBar).
				Bold(true).
				Align(lipgloss.Center).
				MarginBottom(1)

	// Current task display
	CurrentTaskStyle = lipgloss.NewStyle().
				Foreground(ColorText).
				Align(lipgloss.Center).
				MarginBottom(1)

	// Stop button
	StopButtonStyle = lipgloss.NewStyle().
			Foreground(ColorStopButton).
			Bold(true).
			Align(lipgloss.Center).
			MarginBottom(1)

	// Tasks header
	TasksHeaderStyle = lipgloss.NewStyle().
				Foreground(ColorTasksHeader).
				MarginTop(1)

	// Task style
	TaskStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	// Task progress style
	TaskProgressStyle = lipgloss.NewStyle().
				Foreground(ColorTaskTag)

	// Task time style
	TaskTimeStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	// Hide completed tasks
	HideCompletedStyle = lipgloss.NewStyle().
				Foreground(ColorHideCompleted).
				Bold(true)

	// Add new task
	AddNewTaskStyle = lipgloss.NewStyle().
			Foreground(ColorAddNewTask).
			Bold(true)

	// Menu item style
	MenuItemStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			PaddingLeft(1).
			PaddingRight(1)

	// Divider style - match box width
	DividerStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Width(termWidth - 16) // Account for borders and padding
)

// UpdateStyles updates all styles with the current terminal dimensions
func UpdateStyles() {
	termWidth = GetTerminalWidth()

	AppStyle = AppStyle.Width(termWidth - 4)
	BoxStyle = BoxStyle.Width(termWidth - 8)
	DividerStyle = DividerStyle.Width(termWidth - 16)
}

// Renders a styled progress bar
func RenderProgressBar(percentage float64) string {
	// Calculate a width that scales with the terminal width
	width := (termWidth - 40) / 2 // Scale dynamically but keep it reasonable
	if width < 20 {
		width = 20 // Minimum size
	}

	filledWidth := int(percentage * float64(width) / 100.0)

	// Make sure we don't exceed boundaries
	if filledWidth < 0 {
		filledWidth = 0
	} else if filledWidth > width {
		filledWidth = width
	}

	// Create the progress bar string
	var progress string
	progress = "["

	// Add the filled part
	for i := 0; i < width; i++ {
		if i < filledWidth {
			progress += "="
		} else if i == filledWidth {
			progress += "🍅"
		} else {
			progress += "-"
		}
	}

	progress += "]"

	return ProgressBarStyle.Render(progress)
}
