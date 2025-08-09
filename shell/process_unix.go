//go:build !windows

package shell

import (
	"os"
	"os/exec"
	osSignal "os/signal"
	"syscall"
	"time"
)

func setShellProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func registerShellSignals(ch chan<- os.Signal) {
	osSignal.Notify(ch, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
}

func defaultShellTerminationSteps() []terminationStep {
	return []terminationStep{
		{"SIGTERM to process group", syscall.SIGTERM, 200 * time.Millisecond},
		{"SIGKILL to process group", syscall.SIGKILL, 100 * time.Millisecond},
	}
}

func sendShellSignalToGroup(pid int, sig os.Signal) error {
	if sig, ok := sig.(syscall.Signal); ok {
		return syscall.Kill(-pid, sig)
	}
	return nil
}

func isShellKillSignal(sig os.Signal) bool {
	if s, ok := sig.(syscall.Signal); ok {
		return s == syscall.SIGKILL
	}
	return false
}
