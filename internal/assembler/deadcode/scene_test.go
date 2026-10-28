package deadcode_test

import (
	"reflect"
	"testing"

	"gradientcontrol.com/omega/core/internal/assembler/deadcode"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/program"
)

func TestRemovePreservesSceneAndSamplingBounds(t *testing.T) {
	prog, err := assembleUnpruned([]string{"unused = 1; return 0;"}, &source.MockManager{})
	if err != nil {
		t.Fatal(err)
	}
	prog.Primary.Scene = &program.Scene{
		Size: 1105, Grid: 20,
		SpinodoidBounds: []float64{-51, -1, -6, 58, 264, 40},
	}
	pruned, err := deadcode.Remove(prog)
	if err != nil {
		t.Fatal(err)
	}
	got := pruned.Primary.Scene
	if got == nil || got.Size != 1105 || got.Grid != 20 || !reflect.DeepEqual(got.SpinodoidBounds, prog.Primary.Scene.SpinodoidBounds) {
		t.Fatalf("scene metadata lost during dead code removal: %v", got)
	}
	got.SpinodoidBounds[0] = -99
	if prog.Primary.Scene.SpinodoidBounds[0] != -51 {
		t.Fatal("pruned scene aliases the original sampling bounds")
	}
}
