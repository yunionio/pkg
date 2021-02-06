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

package utils

import (
	"context"
	"testing"
	"time"
)

func TestFibonacciRetrier(t *testing.T) {
	jitTol := 5 * time.Millisecond
	t.Run("maxTries=3", func(t *testing.T) {
		startTime := time.Now()
		fibr := NewFibonacciRetrierMaxTries(3, func(FibonacciRetrier) (bool, error) {
			return false, nil
		})
		done, err := fibr.Start(context.Background())
		if done {
			t.Errorf("should never done")
		}
		ok, _ := matchFibonacciRetrierErrorType(err, FibonacciRetrierErrorMaxTriesExceeded)
		if !ok {
			t.Errorf("should be fibonacci err max tries exceeded: got %v", err)
		}
		elapsed := time.Since(startTime)
		if fibr.Elapsed()-elapsed > time.Millisecond {
			t.Errorf("wall time elapsed %s, got %s", elapsed, fibr.Elapsed())
		}
		if fibr.Elapsed()-3*time.Second > time.Millisecond {
			t.Errorf("should wait no more than %s, got %s", 3*time.Second, fibr.Elapsed())
		}
	})

	t.Run("maxElapse=5s", func(t *testing.T) {
		startTime := time.Now()
		fibr := NewFibonacciRetrierMaxElapse(5*time.Second, func(FibonacciRetrier) (bool, error) {
			return false, nil
		})
		done, err := fibr.Start(context.Background())
		if done {
			t.Errorf("should never done")
		}
		ok, _ := matchFibonacciRetrierErrorType(err, FibonacciRetrierErrorMaxElapseExceeded)
		if !ok {
			t.Errorf("should be fibonacci err max elapse exceeded: got %v", err)
		}
		elapsed := time.Since(startTime)
		if elapsed-6*time.Second > jitTol {
			t.Errorf("wall time elapsed %s, expecting around %s", elapsed, 6*time.Second)
		}
		gotElapsed1 := fibr.Elapsed()
		if gotElapsed1-elapsed > jitTol {
			t.Errorf("wall time elapsed %s, got %s", elapsed, gotElapsed1)
		}

		time.Sleep(time.Second)
		gotElapsed2 := fibr.Elapsed()
		if gotElapsed1 != gotElapsed2 {
			t.Errorf("two calls to Elapsed() should return equal value %s != %s", gotElapsed1, gotElapsed2)
		}
	})

	t.Run("ctx-done", func(t *testing.T) {
		fibr := NewFibonacciRetrierMaxElapse(time.Hour, func(FibonacciRetrier) (bool, error) {
			return false, nil
		})
		ctx := context.Background()
		ctx, _ = context.WithTimeout(ctx, time.Duration(0))

		doneC := make(chan int)
		go func() {
			wait := time.NewTimer(time.Second)
			select {
			case <-wait.C:
				t.Fatalf("1s passed but not done yet")
			case <-doneC:
			}
		}()
		fibr.Start(ctx)
		close(doneC)
		if fibr.Elapsed() > time.Millisecond {
			t.Errorf("should return immediately")
		}
	})
}
