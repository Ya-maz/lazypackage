package ui

import (
	"lazypackage/msg"
	"lazypackage/node"
	"fmt"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

// model stores output lines and the current navigation state
type model struct {
	lines    []string
	screen   string // "menu" or "execution"
	stack    []Stage
	selected string // Name of the script currently running
}

func InitialModel() model {
	scripts := node.ReadScriptsFromPackageJSON()
	symbols := node.ReadSymbolsFromJSON()

	// Initializing the first stage (Main Menu)
	mainStage := Stage{
		Title:    "Main Menu",
		Subtitle: "Available npm scripts:",
		Options:  []Option{},
		Cursor:   0,
	}

	keys := make([]string, 0, len(scripts))
	for k := range scripts {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		mainStage.Options = append(mainStage.Options, Option{
			Name:        k,
			Icon:        symbols[k],
			Description: scripts[k],
			Script:      k,
		})
	}

	// Example of nesting: If we have "test", let's make it a sub-menu instead
	if _, ok := scripts["test"]; ok {
		testSubStage := Stage{
			Title:    "Test Suite",
			Subtitle: "Which tests to run?",
			Options:  []Option{
				{Name: "Run All Tests", Icon: "🧪", Script: "test"},
				{Name: "Run Lint", Icon: "🧹", Script: "lint"},
				{Name: "Back to Main Menu", Icon: "⬅️", SubStage: nil},
			},
		}
		
		for i, opt := range mainStage.Options {
			if opt.Name == "test" {
				mainStage.Options[i].Script = ""
				mainStage.Options[i].SubStage = &testSubStage
				break
			}
		}
	}

	return model{
		lines:  []string{},
		stack:  []Stage{mainStage},
		screen: "menu",
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	
	case msg.NodeDone:
		m.screen = "menu"
		m.lines = append(m.lines, successStyle.Render(fmt.Sprintf("✅ Script '%s' finished!", m.selected)))
		// Optionally trigger a redraw or similar if needed, but bubbletea handles text updates
		return m, nil

	case msg.NodeErr:
		m.screen = "menu"
		m.lines = append(m.lines, errorStyle.Render(fmt.Sprintf("❌ Script '%s' failed: %s", m.selected, string(message))))
		return m, nil

	case tea.KeyMsg:
		// If execution is transitioning, ignore keys
		if m.screen == "execution" {
			return m, nil
		}

		currentStageIdx := len(m.stack) - 1
		if currentStageIdx < 0 {
			return m, tea.Quit
		}
		currentStage := &m.stack[currentStageIdx]

		switch message.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "up", "k":
			if currentStage.Cursor > 0 {
				currentStage.Cursor--
			}

		case "down", "j":
			if currentStage.Cursor < len(currentStage.Options)-1 {
				currentStage.Cursor++
			}

		case "backspace", "esc", "h":
			if len(m.stack) > 1 {
				m.stack = m.stack[:len(m.stack)-1]
			}

		case "enter", "l":
			selectedOpt := currentStage.Options[currentStage.Cursor]
			
			if selectedOpt.Name == "Back to Main Menu" || (selectedOpt.SubStage == nil && selectedOpt.Script == "") {
				if len(m.stack) > 1 {
					m.stack = m.stack[:len(m.stack)-1]
					return m, nil
				}
			}

			if selectedOpt.SubStage != nil {
				m.stack = append(m.stack, *selectedOpt.SubStage)
				return m, nil
			}

			if selectedOpt.Script != "" {
				m.selected = selectedOpt.Script
				m.screen = "execution"
				
				// Get the exec.Cmd configured for the script
				c := node.GetNodeCommand(m.selected)
				
				// Execute using tea.ExecProcess which manages terminal ownership
				return m, tea.ExecProcess(c, func(err error) tea.Msg {
					if err != nil {
						return msg.NodeErr(err.Error())
					}
					return msg.NodeDone{}
				})
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.screen == "execution" {
		// This text might briefly appear before tea.ExecProcess takes over terminal
		return "\n  🚀 Launching " + m.selected + "...\n"
	}

	// Render current stage
	currentStageIdx := len(m.stack) - 1
	if currentStageIdx < 0 {
		return "Closing..."
	}
	currentStage := m.stack[currentStageIdx]

	title := titleStyle.Render(currentStage.Title)
	subtitle := subTitleStyle.Render(currentStage.Subtitle)

	items := []string{}
	for i, opt := range currentStage.Options {
		prefix := itemNumberStyle.Render(fmt.Sprintf("[ %d ]", i+1))
		name := itemTextStyle.Render(opt.Name)
		icon := successStyle.Render(opt.Icon)
		line := fmt.Sprintf("%s %s %s", prefix, name, icon)
		if i == currentStage.Cursor {
			line = selectStyle.Render(line)
		}
		items = append(items, line)
	}

	contentLines := []string{
		title,
		"",
		subtitle,
		"",
	}
	contentLines = append(contentLines, items...)
	
	// Show recent status messages at the bottom
	if len(m.lines) > 0 {
		contentLines = append(contentLines, "", subTitleStyle.Render("Last run status:"))
		
		// Show last 3 lines
		start := len(m.lines) - 3
		if start < 0 { start = 0 }
		for _, l := range m.lines[start:] {
			contentLines = append(contentLines, l)
		}
	}
	
	contentLines = append(contentLines, "", hintStyle.Render("q/ctrl+c: quit | esc: back | enter: select"))

	content := lipgloss.JoinVertical(lipgloss.Left, contentLines...)
	return boxStyle.Render(content)
}
