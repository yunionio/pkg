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
