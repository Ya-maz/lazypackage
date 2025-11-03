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

type menu struct {
	scripts  map[string]string // список npm команд
	keys     []string
	symbols  map[string]string // список npm команд
	cursor   int               // какая команда выбрана
	selected string            // что выбрал пользователь
}

// model stores output lines and the channel to read from
type model struct {
	lines  []string
	ch     chan tea.Msg
	screen string
	menu
}

// initial model (channel assigned later)
func InitialModel() model {
	scripts := node.ReadScriptsFromPackageJSON()
	symbols := node.ReadSymbolsFromJSON()
	keys := make([]string, 0, len(scripts))
	for k := range scripts {
		keys = append(keys, k)
	}

	ch := make(chan tea.Msg)

	// ИСПРАВЛЕНО: Сортируем ключи для стабильного порядка
	sort.Strings(keys)

	return model{
		lines: []string{},
		ch:    ch,
		menu: menu{
			scripts:  scripts,
			keys:     keys, // Сохраняем отсортированный список
			symbols:  symbols,
			cursor:   0,
			selected: "",
		},
		screen: "menu",
	}
}

// Init returns a Cmd that reads the first message from channel
func (m model) Init() tea.Cmd {
	// return readNextLine(m.ch)
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
		switch m.screen {
		case "menu":
			switch message.String() {
			case "q", "ctrl+c":
				return m, tea.Quit

			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}

			case "down", "j":
				if m.cursor < len(m.keys)-1 {
					m.cursor++
				}

			case "enter", "l":
				m.selected = m.keys[m.cursor]
				m.screen = "finished"
				return m, tea.Batch(
					node.RunNodeScript(m.ch, m.selected),
					node.ReadNextLine(m.ch),
				)
			}
		case "finished":
			switch message.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	title := titleStyle.Render("Node Process Output")
	switch m.screen {
	case "finished":
		subtitle := subTitleStyle.Render("Running script: " + m.selected)
		selectedScript := runningScriptTextStyle.Render("> " + m.scripts[m.selected])

		allLogs := strings.Join(m.lines, "\n")
		logstring := itemTextStyle.Render(allLogs)

		// maxLogLines := 10
		// if len(m.lines) > maxLogLines {
		// 	start := len(m.lines) - maxLogLines
		// 	allLogs = strings.Join(m.lines[start:], "\n")
		// 	logstring = itemTextStyle.Render(allLogs)
		// }

		content := lipgloss.JoinVertical(lipgloss.Left,
			title,
			"",
			subtitle,
			selectedScript,
			"",
			logstring,
			"",
			hintStyle.Render(hint),
		)
		return boxStyle.Render(content)

	case "menu":
		subtitle := subTitleStyle.Render("Available npm scripts:")

		items := []string{}
		for i, script := range m.keys {
			prefix := itemNumberStyle.Render(fmt.Sprintf("[ %d ]", i+1))
			name := itemTextStyle.Render(script)
			icon := successStyle.Render(m.symbols[script])
			line := fmt.Sprintf("%s %s %s", prefix, name, icon)
			if i == m.cursor {
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

	return "Unknown screen state"
}
