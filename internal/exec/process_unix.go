//go:build linux || darwin

package exec

import (
	"os"
	osexec "os/exec"
	"syscall"
)

func configureProcess(command *osexec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func interruptProcess(command *osexec.Cmd) {
	if command.Process != nil {
		_ = syscall.Kill(-command.Process.Pid, syscall.SIGINT)
	}
}

func killProcess(command *osexec.Cmd) {
	if command.Process != nil {
		_ = syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		_ = command.Process.Signal(os.Kill)
	}
}
