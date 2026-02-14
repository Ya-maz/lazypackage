package ui

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.screen == "execution" {
		return "\n  🚀 Launching " + m.selected + "...\n"
	}

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

	if len(m.lines) > 0 {
		contentLines = append(contentLines, "", subTitleStyle.Render("Last run status:"))

		start := len(m.lines) - 3
		if start < 0 {
			start = 0
		}
		for _, l := range m.lines[start:] {
			contentLines = append(contentLines, l)
		}
	}

	contentLines = append(contentLines, "", hintStyle.Render("q/ctrl+c: quit | esc: back | enter: select"))

	content := lipgloss.JoinVertical(lipgloss.Left, contentLines...)
	return boxStyle.Render(content)
}
