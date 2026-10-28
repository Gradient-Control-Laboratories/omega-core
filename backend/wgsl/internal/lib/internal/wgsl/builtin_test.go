package wgsl

import (
	"testing"
)

func TestSanityCheck(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SanityCheck() panicked: %v", r)
		}
	}()
	SanityCheck()
}
