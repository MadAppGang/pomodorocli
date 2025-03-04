package ui

import (
	"fmt"
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
	// Create a consistent style for all timer elements
	baseStyle := lipgloss.NewStyle().
		Background(ColorBoxBackground).
		Width(t.width - 12)

	// Apply this style to each element
	currentTask := baseStyle.Align(lipgloss.Center).Render(t.renderCurrentTask())
	timer := baseStyle.Align(lipgloss.Center).Render(t.renderTimer())
	progressBar := baseStyle.Align(lipgloss.Center).Render(t.renderProgressBar())
	controls := baseStyle.Align(lipgloss.Center).Render(t.renderControls())

	// Join the timer section elements with consistent background
	timerContent := lipgloss.JoinVertical(lipgloss.Center,
		currentTask,
		timer,
		progressBar,
		controls,
		fmt.Sprintf("width: %d", t.width),
	)

	// The container should already have a consistent background from the joined elements
	return timerContent
}

// renderCurrentTask renders the current active task (if any)
func (t *TimerView) renderCurrentTask() string {
	if t.timer.CurrentTask != nil {
		return CurrentTaskStyle.Background(nil).Render("+task " + t.timer.CurrentTask.Description)
	}
	return CurrentTaskStyle.Background(nil).Render("Select a task to start")
}

// renderTimer renders the timer display using large ASCII characters
func (t *TimerView) renderTimer() string {
	timeStr := t.timer.FormatTime() // Format like "25:00"

	// If we have a font manager, use it to render the time string
	if t.fontManager != nil {
		// Get the currently active font name for display
		currentFont := t.fontManager.CurrentFont

		// Render the time with the current font
		bigTimeStr := t.fontManager.RenderTimeString(timeStr)

		// Add the font name at the bottom of the timer
		fontInfo := fmt.Sprintf("Font: %s (Press F to change)", currentFont)
		return TimerStyle.Background(nil).Render(bigTimeStr + "\n" + fontInfo)
	}

	// Fallback to hardcoded big digits if no font manager is available
	bigTimeStr := renderBigDigits(timeStr)
	return TimerStyle.Background(nil).Render(bigTimeStr)
}

// Define the big digits patterns for ASCII art rendering
var bigDigits = map[rune][]string{
	'0': {
		"  ░░░░  ",
		" ░    ░ ",
		"░      ░",
		"░      ░",
		"░      ░",
		"░      ░",
		" ░    ░ ",
		"  ░░░░  ",
	},
	'1': {
		"   ░░   ",
		"  ░░░   ",
		" ░ ░░   ",
		"   ░░   ",
		"   ░░   ",
		"   ░░   ",
		"   ░░   ",
		" ░░░░░░ ",
	},
	'2': {
		" ░░░░░░ ",
		"░      ░",
		"       ░",
		"      ░ ",
		"    ░░  ",
		"  ░░    ",
		" ░      ",
		"░░░░░░░░",
	},
	'3': {
		" ░░░░░░ ",
		"░      ░",
		"       ░",
		"   ░░░░ ",
		"   ░░░░ ",
		"       ░",
		"░      ░",
		" ░░░░░░ ",
	},
	'4': {
		"     ░░ ",
		"    ░░░ ",
		"   ░ ░░ ",
		"  ░  ░░ ",
		" ░   ░░ ",
		"░░░░░░░░",
		"     ░░ ",
		"     ░░ ",
	},
	'5': {
		"░░░░░░░░",
		"░       ",
		"░       ",
		"░░░░░░░ ",
		"       ░",
		"       ░",
		"░      ░",
		" ░░░░░░ ",
	},
	'6': {
		"  ░░░░░ ",
		" ░      ",
		"░       ",
		"░░░░░░░ ",
		"░      ░",
		"░      ░",
		"░      ░",
		" ░░░░░░ ",
	},
	'7': {
		"░░░░░░░░",
		"      ░ ",
		"     ░  ",
		"    ░   ",
		"   ░    ",
		"  ░     ",
		" ░      ",
		"░       ",
	},
	'8': {
		" ░░░░░░ ",
		"░      ░",
		"░      ░",
		" ░░░░░░ ",
		" ░░░░░░ ",
		"░      ░",
		"░      ░",
		" ░░░░░░ ",
	},
	'9': {
		" ░░░░░░ ",
		"░      ░",
		"░      ░",
		"░      ░",
		" ░░░░░░░",
		"       ░",
		"      ░ ",
		" ░░░░░  ",
	},
	':': {
		"        ",
		"   ░░   ",
		"   ░░   ",
		"        ",
		"        ",
		"   ░░   ",
		"   ░░   ",
		"        ",
	},
}

// renderBigDigits converts a time string (e.g. "25:00") into large ASCII art
func renderBigDigits(timeStr string) string {
	// Number of lines in each digit pattern
	lines := 8

	// Initialize result lines
	result := make([]string, lines)

	// Process each character in the time string
	for _, char := range timeStr {
		// Get the digit pattern (default to empty if not found)
		pattern, ok := bigDigits[char]
		if !ok {
			pattern = make([]string, lines)
			for i := range pattern {
				pattern[i] = "        "
			}
		}

		// Add this digit's pattern to each line of the result
		for i := 0; i < lines; i++ {
			result[i] += pattern[i]
		}
	}

	// Join all lines with newlines
	return strings.Join(result, "\n")
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
	progressBarWidth := t.width - 40
	if progressBarWidth < 20 {
		progressBarWidth = 20 // Minimum size
	}

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
	if t.timer.State == model.TimerRunning {
		return StopButtonStyle.Background(nil).Render("Stop [S]")
	}
	return StopButtonStyle.Background(nil).Render("Start [S]")
}
