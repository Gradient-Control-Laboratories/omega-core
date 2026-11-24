package compile_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/compile"
)

// The viewer reads captions off the compile response, so they must survive
// assembly (dead-code removal rebuilds the package) and reach the JSON under
// the names main.js reads.
func TestCompile_CaptionsReachResponse(t *testing.T) {
	resp := compile.Process(compile.Request{
		ID: "req-captions",
		Source: `@caption("Opening", 0, 3)
r = 5;
unused = 7;
@caption("Grow", 3, 6, 1)
return sphere(r);
`,
	}, nil, nil)
	if len(resp.Errors) > 0 {
		t.Fatalf("unexpected errors: %+v", resp.Errors)
	}
	if len(resp.Captions) != 2 {
		t.Fatalf("got %d captions, want 2: %+v", len(resp.Captions), resp.Captions)
	}
	if c := resp.Captions[1]; c.Text != "Grow" || c.StartTime != 3 || c.EndTime != 6 || c.Fade != 1 {
		t.Errorf("caption 1 = %+v, want Grow 3..6 fade 1", c)
	}

	raw, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"captions":[`, `"text":"Grow"`, `"startTime":3`, `"endTime":6`, `"fade":1`} {
		if !strings.Contains(string(raw), want) {
			t.Errorf("response JSON lacks %s: %s", want, raw)
		}
	}
}

func TestCompile_NoCaptionsOmitted(t *testing.T) {
	resp := compile.Process(compile.Request{ID: "req-none", Source: "return sphere(5);\n"}, nil, nil)
	if len(resp.Errors) > 0 {
		t.Fatalf("unexpected errors: %+v", resp.Errors)
	}
	raw, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), `"captions"`) {
		t.Errorf("expected no captions key, got %s", raw)
	}
}
