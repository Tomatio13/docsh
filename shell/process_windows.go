//go:build windows

package shell

import (
	"os"
	"os/exec"
	osSignal "os/signal"
	"time"
)

func setShellProcessGroup(cmd *exec.Cmd) {
	// No special process group handling on Windows
}

func registerShellSignals(ch chan<- os.Signal) {
	// Windows supports os.Interrupt
	osSignal.Notify(ch, os.Interrupt)
}

func defaultShellTerminationSteps() []terminationStep {
	// On Windows, fall back to process.Kill
	return []terminationStep{{"TerminateProcess (Kill)", os.Interrupt, 100 * time.Millisecond}}
}

func sendShellSignalToGroup(pid int, sig os.Signal) error {
	// No PGID signals on Windows, caller will try Process.Signal/Kill
	return nil
}

func isShellKillSignal(sig os.Signal) bool { return true }
