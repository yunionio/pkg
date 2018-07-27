package clock

import (
	"testing"
	"time"
)

func TestFakeClock(t *testing.T) {
	startTime := time.Now()
	tc := NewFakeClock(startTime)
	tc.Step(time.Second)
	now := tc.Now()
	if now.Sub(startTime) != time.Second {
		t.Errorf("input: %s now=%s gap=%s expected=%s", startTime, now, now.Sub(startTime), time.Second)
	}

	tt := tc.Now()
	tc.SetTime(tt.Add(time.Hour))
	if tc.Now().Sub(tt) != time.Hour {
		t.Errorf("input: %s now=%s gap=%s expected=%s", tt, tc.Now(), tc.Now().Sub(tt), time.Hour)
	}
}

func TestFakeClockSleep(t *testing.T) {
	startTime := time.Now()
	tc := NewFakeClock(startTime)
	tc.Sleep(time.Duration(1) * time.Hour)
	now := tc.Now()
	if now.Sub(startTime) != time.Hour {
		t.Errorf("Fake sleep failed, expected time to advance by one hour, instead, its %v", now.Sub(startTime))
	}
}

func TestFakeAfter(t *testing.T) {
	tc := NewFakeClock(time.Now())
	if tc.HasWaiters() {
		t.Errorf("unexpected waiter?")
	}
	oneSec := tc.After(time.Second)
	if !tc.HasWaiters() {
		t.Errorf("unexpected lack of waiter?")
	}

	oneOhOneSec := tc.After(time.Second + time.Millisecond)
	twoSec := tc.After(2 * time.Second)
	select {
	case <-oneSec:
		t.Errorf("unexpected channel read")
	case <-oneOhOneSec:
		t.Errorf("unexpected channel read")
	case <-twoSec:
		t.Errorf("unexpected channel read")
	default:
	}

	tc.Step(999 * time.Millisecond)
	select {
	case <-oneSec:
		t.Errorf("unexpected channel read")
	case <-oneOhOneSec:
		t.Errorf("unexpected channel read")
	case <-twoSec:
		t.Errorf("unexpected channel read")
	default:
	}

	tc.Step(time.Millisecond)
	select {
	case <-oneSec:
		// Expected!
	case <-oneOhOneSec:
		t.Errorf("unexpected channel read")
	case <-twoSec:
		t.Errorf("unexpected channel read")
	default:
		t.Errorf("unexpected non-channel read")
	}
	tc.Step(time.Millisecond)
	select {
	case <-oneSec:
		// should not double-trigger!
		t.Errorf("unexpected channel read")
	case <-oneOhOneSec:
		// Expected!
	case <-twoSec:
		t.Errorf("unexpected channel read")
	default:
		t.Errorf("unexpected non-channel read")
	}
}

func TestFakeTick(t *testing.T) {
	tc := NewFakeClock(time.Now())
	if tc.HasWaiters() {
		t.Errorf("unexpected waiter?")
	}
	oneSec := tc.Tick(time.Second)
	if !tc.HasWaiters() {
		t.Errorf("unexpected lack of waiter?")
	}

	oneOhOneSec := tc.Tick(time.Second + time.Millisecond)
	twoSec := tc.Tick(2 * time.Second)
	select {
	case <-oneSec:
		t.Errorf("unexpected channel read")
	case <-oneOhOneSec:
		t.Errorf("unexpected channel read")
	case <-twoSec:
		t.Errorf("unexpected channel read")
	default:
	}

	tc.Step(999 * time.Millisecond) // t=.999
	select {
	case <-oneSec:
		t.Errorf("unexpected channel read")
	case <-oneOhOneSec:
		t.Errorf("unexpected channel read")
	case <-twoSec:
		t.Errorf("unexpected channel read")
	default:
	}

	tc.Step(time.Millisecond) // t=1.000
	select {
	case <-oneSec:
		// Expected!
	case <-oneOhOneSec:
		t.Errorf("unexpected channel read")
	case <-twoSec:
		t.Errorf("unexpected channel read")
	default:
		t.Errorf("unexpected non-channel read")
	}
	tc.Step(time.Millisecond) // t=1.001
	select {
	case <-oneSec:
		// should not double-trigger!
		t.Errorf("unexpected channel read")
	case <-oneOhOneSec:
		// Expected!
	case <-twoSec:
		t.Errorf("unexpected channel read")
	default:
		t.Errorf("unexpected non-channel read")
	}

	tc.Step(time.Second) // t=2.001
	tc.Step(time.Second) // t=3.001
	tc.Step(time.Second) // t=4.001
	tc.Step(time.Second) // t=5.001

	// The one second ticker should not accumulate ticks
	accumulatedTicks := 0
	drained := false
	for !drained {
		select {
		case <-oneSec:
			accumulatedTicks++
		default:
			drained = true
		}
	}
	if accumulatedTicks != 1 {
		t.Errorf("unexpected number of accumulated ticks: %d", accumulatedTicks)
	}
}
