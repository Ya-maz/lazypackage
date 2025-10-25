package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type menu struct {
	scripts  map[string]string // список npm команд
	keys     []string
	simbol   map[string]string // список npm команд
	cursor   int               // какая команда выбрана
	selected string            // что выбрал пользователь
}

// model stores output lines and the channel to read from
type model struct {
	lines []string
	ch    <-chan tea.Msg
	menu
}

// Msg types
type nodeLineMsg string
type nodeDoneMsg struct{}
type readpackageLineMsg string
type readpackageDoneMsg struct{}
type nodeErrMsg string

func colorize(colorNumber int, text string) string {
	return fmt.Sprintf("\033[1m\033[%dm%s\033[0m", colorNumber, text)
}

// --- Styles ---
var (
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00BFFF")). // Голубая рамка
			Padding(1, 2).
			Margin(1).
			Width(50). // фиксируем ширину
			Height(15) // фиксируем высоту

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00BFFF")).
			Bold(true)

	subTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1E90FF")). // чуть темнее основного
			Italic(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B"))

	selectStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#32CD32")).
			Italic(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555"))
)

// initial model (channel assigned later)
// func initialModel(ch <-chan tea.Msg) model {
func initialModel() model {
	scripts := loadScriptsFromPackageJSON()
	simbol := loadSimbolFromJSON()
	keys := make([]string, 0, len(scripts))
	for k := range scripts {
		keys = append(keys, k)
	}
	return model{
		lines: []string{"🚀 Starting Node..."},
		menu: menu{
			scripts:  scripts,
			keys:     keys,
			simbol:   simbol,
			cursor:   0,
			selected: "",
		},
	}
}

// Init returns a Cmd that reads the first message from channel
func (m model) Init() tea.Cmd {
	// return readNextLine(m.ch)
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// case nodeLineMsg:
	// 	m.lines = append(m.lines, string(msg))
	// 	// после обработки строки — снова ждать следующую
	// 	return m, readNextLine(m.ch)
	// case nodeDoneMsg:
	// 	m.lines = append(m.lines, "✅ Node finished!")
	// 	// можно завершать программу или оставить (завершим)
	// 	return m, tea.Quit
	// case nodeErrMsg:
	// 	m.lines = append(m.lines, fmt.Sprintf("❌ Error: %s", string(msg)))
	// 	return m, tea.Quit
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "up":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down":
			if m.cursor < len(m.scripts)-1 {
				m.cursor++
			}

		case "enter":
			// выбираем текущий скрипт
			keys := m.getKeys()
			m.selected = keys[m.cursor]
			// просто показываем выбранный
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) getKeys() []string {
	keys := make([]string, 0, len(m.scripts))
	for k := range m.menu.scripts {
		keys = append(keys, k)
	}
	return keys
}

//	func (m model) View() string {
//		s := "📦 Node Process Output\n\n"
//		for _, line := range m.lines {
//			s += line + "\n"
//		}
//		s += "\n(press q to quit)\n"
//		return s
//	}

func (m model) menuView() string {
	if m.selected != "" {
		return fmt.Sprintf("\n✅ Selected script: %s\n", m.selected)
	}
	subHeader := colorize(36, "\nAvailable npm scripts:\n\n")
	s := subHeader

	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(">")
	for i, k := range m.keys {
		var line string
		if i == 0 {
			line = fmt.Sprintf("  %s %s \n", k, m.simbol[k])
		} else {
			line = fmt.Sprintf("  %s %s \n", k, m.simbol[k])
		}
		if i == m.cursor {
			selectLine := selectStyle.Render(line)
			s += fmt.Sprintf("%s %s\n", cursorStyle, selectLine) // один пробел после >
		} else {
			s += fmt.Sprintf("  %s\n", line) // можно оставить 2 пробела или 1
		}
	}

	s += "\n↑↓ navigate | enter select | q quit\n"
	return s
}

func (m model) View() string {
	header := titleStyle.Render("📦 Node Process Output\n\n")
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

	menu := m.menuView()

	// footer := "\n(press q to quit)"
	content := fmt.Sprintf("%s%s", header, menu)

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

// --- утилита: загрузка package.json ---
func loadScriptsFromPackageJSON() map[string]string {
	data, err := os.ReadFile("package.json")
	if err != nil {
		return map[string]string{"error": "❌ package.json not found"}
	}

	var pkg map[string]interface{}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return map[string]string{"error": "❌ invalid package.json"}
	}

	scripts := map[string]string{}
	if s, ok := pkg["scripts"].(map[string]interface{}); ok {
		for key, value := range s {
			scripts[key] = fmt.Sprintf("%v", value)
		}
	} else {
		scripts["info"] = "❌ no scripts found"
	}
	return scripts
}

func loadSimbolFromJSON() map[string]string {
	data, err := os.ReadFile("simbol.json")
	if err != nil {
		return map[string]string{"error": "❌ simbol.json not found"}
	}

	var jsn map[string]string
	if err := json.Unmarshal(data, &jsn); err != nil {
		return map[string]string{"error": "❌ invalid package.json"}
	}

	return jsn
}

// normal !!!
func main() {
	// запускаем node до создания программы, получаем канал
	// ch, err := startNodeProcess()
	// if err != nil {
	// 	fmt.Println("Failed to start node:", err)
	// 	return
	// }

	// создаём программу с моделью, у которой уже есть канал
	// p := tea.NewProgram(initialModel(ch))
	p := tea.NewProgram(initialModel())

	// запускаем (Run блокирует)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
	}
}
