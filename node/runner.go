package node

import (
	"os/exec"
)

func GetNodeCommand(name string) *exec.Cmd {
	c := exec.Command("npm", "run", name)
	// Для tea.ExecProcess (и вообще интерактивных задач) важно, 
	// чтобы Stdin/Stdout/Stderr были проксированы.
	// BubbleTea сам все настроит, нам нужно лишь дать ему Cmd.
	// Но если мы хотим логи, то нужно stdout завернуть в MultiWriter.
	
	return c
}
