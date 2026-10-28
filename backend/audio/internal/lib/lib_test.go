package lib

import (
	"testing"
)

func TestExtractAudio(t *testing.T) {
	gslSource := `freq = 440;
return audio(sin(freq * time));`
	audioSource, otherSource, err := ExtractAudio(gslSource)
	if err != nil {
		t.Fatalf("ExtractAudio failed: %v", err)
	}
	expectedAudio := `freq = 440;
return sin(freq * time);`
	if audioSource != expectedAudio {
		t.Errorf("expected audioSource %q, got %q", expectedAudio, audioSource)
	}
	if otherSource != "" {
		t.Errorf("expected empty otherSource, got %q", otherSource)
	}
}
