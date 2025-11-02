package node

import (
	"bufio"
	"cli/msg"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// глобальная переменная для позиции — сбрасывается при старте
var lastOffset int64 = 0

func RunNodeScript(ch chan tea.Msg, name string) tea.Cmd {
	return func() tea.Msg {
		logPath := "temp.log"

		lastOffset = 0
		_ = os.Remove(logPath)

		cmd := exec.Command("npm", "run", name)

		logFile, err := os.Create(logPath)
		if err != nil {
			return msg.NodeErr("cannot create log file: " + err.Error())
		}

		cmd.Stdout = logFile
		cmd.Stderr = logFile

		if err := cmd.Start(); err != nil {
			logFile.Close()
			return msg.NodeErr("failed to start process: " + err.Error())
		}

		// горутина: ждём завершения процесса и уведомим модель
		go func() {
			cmd.Wait()
			// даём ОС записать остаток, затем закрываем файл
			logFile.Sync()
			logFile.Close()
			// небольшой буфер, чтобы все данные успели попасть в файл
			time.Sleep(200 * time.Millisecond)
			ch <- msg.NodeDone{}
		}()

		// горутина: tail-подобное чтение лога
		go func() {
			for {
				// если процесса уже нет и файл не растёт — завершить чтение цикла
				f, err := os.Open(logPath)
				if err != nil {
					time.Sleep(200 * time.Millisecond)
					continue
				}

				fi, err := f.Stat()
				if err != nil {
					f.Close()
					time.Sleep(200 * time.Millisecond)
					continue
				}

				// ничего нового
				if fi.Size() <= lastOffset {
					f.Close()
					// если процесс завершился и файл не растёт — выйти из цикла
					// но чтобы узнать это, лучше смотреть на nodeDoneMsg в Update (не здесь)
					time.Sleep(200 * time.Millisecond)
					continue
				}

				// переместимся в позицию lastOffset
				_, _ = f.Seek(lastOffset, io.SeekStart)
				reader := bufio.NewReader(f)

				for {
					line, err := reader.ReadString('\n')
					if line != "" {
						trimmed := strings.TrimRight(line, "\r\n")
						ch <- msg.NodeLine(trimmed)
						// обновляем offset на длину считанной строки
						lastOffset += int64(len(line))
					}
					if err != nil {
						if err == io.EOF {
							// прочитали всё, закрываем файл и продолжим внешний цикл
							break
						} else {
							// непредвиденная ошибка — уведомим модель и остановим читатель
							ch <- msg.NodeErr("read error: " + err.Error())
							f.Close()
							return
						}
					}
				}

				f.Close()
				// небольшой sleep, чтобы не дергать диск слишком часто
				time.Sleep(200 * time.Millisecond)
			}
		}()

		return nil
	}
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
