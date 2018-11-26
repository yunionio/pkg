package signalutils

import (
	"os"
	"syscall"
	"testing"
)

func TestStartTrap(t *testing.T) {
	exitChan := make(chan struct{}, 1)
	testSIGHUP := func() {
		println("Called test SIGHUP!!!")
		close(exitChan)
	}
	testSIGINT := func() {
		println("Called test SIGINT!!!")
	}

	RegisterSignal(testSIGHUP, syscall.SIGHUP)
	RegisterSignal(testSIGINT, syscall.SIGINT)
	t.Run("Signal test", func(t *testing.T) {
		StartTrap()
		syscall.Kill(os.Getpid(), syscall.SIGINT)
		syscall.Kill(os.Getpid(), syscall.SIGHUP)
		<-exitChan
	})
}
