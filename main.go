package main

import (
	"bufio"
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// model stores output lines and the channel to read from
type model struct {
	lines []string
	ch    <-chan tea.Msg
}

// Msg types
type nodeLineMsg string
type nodeDoneMsg struct{}
type nodeErrMsg string

// --- Styles ---
var (
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00BFFF")). // Голубая рамка
			Padding(1, 2).
			Margin(1)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00BFFF")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555"))
)

// initial model (channel assigned later)
func initialModel(ch <-chan tea.Msg) model {
	return model{
		lines: []string{"🚀 Starting Node..."},
		ch:    ch,
	}
}

// Init returns a Cmd that reads the first message from channel
func (m model) Init() tea.Cmd {
	return readNextLine(m.ch)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case nodeLineMsg:
		m.lines = append(m.lines, string(msg))
		// после обработки строки — снова ждать следующую
		return m, readNextLine(m.ch)
	case nodeDoneMsg:
		m.lines = append(m.lines, "✅ Node finished!")
		// можно завершать программу или оставить (завершим)
		return m, tea.Quit
	case nodeErrMsg:
		m.lines = append(m.lines, fmt.Sprintf("❌ Error: %s", string(msg)))
		return m, tea.Quit
	case tea.KeyMsg:
		// выход по q
		if msg.String() == "q" {
			return m, tea.Quit
		}
	}
	return m, nil
}

//	func (m model) View() string {
//		s := "📦 Node Process Output\n\n"
//		for _, line := range m.lines {
//			s += line + "\n"
//		}
//		s += "\n(press q to quit)\n"
//		return s
//	}

func (m model) View() string {
	header := titleStyle.Render("📦 Node Process Output")
	body := ""

	for _, line := range m.lines {
		runes := []rune(line)
		if len(runes) == 0 {
			continue
		}

		switch runes[0] {
		case '✅':
			body += successStyle.Render(line) + "\n"
		case '❌':
			body += errorStyle.Render(line) + "\n"
		default:
			body += line + "\n"
		}
	}

	footer := "\n(press q to quit)"
	content := fmt.Sprintf("%s\n\n%s%s", header, body, footer)

	return boxStyle.Render(content)
}

// readNextLine возвращает Cmd, которая ждёт следующего сообщения из ch
func readNextLine(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		if ch == nil {
			return nodeErrMsg("internal: channel is nil")
		}
		msg, ok := <-ch
		if !ok {
			// канал закрыт — пришёл done
			return nodeDoneMsg{}
		}
		return msg
	}
}

// startNodeProcess запускает node и возвращает канал, в который будут приходить Msg
func startNodeProcess() (<-chan tea.Msg, error) {
	cmd := exec.Command("node", "script.js")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	outCh := make(chan tea.Msg)

	// горутина читает stdout и шлёт nodeLineMsg в канал
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			outCh <- nodeLineMsg(line)
		}
		// дождёмся завершения процесса
		cmd.Wait()
		close(outCh)
	}()

	return outCh, nil
}

// normal !!!
func main() {
	// запускаем node до создания программы, получаем канал
	ch, err := startNodeProcess()
	if err != nil {
		fmt.Println("Failed to start node:", err)
		return
	}

	// создаём программу с моделью, у которой уже есть канал
	p := tea.NewProgram(initialModel(ch))

	// запускаем (Run блокирует)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
	}
}
