// Copyright 2019 Yunion
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
