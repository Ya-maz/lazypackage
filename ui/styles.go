package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// func colorize(colorNumber int, text string) string {
// 	return fmt.Sprintf("\033[1m\033[%dm%s\033[0m", colorNumber, text)
// }

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
		Margin(1)

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


