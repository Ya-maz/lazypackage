package node

import (
	"bufio"
	"cli/msg"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/creack/pty"
	tea "github.com/charmbracelet/bubbletea"
)

func RunNodeScript(ch chan tea.Msg, inputCh chan string, name string) tea.Cmd {
	return func() tea.Msg {
		// Мы больше не полагаемся на temp.log для UI, 
		// но можем продолжать писать туда для истории если нужно.
		logPath := "temp.log"
		_ = os.Remove(logPath)
		logFile, _ := os.Create(logPath)
		defer logFile.Close()

		c := exec.Command("npm", "run", name)
		
		// Запуск через PTY
		ptmx, err := pty.Start(c)
		if err != nil {
			return msg.NodeErr("failed to start pty: " + err.Error())
		}
		defer func() { _ = ptmx.Close() }()

		// Копируем вывод PTY в лог-файл (опционально)
		go func() { 
			_, _ = io.Copy(logFile, ptmx) 
		}()

		// Канал для сигнала о завершении
		processDone := make(chan bool)

		go func() {
			_ = c.Wait()
			processDone <- true
		}()

		// Горутина для записи ВВОДА из TUI в процесс
		go func() {
			for {
				data, ok := <-inputCh
				if !ok {
					return
				}
				_, _ = ptmx.Write([]byte(data))
			}
		}()

		// Читаем из PTY и шлем в модель ЧАНКАМИ (не строками)
		go func() {
			buf := make([]byte, 1024)
			for {
				n, err := ptmx.Read(buf)
				if n > 0 {
					// Отправляем как сырые данные
					ch <- msg.NodeData(string(buf[:n]))
				}
				if err != nil {
					break
				}
			}
		}()

		// Ждем завершения
		go func() {
			<-processDone
			// Небольшая задержка чтобы вычитать последние байты
			time.Sleep(100 * time.Millisecond)
			close(ch)
		}()

		return nil
	}
}

// Теперь эта функция не нужна для основного потока, так как мы читаем напрямую из PTY
// Но оставим для обратной совместимости или удалим позже
func readRemaining(logPath string, ch chan<- tea.Msg) {
	// ... (старый код)
}

func ReadNextLine(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		if ch == nil {
			return msg.NodeErr("internal: channel is nil")
		}
		m, ok := <-ch
		if !ok {
			return msg.NodeDone{}
		}
		return m
	}
}
