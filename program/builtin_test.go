package program

import "testing"

func TestBuiltins(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("panic: %v", r)
		}
	}()
	SanityCheck()
}
