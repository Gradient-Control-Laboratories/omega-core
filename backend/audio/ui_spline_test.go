package audio_test

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/backend/audio"
	"gradientcontrol.com/omega/core/internal/source"
)

func TestSplinePreservesAudioCode(t *testing.T) {
	code := `m = [[1,0,0,0],[0,1,0,0],[0,0,1,0],[3,4,5,1]];
path = uiSpline([m, m, m]);
last = path[2];
position = last[3];
return sin(position.x * time);`
	var plain string
	for _, src := range []string{strings.Replace(code, "uiSpline([m, m, m])", "[m, m, m]", 1), code} {
		res, err := assembler.Assemble([]string{src}, assembler.SourceManager(&source.MockManager{}))
		if err != nil {
			t.Fatal(err)
		}
		js, _, err := audio.AudioToJs(res.Program)
		if err != nil {
			t.Fatal(err)
		}
		if plain != "" && js != plain {
			t.Fatal("spline annotation changed generated audio")
		}
		plain = js
	}
}
