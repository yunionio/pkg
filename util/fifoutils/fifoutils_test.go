package fifoutils

import "testing"

func TestFIFO_Pop(t *testing.T) {
	fifo := NewFIFO()
	fifo.Push(1)
	fifo.Push(2)
	fifo.Push(3)
	fifo.Push(4)
	t.Logf("%d: %s", fifo.len, fifo.array)
	val, err := fifo.Pop()
	t.Logf("%s %s", val, err)
	t.Logf("%d: %s", fifo.len, fifo.array)
	val, err = fifo.Pop()
	t.Logf("%s %s", val, err)
	t.Logf("%d: %s", fifo.len, fifo.array)
	val, err = fifo.Pop()
	t.Logf("%s %s", val, err)
	t.Logf("%d: %s", fifo.len, fifo.array)
	val, err = fifo.Pop()
	t.Logf("%s %s", val, err)
	t.Logf("%d: %s", fifo.len, fifo.array)
	val, err = fifo.Pop()
	t.Logf("%s %s", val, err)
	t.Logf("%d: %s", fifo.len, fifo.array)
	val, err = fifo.Pop()
	t.Logf("%s %s", val, err)
	t.Logf("%d: %s", fifo.len, fifo.array)
}
