package audio_test

import (
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/assembler"
	"github.com/Gradient-Control-Laboratories/omega-core/assembler/variation"
	"github.com/Gradient-Control-Laboratories/omega-core/backend/audio"
	"github.com/Gradient-Control-Laboratories/omega-core/program"
)

func TestAudioToJs(t *testing.T) {
	src := `return audio(sin(time * 440));`
	res, err := assembler.Assemble([]string{src})
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	if len(res.AudioReturns) == 0 {
		t.Fatalf("expected non-empty AudioReturns")
	}
	audioOps := make([]*program.Operand, len(res.AudioReturns))
	for i, ar := range res.AudioReturns {
		audioOps[i] = ar.Return
	}
	audioProg := variation.Program(res.Program, audioOps...)
	js, channels, err := audio.AudioToJs(audioProg)
	if err != nil {
		t.Fatalf("AudioToJs failed: %v", err)
	}
	if channels < 1 {
		t.Errorf("expected at least 1 channel, got %d", channels)
	}
	if !strings.Contains(js, "function") && !strings.Contains(js, "Math.sin") && !strings.Contains(js, "return") {
		t.Errorf("expected JavaScript audio function, got: %s", js)
	}
}
