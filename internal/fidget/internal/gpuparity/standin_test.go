package gpuparity

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/eval"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/source"
)

// Shaft Modeling, Phase 22: stand-ins anywhere, on the GPU. The weight is
// measured at the point, so a long part simplifies its far end first, and a
// screw placed through a remap keeps its library's stand-in. The Simple rung
// is 16 pixels here, the pixel scale 0.001: a stand-in sized 10 is the part
// nearer than 625 and the stand-in beyond 1250.

func standInGPU(t *testing.T, doc string, eye [3]float64, pts [][3]float32) []float32 {
	t.Helper()
	bin := filepath.Join(repoRoot, "target/release/gpu-eval")
	if _, err := os.Stat(bin); err != nil {
		t.Skipf("gpu-eval is not built (tool/build-fidget builds it): %v", err)
	}
	e := eval.New(&source.MockManager{}, nil)
	res, err := e.Shader(t.Context(), nil, doc, "", eval.ShaderOptions{Target: eval.ShaderWGSL})
	if err != nil {
		t.Fatal(err)
	}
	shader := pruneShader(t, res, eye, 0.001)
	if strings.Contains(res.Code, "var<private> gsl_simpleRung") {
		line := "    gsl_vanishView = vec4<f32>(0.0, 1.0, 2.0, 0.0);"
		shader = strings.Replace(shader, line, "    gsl_simpleRung = 16.0;\n"+line, 1)
	}
	return runGPU(t, bin, shader, pts, 1, false, "")
}

// A bar 10 x 10 x 400 along z with a hole 4 across through it near each end;
// its stand-in is the plain bar. From z = 900 the near hole, 710 away, is
// drawn, and the far one, 1090 away, is the stand-in's solid; one weight for
// the whole part would draw both.
const standInBarDoc = `bar = max(box([-5.0, -5.0, -200.0], [5.0, 5.0, 200.0]),
          neg(min(length(p.yz - [0.0, 190.0]) - 2.0, length(p.yz - [0.0, -190.0]) - 2.0)));
bar.simple = box([-5.0, -5.0, -200.0], [5.0, 5.0, 200.0]);
return bar;
`

func TestGpuStandIn_FarEndSimplifiesFirst(t *testing.T) {
	got := standInGPU(t, standInBarDoc, [3]float64{0, 0, 900}, [][3]float32{{0, 0, 190}, {0, 0, -190}})
	if got[0] <= 0 {
		t.Errorf("the near hole reads %g, want it open (> 0)", got[0])
	}
	if got[1] >= 0 {
		t.Errorf("the far hole reads %g, want the stand-in's solid (< 0)", got[1])
	}
}

// An M6 cap screw placed through a remap: in its socket the screw is open
// near and, from far away, the stand-in's solid head.
func TestGpuStandIn_RemappedScrewSwitches(t *testing.T) {
	doc := "screw = remap(fastener.shcs(6.0, 20.0), [p.x - 50.0, p.y, p.z]);\nreturn screw;\n"
	pt := [][3]float32{{50, 0, 5}}
	for _, c := range []struct {
		dist float64
		open bool
	}{{100, true}, {3000, false}} {
		d := standInGPU(t, doc, [3]float64{50, -c.dist, 5}, pt)[0]
		if (d > 0) != c.open {
			t.Errorf("%s: the socket reads %g", fmt.Sprintf("from %g", c.dist), d)
		}
	}
	_ = math.Abs
}
