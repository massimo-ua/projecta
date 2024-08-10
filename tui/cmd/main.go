package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"

	"projecta/tui"
)

func main() {
	p := tea.NewProgram(tui.NewModel())
	if err, _ := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
