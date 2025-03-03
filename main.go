package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jackrudenko/pomodorocli/ui"
)

func main() {
	p := tea.NewProgram(ui.NewApp(), tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
		os.Exit(1)
	}
}
