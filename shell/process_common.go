package shell

import (
	"os"
	"time"
)

// terminationStep represents one step of a graceful termination sequence
type terminationStep struct {
	name   string
	signal os.Signal
	wait   time.Duration
}

// OS別に以下の関数を提供（process_unix.go / process_windows.go に実装あり）:
//   setShellProcessGroup(cmd *exec.Cmd)
//   registerShellSignals(ch chan<- os.Signal)
//   defaultShellTerminationSteps() []terminationStep
//   sendShellSignalToGroup(pid int, sig os.Signal) error
//   isShellKillSignal(sig os.Signal) bool
