package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"

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
	lines  []string
	ch     chan tea.Msg
	screen string
	menu
	// ДОБАВЛЕНО: поле для хранения запущенного процесса
	runningCmd *exec.Cmd
}

// Msg types
type nodeLineMsg string
type nodeDoneMsg struct{}
type nodeErrMsg string

func colorize(colorNumber int, text string) string {
	return fmt.Sprintf("\033[1m\033[%dm%s\033[0m", colorNumber, text)
}

var (
	// --- Styles ---
	// Стили на основе Tokyo Nigh
	// Основные поверхности
	Background = "#1a1b26" // Основной фон
	Surface0   = "#1f2335" // Чуть светлее фона
	Surface1   = "#24283b" // Панели, карточки
	Surface2   = "#292e42" // Ховер, подсветка
	Overlay0   = "#3b4261" // Второстепенные элементы
	Overlay1   = "#565f89" // Комментарии, неактивный текст

	// Текст
	Foreground = "#c0caf5" // Основной текст
	Subtext0   = "#a9b1d6" // Второстепенный текст
	Subtext1   = "#737aa2" // Ещё менее важный

	// Акценты
	Blue  = "#7aa2f7" // Ссылки, активные элементы, start
	Blue0 = "#3d59a1" // Тёмный синий
	Blue1 = "#2ac3de" // Яркий голубой (альтернатива)

	Green = "#9ece6a" // Успех, build
	Teal  = "#73daca" // Альтернатива зелёному

	Yellow = "#e0af68" // Предупреждения, lint
	Orange = "#ff9e64" // Альтернатива жёлтому

	Red = "#f7768e" // Ошибки, test (если упал)

	Magenta  = "#bb9af7" // Ключевые слова, npm, нумерация [1]
	Magenta2 = "#ff007c" // Яркий розовый (опционально)

	Cyan = "#7dcfff" // Строки, info, логи

	// Дополнительно
	GitAdd    = "#449dab"
	GitDelete = "#f7768e"
	GitChange = "#e0af68"

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(Blue)). // Рамка — яркий синий акцент
		// Background(lipgloss.Color(Background)).
		Padding(1, 2).
		Margin(1).
		Width(50).
		Height(15)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Blue)). // Яркий синий заголовок
			Bold(true)

	subTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Subtext0)). // Мягкий серо-голубой подзаголовок
			Italic(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Green)) // Успех — насыщенный зелёный

	selectStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Cyan)).     // Выделение — яркий голубой
			Background(lipgloss.Color(Surface2)). // Фон выделенного пункта
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Red)) // Ошибки — красный

	itemNumberStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Magenta)) // Нумерация [1], [2] — фиолетовый

	itemTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Foreground))
	runningScriptTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(Magenta))

	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Subtext1))

	// константы
	hint = "↑↓(jk) navigate | enter(l) select | q(h) quit"
)

// initial model (channel assigned later)
func initialModel(ch chan tea.Msg) model {
	scripts := loadScriptsFromPackageJSON()
	simbol := loadSimbolFromJSON()
	keys := make([]string, 0, len(scripts))
	for k := range scripts {
		keys = append(keys, k)
	}

	// ИСПРАВЛЕНО: Сортируем ключи для стабильного порядка
	sort.Strings(keys)

	return model{
		lines: []string{},
		ch:    ch,
		menu: menu{
			scripts:  scripts,
			keys:     keys, // Сохраняем отсортированный список
			simbol:   simbol,
			cursor:   0,
			selected: "",
		},
		screen:     "first",
		runningCmd: nil,
	}
}

// Init returns a Cmd that reads the first message from channel
func (m model) Init() tea.Cmd {
	// return readNextLine(m.ch)
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// ДОБАВЛЕНО: Сохраняем указатель на команду
	case *exec.Cmd:
		m.runningCmd = msg
		return m, nil // Просто сохраняем, не делаем I/O

	case nodeLineMsg:
		m.lines = append(m.lines, string(msg))
		return m, readNextLine(m.ch)
	case nodeDoneMsg:
		m.lines = append(m.lines, successStyle.Render("✅ Process finished!"))
		m.runningCmd = nil // Сбрасываем
		return m, nil
	case nodeErrMsg:
		m.lines = append(m.lines, errorStyle.Render(fmt.Sprintf("❌ Error: %s", string(msg))))
		m.runningCmd = nil // Сбрасываем
		return m, nil
	case tea.KeyMsg:
		switch m.screen {
		case "first":
			switch msg.String() {
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
				// ИСПРАВЛЕНО: Возвращаем команду "пачкой"
				return m, tea.Batch(
					runNodeScript(m.ch, m.selected), // Запускаем процесс
					readNextLine(m.ch),              // Начинаем слушать
				)
			}
		case "finished":
			switch msg.String() {
			case "q", "ctrl+c":
				// ДОБАВЛЕНО: Убиваем процесс перед выходом
				if m.runningCmd != nil && m.runningCmd.Process != nil {
					// Посылаем SIGTERM (мягкое завершение)
					if err := m.runningCmd.Process.Signal(os.Interrupt); err != nil {
						// Если не получилось, убиваем (SIGKILL)
						m.runningCmd.Process.Kill()
					}
				}
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

	case "first":
		subtitle := subTitleStyle.Render("Available npm scripts:")

		items := []string{}
		for i, script := range m.keys {
			prefix := itemNumberStyle.Render(fmt.Sprintf("[ %d ]", i+1))
			name := itemTextStyle.Render(script)
			icon := successStyle.Render(m.simbol[script])
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

// runNodeScript
func runNodeScript(ch chan tea.Msg, name string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("npm", "run", name)
		// УДАЛЕНО: m.runningCmd = cmd - изменена логика, чтобы избежать модификации состояния вне Update.

		// ИСПРАВЛЕНИЕ №1: Устанавливаем CI=true
		cmd.Env = append(os.Environ(), "CI=true")

		// ИСПРАВЛЕНИЕ №2: Явно закрываем Stdin
		stdin, err := cmd.StdinPipe()
		if err == nil {
			stdin.Close() // Говорим процессу "ввода не будет"
		}

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return nodeErrMsg("stdout pipe error: " + err.Error())
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			return nodeErrMsg("stderr pipe error: " + err.Error())
		}

		if err := cmd.Start(); err != nil {
			return nodeErrMsg("start process error: " + err.Error())
		}

		// ДОБАВЛЕНО: Отправляем саму команду в Update,
		// чтобы модель могла ее сохранить и убить при выходе
		// Это единственный безопасный способ обновить runningCmd.
		ch <- tea.Msg(cmd) 

		var wg sync.WaitGroup
		wg.Add(2)

		// Горутина для STDOUT
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				line := scanner.Text()
				ch <- nodeLineMsg(line)
			}
		}()

		// Горутина для STDERR
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				line := scanner.Text()
				fmt.Println("DEBUG (stderr):", line)
				ch <- nodeLineMsg(line) // Отправляем в тот же канал
			}
		}()

		// Горутина-наблюдатель
		go func() {
			cmd.Wait() // Ждем завершения `npm run ...`
			wg.Wait()  // Ждем, пока сканеры закончат читать
			close(ch)  // *Только* после этого закрываем канал
			fmt.Println("DEBUG: All processes and scanners finished. Channel closed.")
		}()

		return nil
	}
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
		fmt.Println("Warning: simbol.json not found, using empty map.")
		return map[string]string{} // Лучше вернуть пустую мапу
	}

	var jsn map[string]string
	if err := json.Unmarshal(data, &jsn); err != nil {
		fmt.Println("Warning: invalid simbol.json, using empty map.")
		return map[string]string{} // И здесь тоже
	}

	return jsn
}

// normal !!!
func main() {
	ch := make(chan tea.Msg)

	p := tea.NewProgram(initialModel(ch))

	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
