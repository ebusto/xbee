package xbee

import (
	"testing"
)

func TestSequence(t *testing.T) {
	s := NewSequence()
	l := byte(0)

	for i := 0; i < 1024; i++ {
		v := <-s   // Current value.
		e := l + 1 // Expected value.

		// Zero is skipped.
		if e == 0 {
			e = 1
		}

		if i == 0 && v != 1 {
			t.Fatalf("Initial value is not 1: %d", v)
		}

		if i > 0 && v != e {
			t.Fatal("Next value not last + 1: %d", e)
		}

		l = v
	}
}
