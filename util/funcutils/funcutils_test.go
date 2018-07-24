package funcutils

import (
	"fmt"
	"testing"
	"time"
)

var tried int = 0

const retryMax = 5

func failCall() {
	if tried < retryMax {
		tried += 1
		panic(fmt.Sprintf("failed until %d", retryMax))
	}
	fmt.Sprintf("Success!")
}

func TestRetryUntilSuccess(t *testing.T) {
	RetryUntilSuccess(failCall, time.Second, func() {
		t.Logf("Success!! tried %d", tried)
	})
}
