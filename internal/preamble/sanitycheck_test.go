package preamble

import "testing"

// TestSanityCheck runs the same startup check the ai agent runs in main(), so a
// new Operation or builtin without a description (or reserved-word entry) is
// caught by test-local instead of panicking the agent at launch.
func TestSanityCheck(t *testing.T) {
	SanityCheck()
}
