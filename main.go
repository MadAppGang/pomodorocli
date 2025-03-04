package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jackrudenko/pomodorocli/ui"
)

func main() {
	// Define command-line flags
	timerOnly := flag.Bool("timer", false, "Show only the timer component")
	tasksOnly := flag.Bool("tasks", false, "Show only the task list component")
	showHelp := flag.Bool("help", false, "Show help information")

	// Parse command-line flags
	flag.Parse()

	// Show help text if requested
	if *showHelp {
		fmt.Println("Pomodoro CLI - A terminal-based Pomodoro timer")
		fmt.Println("\nUsage:")
		fmt.Println("  pomodorocli [options]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	// Create a new application
	app := ui.NewApp()

	// Set timer-only mode if requested via command-line flag
	if *timerOnly {
		app.SetTimerOnlyMode(true)
	} else if *tasksOnly {
		// Set task-list-only mode if requested
		app.SetTaskListOnlyMode(true)
	}

	// Create a new bubble tea program
	p := tea.NewProgram(app, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
