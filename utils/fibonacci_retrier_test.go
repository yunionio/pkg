package utils

import (
	"testing"
	"time"
)

func TestFibonacciRetrier(t *testing.T) {
	t.Run("maxTries=3", func(t *testing.T) {
		startTime := time.Now()
		fibr := NewFibonacciRetrierMaxTries(3, func(FibonacciRetrier) (bool, error) {
			return false, nil
		})
		done, err := fibr.Start()
		if done {
			t.Errorf("should never done")
		}
		ok, _ := matchFibonacciRetrierErrorType(err, FibonacciRetrierErrorMaxTriesExceeded)
		if !ok {
			t.Errorf("should be fibonacci err max tries exceeded: got %v", err)
		}
		elapsed := time.Since(startTime)
		t.Logf("wall time elapsed %s", elapsed)
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
		done, err := fibr.Start()
		if done {
			t.Errorf("should never done")
		}
		ok, _ := matchFibonacciRetrierErrorType(err, FibonacciRetrierErrorMaxElapseExceeded)
		if !ok {
			t.Errorf("should be fibonacci err max elapse exceeded: got %v", err)
		}
		elapsed := time.Since(startTime)
		if elapsed-6*time.Second > time.Millisecond {
			t.Errorf("wall time elapsed %s, expecting around %s", 3*time.Second, 6*time.Second)
		}
		t.Logf("wall time elapsed %s", elapsed)
		gotElapsed1 := fibr.Elapsed()
		if gotElapsed1-elapsed > time.Millisecond {
			t.Errorf("wall time elapsed %s, got %s", elapsed, gotElapsed1)
		}

		time.Sleep(time.Second)
		gotElapsed2 := fibr.Elapsed()
		if gotElapsed1 != gotElapsed2 {
			t.Errorf("two calls to Elapsed() should return equal value %s != %s", gotElapsed1, gotElapsed2)
		}
	})
}
