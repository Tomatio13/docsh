//go:build !windows

package executor

import (
	"os"
	"os/exec"
	osSignal "os/signal"
	"syscall"
	"time"
)

// setNewProcessGroup sets a new process group for the command (Unix-like only)
func setNewProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// terminateProcess sends signals to the process group, then force kills if needed
func terminateProcess(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	pid := cmd.Process.Pid
	if pid > 0 {
		// Try graceful terminate
		_ = syscall.Kill(-pid, syscall.SIGTERM)
		time.Sleep(100 * time.Millisecond)
		// Force kill if still running
		if cmd.ProcessState == nil || !cmd.ProcessState.Exited() {
			_ = syscall.Kill(-pid, syscall.SIGKILL)
			_ = cmd.Process.Kill()
		}
	}
}

// registerTerminationSignals registers signals to notify for graceful termination
func registerTerminationSignals(ch chan<- os.Signal) {
	osSignal.Notify(ch, os.Interrupt, syscall.SIGTERM)
}
