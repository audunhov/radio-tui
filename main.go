package main

import (
	"fmt"
	"os"
	"radio-tui/player"
	"radio-tui/storage"
	"radio-tui/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := player.NewAudioService()
	s, err := storage.NewStorageService()
	if err != nil {
		fmt.Printf("Error initializing storage: %v\n", err)
		os.Exit(1)
	}

	m := tui.NewModel(p, s)
	prog := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := prog.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
