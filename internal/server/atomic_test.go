package server

import "testing"

func TestAtomicInt(t *testing.T) {
	var x aint
	x.set(10)
	if x.get() != 10 {
		t.Fatalf("expected %v, got %v", 10, x.get())
	}
	x.add(-9)
	if x.get() != 1 {
		t.Fatalf("expected %v, got %v", 1, x.get())
	}
	x.add(-1)
	if x.get() != 0 {
		t.Fatalf("expected %v, got %v", 0, x.get())
	}
}
