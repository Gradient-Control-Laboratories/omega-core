package codegen_test

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
)

func renderWGSL(t *testing.T, src string) string {
	t.Helper()
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(mockMgr))
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	rr, err := renderHelper(src, res.Program)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	return rr.WGSL
}

// A program using the noise library threads the noise tile texture into the
// shader (sdShape + the cellular helper) and samples it via textureLoad.
func TestNoiseHash3ThreadsTileTexture(t *testing.T) {
	w := renderWGSL(t, "import \"noise\";\nreturn noise.cellular(5) - 1;")

	if !strings.Contains(w, "noiseTex: texture_3d<f32>") {
		t.Errorf("expected noiseTex parameter threaded into the shader; got:\n%s", w)
	}
	if !strings.Contains(w, "textureLoad(noiseTex") {
		t.Errorf("expected noiseHash3 to lower to textureLoad(noiseTex, ...); got:\n%s", w)
	}
	// sdShape must carry the param so the browser can supply the texture.
	if !strings.Contains(w, "fn sdShape(") || !strings.Contains(w[strings.Index(w, "fn sdShape("):], "noiseTex") {
		t.Errorf("expected sdShape signature to include noiseTex; got:\n%s", w)
	}
}

// A program that doesn't touch noise must not pay for the noise tile.
func TestNoNoiseNoTileThreading(t *testing.T) {
	w := renderWGSL(t, "return length(p) - 1;")
	if strings.Contains(w, "noiseTex") {
		t.Errorf("non-noise program should not thread noiseTex; got:\n%s", w)
	}
}
