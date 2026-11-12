package audio_test

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/backend/audio"
)

func TestExtractAudio(t *testing.T) {
	src := `
x = 10;
return sphere(x), audio(sin(t));
`
	audioSrc, otherSrc, err := audio.ExtractAudio(src)
	if err != nil {
		t.Fatalf("ExtractAudio failed: %v", err)
	}
	if !strings.Contains(audioSrc, "sin(t)") {
		t.Errorf("expected audioSrc to contain sin(t), got: %s", audioSrc)
	}
	if !strings.Contains(otherSrc, "sphere(x)") {
		t.Errorf("expected otherSrc to contain sphere(x), got: %s", otherSrc)
	}

	monoSrc := "freq = 440;\nreturn audio(sin(freq * time));"
	audioOnly, otherEmpty, err := audio.ExtractAudio(monoSrc)
	if err != nil {
		t.Fatalf("ExtractAudio mono failed: %v", err)
	}
	if want := "freq = 440;\nreturn sin(freq * time);"; audioOnly != want {
		t.Errorf("expected audioSource %q, got %q", want, audioOnly)
	}
	if otherEmpty != "" {
		t.Errorf("expected empty otherSource, got %q", otherEmpty)
	}
}

func TestAudioToJs(t *testing.T) {
	src := `return audio(sin(time * 440));`
	audioSrc, _, err := audio.ExtractAudio(src)
	if err != nil {
		t.Fatalf("ExtractAudio failed: %v", err)
	}
	res, err := assembler.Assemble([]string{audioSrc})
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	js, channels, err := audio.AudioToJs(res.Program)
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
