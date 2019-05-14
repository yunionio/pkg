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
	"sync"
	"testing"
	"time"
)

func TestWaitGroupWrapper_Wrap(t *testing.T) {
	type fields struct {
		WaitGroup sync.WaitGroup
	}
	type args struct {
		cb func()
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test1",
			fields: fields{
				WaitGroup: sync.WaitGroup{},
			},
			args: args{
				func() {
					time.Sleep(3 * time.Second)
				},
			},
		},
		{
			name: "test2",
			fields: fields{
				WaitGroup: sync.WaitGroup{},
			},
			args: args{
				func() {
					time.Sleep(0 * time.Second)
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WaitGroupWrapper{
				WaitGroup: tt.fields.WaitGroup,
			}
			w.Wrap(tt.args.cb)
		})
	}
}

func TestWaitTimeOut(t *testing.T) {
	type args struct {
		wg      *WaitGroupWrapper
		timeout time.Duration
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test1",
			args: args{
				wg:      new(WaitGroupWrapper),
				timeout: time.Duration(time.Second * 5),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WaitTimeOut(tt.args.wg, tt.args.timeout); got != tt.want {
				t.Errorf("WaitTimeOut() = %v, want %v", got, tt.want)
			}
		})
	}
}
