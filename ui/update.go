package ui

import (
	"fmt"
	"lazypackage/msg"
	"lazypackage/node"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {

	case msg.NodeDone:
		m.screen = "menu"
		m.lines = append(m.lines, successStyle.Render(fmt.Sprintf("✅ Script '%s' finished!", m.selected)))
		return m, nil

	case msg.NodeErr:
		m.screen = "menu"
		m.lines = append(m.lines, errorStyle.Render(fmt.Sprintf("❌ Script '%s' failed: %s", m.selected, string(message))))
		return m, nil

	case tea.KeyMsg:
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

				c := node.GetNodeCommand(m.selected)

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
