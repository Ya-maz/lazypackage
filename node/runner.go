package node

import (
	"cli/msg"
	"os"
	"os/exec"
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
		// НЕ закрываем ptmx здесь через defer, иначе он закроется сразу же после выхода из функции,
		// а горутины еще работают! Закроем его позже, когда процесс завершится.

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

		// Читаем из PTY, пишем в лог и шлем в модель ЧАНКАМИ
		go func() {
			buf := make([]byte, 1024)
			for {
				n, err := ptmx.Read(buf)
				if n > 0 {
					chunk := buf[:n]
					// Пишем в лог-файл (опционально, для дебага)
					_, _ = logFile.Write(chunk)
					// Отправляем в UI как сырые данные
					ch <- msg.NodeData(string(chunk))
				}
				if err != nil {
					// EOF или ошибка PTY (например, закрылся master)
					break
				}
			}
		}()

		// Ждем завершения
		go func() {
			<-processDone
			// Небольшая задержка чтобы вычитать последние байты, если ридер еще не отвалился
			time.Sleep(100 * time.Millisecond)
			_ = ptmx.Close() // Вот теперь можно закрывать
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
