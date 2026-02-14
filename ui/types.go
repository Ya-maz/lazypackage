package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Option represents a single item in a menu
type Option struct {
	Name        string
	Icon        string
	Description string // The actual command string from package.json
	Script      string // If set, selecting this option runs the script
	SubStage    *Stage // If set, selecting this option pushes a new stage
}

// Stage represents a single screen/menu
type Stage struct {
	Title    string
	Subtitle string
	Options  []Option
	Cursor   int
}

// Model stores output lines and the current navigation state
type Model struct {
	lines    []string
	screen   string // "menu" or "execution"
	stack    []Stage
	selected string // Name of the script currently running
}

func (m Model) Init() tea.Cmd {
	return nil
}
