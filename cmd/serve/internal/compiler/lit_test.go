package compiler_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/cmd/serve/internal/compiler"
)

// litCamera looks along +Y at the origin from 50 units away, with the lens the tests' frames share.
func litCamera() *compiler.LitCamera {
	return &compiler.LitCamera{
		Pos:    [3]float64{0, -50, 0},
		Right:  [3]float64{1, 0, 0},
		Up:     [3]float64{0, 0, 1},
		Fwd:    [3]float64{0, 1, 0},
		TanFov: [2]float64{0.4, 0.3},
	}
}

func TestRenderLit_DrawsTheSphereAtTheCentre(t *testing.T) {
	lin, err := compiler.RenderLit(compiler.LitRequest{
		Source: "return sphere(5);", Width: 32, Height: 24, WorldExtent: 40, Camera: litCamera(),
	}, []string{"."}, nil)
	if err != nil {
		t.Fatalf("RenderLit: %v", err)
	}
	if lin.Width != 32 || lin.Height != 24 {
		t.Fatalf("frame is %dx%d, want 32x24", lin.Width, lin.Height)
	}
	centre := 12*32 + 16
	if lin.Shape[centre] != 1 {
		t.Errorf("the centre pixel hit return %d, want 1", lin.Shape[centre])
	}
	if d := lin.Depth[centre]; d < 44 || d > 46 {
		t.Errorf("the centre pixel's depth is %g, want about 45 (the near side of the sphere)", d)
	}
	if lin.Shape[0] != 0 {
		t.Errorf("the corner pixel hit return %d, want none", lin.Shape[0])
	}
}

func TestRenderLit_ProgressRisesToTheEnd(t *testing.T) {
	var shares []uint32
	var total uint32
	_, err := compiler.RenderLit(compiler.LitRequest{
		Source: "return sphere(5);", Width: 96, Height: 72, WorldExtent: 40, Camera: litCamera(),
	}, []string{"."}, func(stage, label string, completed, tot uint32) {
		if stage != "render" {
			t.Errorf("stage %q, want render", stage)
		}
		shares = append(shares, completed)
		total = tot
	})
	if err != nil {
		t.Fatalf("RenderLit: %v", err)
	}
	if len(shares) == 0 {
		t.Fatal("no progress was reported")
	}
	if total != 1000 {
		t.Errorf("total is %d, want thousandths (1000)", total)
	}
	for i := 1; i < len(shares); i++ {
		if shares[i] < shares[i-1] {
			t.Fatalf("progress went back: %v", shares)
		}
	}
	if shares[len(shares)-1] != total {
		t.Errorf("the last share is %d of %d, want the whole", shares[len(shares)-1], total)
	}
}

func TestRenderLit_RefusesWhatItWillNotDraw(t *testing.T) {
	good := func() compiler.LitRequest {
		return compiler.LitRequest{Source: "return sphere(5);", Width: 8, Height: 8, WorldExtent: 40, Camera: litCamera()}
	}
	cases := []struct {
		name string
		edit func(*compiler.LitRequest)
		want string
	}{
		{"no width", func(r *compiler.LitRequest) { r.Width = 0 }, "out of range"},
		{"too many pixels", func(r *compiler.LitRequest) { r.Width, r.Height = 5000, 5000 }, "out of range"},
		{"no camera", func(r *compiler.LitRequest) { r.Camera = nil }, "needs the view's camera"},
		{"unknown mode", func(r *compiler.LitRequest) { r.Mode = "glossy" }, `unknown mode "glossy"`},
		{"simple without constants", func(r *compiler.LitRequest) { r.Mode = "simple" }, "needs its constants"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := good()
			c.edit(&req)
			_, err := compiler.RenderLit(req, []string{"."}, nil)
			if !errors.Is(err, compiler.ErrLitRequest) {
				t.Fatalf("error %v, want one wrapping ErrLitRequest", err)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("error %q does not say %q", err, c.want)
			}
		})
	}
}

func TestRenderLit_ADocumentThatDoesNotCompileIsNotABadRequest(t *testing.T) {
	_, err := compiler.RenderLit(compiler.LitRequest{
		Source: "return y;", Width: 8, Height: 8, WorldExtent: 40, Camera: litCamera(),
	}, []string{"."}, nil)
	if err == nil || !strings.Contains(err.Error(), "undefined identifier: y") {
		t.Fatalf("error %v, want the compile error", err)
	}
	if errors.Is(err, compiler.ErrLitRequest) {
		t.Errorf("a compile error was reported as a bad request: %v", err)
	}
}
