package state

import (
	"sync"
	"testing"
)

func TestSetHasClear(t *testing.T) {
	s := NewFlagSet()
	if s.Has(7) {
		t.Fatal("empty set reports a flag")
	}
	s.Set(7)
	if !s.Has(7) {
		t.Fatal("flag not set")
	}
	s.Clear(7)
	if s.Has(7) {
		t.Fatal("flag not cleared")
	}
}

func TestConcurrentAccess(t *testing.T) {
	s := NewFlagSet()
	var wg sync.WaitGroup
	for i := int64(0); i < 200; i++ {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			s.Set(id)
			_ = s.Has(id)
			s.Clear(id)
		}(i)
	}
	wg.Wait()
}
