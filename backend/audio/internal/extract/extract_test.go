package extract

import (
	"testing"
)

func TestExtractAudio(t *testing.T) {
	tests := []struct {
		name      string
		gslSource string
		wantAudio string
		wantOther string
		wantErr   bool
	}{
		{
			name: "Standard single argument audio",
			gslSource: `freq = 440;
return audio(sin(freq * time));`,
			wantAudio: `freq = 440;
return sin(freq * time);`,
			wantOther: "",
			wantErr:   false,
		},
		{
			name: "Two arguments audio",
			gslSource: `freq = 440;
return audio(sin(freq * time), cos(freq * time));`,
			wantAudio: `freq = 440;
return sin(freq * time), cos(freq * time);`,
			wantOther: "",
			wantErr:   false,
		},
		{
			name: "Multiple returns with audio last",
			gslSource: `val1 = 123;
val2 = 456;
return val1, val2, audio(sin(time));`,
			wantAudio: `val1 = 123;
val2 = 456;
return sin(time);`,
			wantOther: `val1 = 123;
val2 = 456;
return val1, val2;`,
			wantErr: false,
		},
		{
			name: "Multiple returns with audio last and spaces",
			gslSource: `val1 = 123;
return val1  ,   audio(sin(time))   ;`,
			wantAudio: `val1 = 123;
return sin(time);`,
			wantOther: `val1 = 123;
return val1;`,
			wantErr: false,
		},
		{
			name: "No return statement",
			gslSource: `freq = 440;
x = sin(freq * time);`,
			wantAudio: "",
			wantOther: "",
			wantErr:   false,
		},
		{
			name: "No semicolon terminating return",
			gslSource: `freq = 440;
return audio(sin(freq * time))`,
			wantAudio: "",
			wantOther: "",
			wantErr:   false,
		},
		{
			name: "No audio returns",
			gslSource: `val = 123;
return val;`,
			wantAudio: "",
			wantOther: "",
			wantErr:   false,
		},
		{
			name:      "Error: audio not the last element",
			gslSource: `return audio(sin(time)), val;`,
			wantAudio: "",
			wantOther: "",
			wantErr:   true,
		},
		{
			name:      "Error: 0 arguments to audio",
			gslSource: `return audio();`,
			wantAudio: "",
			wantOther: "",
			wantErr:   true,
		},
		{
			name:      "Error: 3 arguments to audio",
			gslSource: `return audio(a, b, c);`,
			wantAudio: "",
			wantOther: "",
			wantErr:   true,
		},
		{
			name:      "Complex nested parentheses and brackets in audio arguments",
			gslSource: `return val1, audio(voices[0], mix(a, b, list[1]));`,
			wantAudio: `return voices[0], mix(a, b, list[1]);`,
			wantOther: `return val1;`,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAudio, gotOther, err := ExtractAudio(tt.gslSource)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ExtractAudio() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if gotAudio != tt.wantAudio {
					t.Errorf("ExtractAudio() gotAudio = %q, want %q", gotAudio, tt.wantAudio)
				}
				if gotOther != tt.wantOther {
					t.Errorf("ExtractAudio() gotOther = %q, want %q", gotOther, tt.wantOther)
				}
			}
		})
	}
}
