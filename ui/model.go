package ui

import (
	"cli/msg"
	"cli/node"
	"fmt"
	"sort"
	"strings"

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
	ch       chan tea.Msg
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

	// For demonstration, let's group some scripts if they exist
	// e.g., if there are multiple test scripts, we could nest them
	// For simplicity now, we just populate the main menu
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
			Script:      k, // By default, everything is a script
		})
	}

	// Example of nesting: If we have "test", let's make it a sub-menu instead
	// (Check if "test", "lint", "build" etc exist to simulate nested questions)
	// This is just to WOO the user as requested.
	if _, ok := scripts["test"]; ok {
		// Mocking a sub-menu for tests
		testSubStage := Stage{
			Title:    "Test Suite",
			Subtitle: "Which tests to run?",
			Options:  []Option{
				{Name: "Run All Tests", Icon: "🧪", Script: "test"},
				{Name: "Run Lint", Icon: "🧹", Script: "lint"},
				{Name: "Back to Main Menu", Icon: "⬅️", SubStage: nil}, // Special case for back handle
			},
		}
		
		// Update "test" option in main menu to lead to sub-menu
		for i, opt := range mainStage.Options {
			if opt.Name == "test" {
				mainStage.Options[i].Script = ""
				mainStage.Options[i].SubStage = &testSubStage
				break
			}
		}
	}

	ch := make(chan tea.Msg)

	return model{
		lines:  []string{},
		ch:     ch,
		stack:  []Stage{mainStage},
		screen: "menu",
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case msg.NodeLine:
		m.lines = append(m.lines, string(message))
		return m, node.ReadNextLine(m.ch)

	case msg.NodeDone:
		m.lines = append(m.lines, successStyle.Render("✅ Process finished!"))
		return m, nil

	case msg.NodeErr:
		m.lines = append(m.lines, errorStyle.Render(fmt.Sprintf("❌ Error: %s", string(message))))
		return m, nil

	case tea.KeyMsg:
		if m.screen == "execution" {
			switch message.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "backspace", "esc": // Allow going back to menu after execution?
				m.screen = "menu"
				m.lines = []string{}
				return m, nil
			}
			return m, nil
		}

		// Menu navigation
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
			// Go back to previous stage
			if len(m.stack) > 1 {
				m.stack = m.stack[:len(m.stack)-1]
			}

		case "enter", "l":
			selectedOpt := currentStage.Options[currentStage.Cursor]
			
			// Option 1: It's a "Back" button (manual)
			if selectedOpt.Name == "Back to Main Menu" || (selectedOpt.SubStage == nil && selectedOpt.Script == "") {
				if len(m.stack) > 1 {
					m.stack = m.stack[:len(m.stack)-1]
					return m, nil
				}
			}

			// Option 2: Transition to SubStage
			if selectedOpt.SubStage != nil {
				m.stack = append(m.stack, *selectedOpt.SubStage)
				return m, nil
			}

			// Option 3: Run Script
			if selectedOpt.Script != "" {
				m.selected = selectedOpt.Script
				m.screen = "execution"
				m.ch = make(chan tea.Msg) // Create a fresh channel for this run
				return m, tea.Batch(
					node.RunNodeScript(m.ch, m.selected),
					node.ReadNextLine(m.ch),
				)
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	title := titleStyle.Render("Node Process Output")

	if m.screen == "execution" {
		subtitle := subTitleStyle.Render("Running script: " + m.selected)
		
		// Find the script description if possible (hacky for now, ideally stored in model)
		scriptCmd := ""
		if len(m.stack) > 0 {
			for _, opt := range m.stack[len(m.stack)-1].Options {
				if opt.Script == m.selected {
					scriptCmd = opt.Description
					break
				}
			}
		}
		
		selectedScript := runningScriptTextStyle.Render("> " + scriptCmd)
		
		// Truncate logs to last 20 lines
		displayLines := m.lines
		if len(displayLines) > 20 {
			displayLines = displayLines[len(displayLines)-20:]
		}
		
		allLogs := strings.Join(displayLines, "\n")
		logstring := itemTextStyle.Render(allLogs)

		content := lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			subtitle,
			selectedScript,
			"",
			logstring,
			"",
			hintStyle.Render("q/ctrl+c: quit | esc: back to menu"),
		)
		return boxStyle.Render(content)
	}

	// Render current stage
	currentStageIdx := len(m.stack) - 1
	if currentStageIdx < 0 {
		return "Closing..."
	}
	currentStage := m.stack[currentStageIdx]

	title = titleStyle.Render(currentStage.Title)
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
	contentLines = append(contentLines, "", hintStyle.Render(hint))

	content := lipgloss.JoinVertical(lipgloss.Left, contentLines...)
	return boxStyle.Render(content)
}
