package tape_test

import (
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/assembler"
	"github.com/Gradient-Control-Laboratories/omega-core/backend/tape"
	"google.golang.org/protobuf/proto"
)

// A shape with a colour and a material reaches the tape with the colour's three
// trees and the material's sixteen (four rows of four, the lay last). The
// converter writes sixteen; the serializer once took only twelve and dropped
// every authored material, so no tape carried one.
func TestCompileWithAppearanceCarriesTheWholeMaterial(t *testing.T) {
	src := `
ball = sphere(3);
ball.material = material.matte;
ball.color = color.cherry;

bare = transform.translate(sphere(2), [10, 0, 0]);

return ball, bare;
`
	res, err := assembler.Assemble([]string{src})
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}
	tapeBytes, err := tape.Compile(res.Program, tape.WithAppearance(true))
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	var tp tape.Tape
	if err := proto.Unmarshal(tapeBytes, &tp); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(tp.Appearance) != 2 {
		t.Fatalf("%d appearances, want one per return (2)", len(tp.Appearance))
	}

	ball := tp.Appearance[0]
	if len(ball.Color) != 3 {
		t.Errorf("the ball's colour has %d channels, want 3", len(ball.Color))
	}
	if len(ball.Material) != 16 {
		t.Fatalf("the ball's material has %d entries, want 16", len(ball.Material))
	}
	for i, id := range ball.Material {
		if int(id) >= len(tp.Nodes) {
			t.Errorf("material entry %d names node %d of %d", i, id, len(tp.Nodes))
		}
	}

	// A return that authors no material has none, or the default record whole:
	// never a partial one.
	if n := len(tp.Appearance[1].Material); n != 0 && n != 16 {
		t.Errorf("the bare shape's material has %d entries, want 0 or 16", n)
	}
}
