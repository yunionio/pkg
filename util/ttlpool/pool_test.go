package ttlpool

import (
	"reflect"
	"testing"
	"time"

	"yunion.io/x/log"
)

type testObj struct {
	index string
}

func (o *testObj) Index() (string, error) {
	return o.index, nil
}

func (o *testObj) String() string {
	return o.index
}

func TestBasicOperate(t *testing.T) {
	p := NewTTLPool(1 * time.Second)
	o1 := &testObj{"1"}
	p.Add(o1)
	no1, exists, err := p.GetByKey("1")
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Errorf("%v should exists in pool", o1)
	}
	if !reflect.DeepEqual(no1, o1) {
		t.Errorf("%#v in cache not deep equal orgin %#v", o1, no1)
	}

	time.Sleep(1 * time.Second)
	_, exists, err = p.Get(&testObj{"1"})
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Errorf("ttl not working!")
	}

	p.Add(no1)
	in, err := p.HasByKey("1")
	if err != nil {
		t.Fatal(err)
	}
	if !in {
		t.Errorf("item should in pool")
	}
	p.Add(no1)
	p.Add(&testObj{"2"})
	p.Add(&testObj{"2"})
	p.Add(no1)
	p.Add(no1)

	if len(p.List()) != 2 {
		t.Errorf("item count > 2 in pool")
	}
	log.Debugf("now in pool list: %#v", p.List())

	p.Delete(no1)
	err = p.Delete(no1)
	err = p.DeleteByKey("1")
	if err != nil {
		t.Fatalf("delete twice err: %v", err)
	}
	if len(p.List()) != 1 {
		t.Errorf("item count != 1 in pool after delete one")
	}
}

func TestCountPool_Add(t *testing.T) {
	pool := NewCountPool()
	err := pool.Add(&testObj{"1"}, 2)
	if err != nil {
		t.Fatal(err)
	}

	obj, exists, err := pool.Get(&testObj{"1"})
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Errorf("obj 1 must exist in pool")
	}
	count, exists, err := pool.GetCount(&testObj{"1"})
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Errorf("obj 1 must exists in pool")
	}
	if count != 2 {
		t.Errorf("obj count must == 2")
	}
	log.Debugf("%#v, countMap: %#v", obj, pool.countMap)
}

func TestCountPool_Delete(t *testing.T) {
	pool := NewCountPool()
	err := pool.Add(&testObj{"1"}, 2)
	if err != nil {
		t.Fatal(err)
	}
	pool.Delete(&testObj{"1"})
	count, _, _ := pool.GetCount(&testObj{"1"})
	obj, exists, _ := pool.GetByKey("1")
	if count != 1 || obj == nil || !exists {
		t.Errorf("Delete should = 1")
	}

	pool.Delete(&testObj{"1"})
	count, ok := pool.GetCountByKey("1")
	if ok {
		t.Errorf("already delete item should not be found.")
	}

	exists, _ = pool.HasByKey("1")
	if count != 0 || exists {
		t.Errorf("Delete item should not exists")
	}
}
