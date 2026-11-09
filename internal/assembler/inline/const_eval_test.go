package inline_test

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// A select's condition is tried for a compile-time value by following its
// references into their definitions. A property expansion writes its shared
// nodes as a chain of locals, each naming earlier ones more than once, and
// walked as a tree such a chain costs 2^n: .second of Cracked Mud never
// finished assembling. Forty links would take hours unmemoized.
func TestInline_ConstConditionThroughSharedChain(t *testing.T) {
	var b strings.Builder
	b.WriteString("a0 = p.x;\n")
	for i := 1; i <= 40; i++ {
		fmt.Fprintf(&b, "a%d = a%d + a%d;\n", i, i-1, i-1)
	}
	b.WriteString("return a40 < 1 ? length(p) - 1 : length(p) - 2;\n")

	done := make(chan error, 1)
	go func() {
		_, err := runPipeline(b.String())
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("inlining a select over a chain of shared locals did not finish in 30 s")
	}
}
