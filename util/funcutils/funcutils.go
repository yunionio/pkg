package funcutils

import (
	"time"

	"github.com/yunionio/log"
)

func RetryUntilSuccess(callee func(), interval time.Duration, onSucc func()) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("call function failed, retry ...")
			time.Sleep(interval)
			RetryUntilSuccess(callee, interval, onSucc)
		} else {
			onSucc()
		}
	}()
	callee()
}
