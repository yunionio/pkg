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
