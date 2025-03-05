package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jackrudenko/pomodorocli/model"
)

const (
	maxWidth = 120
	minWidth = 80
)

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// TimerView represents the timer component
type TimerView struct {
	timer       *model.Timer
	width       int
	fontManager *FontManager
}

// NewTimerView creates a new timer view
func NewTimerView(timer *model.Timer, width int) *TimerView {
	width = clamp(width, minWidth, maxWidth)
	return &TimerView{
		timer:       timer,
		width:       width,
		fontManager: nil,
	}
}

// SetWidth updates the width of the timer view
func (t *TimerView) SetWidth(width int) {
	width = clamp(width, minWidth, maxWidth)
	t.width = width
}

// SetFontManager sets the font manager for rendering big digits
func (t *TimerView) SetFontManager(fontManager *FontManager) {
	t.fontManager = fontManager
}

// Render renders the timer view
func (t *TimerView) Render() string {
	// Display font information if font manager is available
	var fontInfo string
	if t.fontManager != nil {
		fontInfo = lipgloss.NewStyle().
			Foreground(ColorGrayText).
			Align(lipgloss.Center).
			PaddingBottom(1).
			Render("Font: " + t.fontManager.CurrentFont + " [F]")
	}

	// Render each component without background
	currentTask := t.renderCurrentTask()
	timer := t.renderTimer()
	progressBar := t.renderProgressBar()
	controls := t.renderControls()

	// Join the timer section elements vertically with minimal spacing
	var components []string
	if fontInfo != "" {
		components = append(components, fontInfo)
	}

	// Compact rendering of components
	components = append(components, currentTask, timer, progressBar, controls)

	// Join all components with center alignment and no extra padding
	timerContent := lipgloss.NewStyle().
		Padding(0, 0).
		Align(lipgloss.Center).
		Render(lipgloss.JoinVertical(lipgloss.Center, components...))

	return timerContent
}

// renderCurrentTask renders the current active task (if any)
func (t *TimerView) renderCurrentTask() string {
	// For break modes, show a different message
	if t.timer.Mode == model.ShortBreakMode || t.timer.Mode == model.LongBreakMode {
		breakType := "Short break"
		if t.timer.Mode == model.LongBreakMode {
			breakType = "Long break"
		}

		// Use a teal/blue color for breaks
		breakStyle := CurrentTaskStyle.Copy().
			Foreground(lipgloss.Color("#7BC0AB"))

		return breakStyle.
			PaddingBottom(1).
			Render(breakType + " - Take a rest")
	}

	// Standard task display for focus mode
	if t.timer.CurrentTaskID != "" && t.timer.State == model.TimerRunning {
		// Get the current task from the task manager
		task, found := t.timer.TaskManager.GetTask(t.timer.CurrentTaskID)
		if found {
			return CurrentTaskStyle.
				PaddingBottom(1).
				Render(TaskProgressStyle.Render("+task ") + task.Description)
		}
	}
	return CurrentTaskStyle.
		PaddingBottom(1).
		Render("Select a task to start")
}

// renderTimer renders the timer display using large ASCII characters
func (t *TimerView) renderTimer() string {
	timeStr := t.timer.FormatTime() // Format like "25:00"

	// Use a different color for break modes
	timerStyle := TimerStyle
	if t.timer.Mode == model.ShortBreakMode || t.timer.Mode == model.LongBreakMode {
		// Use a teal/blue color for breaks
		timerStyle = timerStyle.Copy().Foreground(lipgloss.Color("#7BC0AB"))
	} else {
		// Use the default white color for focus mode
		timerStyle = timerStyle.Copy().Foreground(ColorText)
	}

	// If we have a font manager, use it to render the time string
	if t.fontManager != nil {
		// Render the time with the current font
		bigTimeStr := t.fontManager.RenderTimeString(timeStr)
		return timerStyle.Background(nil).Render(bigTimeStr)
	}

	// Fallback to regular string
	return timerStyle.Background(nil).Render(timeStr)
}

// renderProgressBar renders the timer progress bar
func (t *TimerView) renderProgressBar() string {
	// Build the progress bar with two colored segments:
	// The completed (left) portion is white,
	// and the remaining (right) portion is colored #808183.
	return ProgressBarStyle.Background(nil).Render(t.buildProgressBar(t.timer.ProgressPercentage()))
}

// buildProgressBar creates the progress bar string without styling as a method on TimerView using its width
func (t *TimerView) buildProgressBar(percentage float64) string {
	// Calculate a width that scales with TimerView width
	progressBarWidth := clamp(t.width-40, 20, GetTerminalWidth()-20)

	filledWidth := int(percentage * float64(progressBarWidth) / 100.0)
	if filledWidth < 0 {
		filledWidth = 0
	} else if filledWidth > progressBarWidth {
		filledWidth = progressBarWidth
	}

	// Define styles for the completed (left) and remaining (right) segments.
	leftStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("white"))
	rightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#808183"))

	// Render opening bracket in white using leftStyle
	progress := leftStyle.Render("[")
	if filledWidth == progressBarWidth {
		// Full progress: show the complete bar in white
		progress += leftStyle.Render(strings.Repeat("=", filledWidth))
	} else {
		// Left segment: completed progress, rendered in white.
		leftSegment := leftStyle.Render(strings.Repeat("=", filledWidth))
		// Marker: tomato emoji, rendered in white.
		marker := leftStyle.Render("🍅")
		// Right segment: remaining progress, rendered in #808183.
		remainingLength := progressBarWidth - filledWidth - 1
		if remainingLength < 0 {
			remainingLength = 0
		}
		rightSegment := rightStyle.Render(strings.Repeat("-", remainingLength))
		progress += leftSegment + marker + rightSegment
	}
	// Render closing bracket in white using leftStyle
	progress += leftStyle.Render("]")

	return progress
}

// renderControls renders the timer control buttons
func (t *TimerView) renderControls() string {
	var controls string

	// Base control - Start/Stop
	if t.timer.State == model.TimerRunning {
		controls = StopButtonStyle.Background(nil).Render("Stop [S]")
	} else {
		controls = StopButtonStyle.Background(nil).Render("Start [S]")
	}

	// Add Skip button during breaks
	if t.timer.Mode == model.ShortBreakMode || t.timer.Mode == model.LongBreakMode {
		skipStyle := StopButtonStyle.Copy().
			Foreground(lipgloss.Color("#7BC0AB"))

		skipButton := skipStyle.Render("   Skip Break [B]")
		controls = lipgloss.JoinHorizontal(lipgloss.Center, controls, skipButton)
	}

	return controls
}
