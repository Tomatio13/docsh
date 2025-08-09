//go:build windows

package executor

import (
	"os"
	"os/exec"
	osSignal "os/signal"
	"time"
)

// setNewProcessGroup is a no-op on Windows (different semantics)
func setNewProcessGroup(cmd *exec.Cmd) {
	// No special process group handling required here
}

// terminateProcess terminates the process on Windows
func terminateProcess(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	// Try graceful signal via Process.Signal if supported; otherwise Kill
	// On Windows, os.Process.Kill performs TerminateProcess
	_ = cmd.Process.Kill()
	time.Sleep(100 * time.Millisecond)
}

// registerTerminationSignals registers signals to notify (limited on Windows)
func registerTerminationSignals(ch chan<- os.Signal) {
	// os.Interrupt は利用可能
	osSignal.Notify(ch, os.Interrupt)
}
