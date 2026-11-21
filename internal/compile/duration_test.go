package compile_test

import (
	"encoding/json"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/internal/compile"
)

// The viewer reads the duration off the compile response, so it must survive
// assembly (dead-code removal rebuilds the package) and reach the JSON under
// the names main.js reads.
func TestCompile_DurationReachesResponse(t *testing.T) {
	resp := compile.Process(compile.Request{
		ID: "req-duration",
		Source: `@scene(60)
@duration(12.5)
r = 5 + time;
unused = 7;
return sphere(r);
`,
	}, nil, nil)
	if len(resp.Errors) > 0 {
		t.Fatalf("unexpected errors: %+v", resp.Errors)
	}
	if resp.Duration == nil || resp.Duration.Seconds != 12.5 {
		t.Fatalf("duration = %+v, want 12.5 s", resp.Duration)
	}

	raw, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"duration":{`, `"seconds":12.5`} {
		if !strings.Contains(string(raw), want) {
			t.Errorf("response JSON lacks %s: %s", want, raw)
		}
	}
}

func TestCompile_NoDurationOmitted(t *testing.T) {
	resp := compile.Process(compile.Request{ID: "req-none", Source: "return sphere(5);\n"}, nil, nil)
	if len(resp.Errors) > 0 {
		t.Fatalf("unexpected errors: %+v", resp.Errors)
	}
	raw, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), `"duration"`) {
		t.Errorf("expected no duration key, got %s", raw)
	}
}
