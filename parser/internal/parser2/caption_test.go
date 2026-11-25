package parser2_test

import (
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/parser/internal/parser2"
)

func TestParseCaptions(t *testing.T) {
	src := `@scene(20)
@caption("Leading, under the scene", 0, 4)
import "std/math";
r = 1;
@caption("Two\nlines", 2, 3)
@caption("Hard cut", 5, 6, 0)
@caption("Slow fade", 10, 20, 2.5)
return sphere(r);`
	pkg, _, err := parser2.Parse([]string{src})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []struct {
		text             string
		start, end, fade float64
	}{
		{"Leading, under the scene", 0, 4, 0.5},
		{"Two\nlines", 2, 3, 0.5}, // the default fade, at exactly half the interval
		{"Hard cut", 5, 6, 0},
		{"Slow fade", 10, 20, 2.5},
	}
	if len(pkg.Captions) != len(want) {
		t.Fatalf("got %d captions, want %d: %v", len(pkg.Captions), len(want), pkg.Captions)
	}
	for i, w := range want {
		c := pkg.Captions[i]
		if c.Text != w.text || c.StartTime != w.start || c.EndTime != w.end || c.Fade != w.fade {
			t.Errorf("caption %d = {%q %v %v %v}, want {%q %v %v %v}",
				i, c.Text, c.StartTime, c.EndTime, c.Fade, w.text, w.start, w.end, w.fade)
		}
	}
	if pkg.Captions[1].Start.GetLine() != 5 {
		t.Errorf("caption 1 starts on line %d, want 5", pkg.Captions[1].Start.GetLine())
	}
	// The captions are lifted out; the declarations around them still build.
	if len(pkg.Files[0].Assignments) != 1 {
		t.Errorf("got %d assignments, want 1", len(pkg.Files[0].Assignments))
	}
}

func TestParseCaptionDefaultFadeCapped(t *testing.T) {
	pkg, _, err := parser2.Parse([]string{"@caption(\"Blink\", 1, 1.4)\nreturn sphere(1);"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := pkg.Captions[0].Fade; got < 0.2-1e-12 || got > 0.2+1e-12 {
		t.Errorf("fade = %v, want 0.2 (half of a 0.4 s caption)", got)
	}
}

func TestParseCaptionAcrossFiles(t *testing.T) {
	pkg, _, err := parser2.Parse([]string{
		"@caption(\"First file\", 0, 1)\nf(x) { return x; }",
		"@caption(\"Second file\", 1, 2)\nreturn sphere(f(1));",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkg.Captions) != 2 || pkg.Captions[0].Text != "First file" || pkg.Captions[1].Text != "Second file" {
		t.Errorf("got %v, want both files' captions in order", pkg.Captions)
	}
}

func TestParseCaptionErrors(t *testing.T) {
	cases := []struct {
		name, src, want string
	}{
		{"end before start", `@caption("x", 3, 2)` + "\nreturn sphere(1);", "must be greater than its start time"},
		{"equal times", `@caption("x", 2, 2)` + "\nreturn sphere(1);", "must be greater than its start time"},
		{"fade too long", `@caption("x", 0, 2, 1.5)` + "\nreturn sphere(1);", "at most half its interval"},
		{"name for a time", "t0 = 1;\n" + `@caption("x", t0, 2)` + "\nreturn sphere(1);", "must be a number literal"},
		{"unquoted text", "@caption(hello, 0, 2)\nreturn sphere(1);", "caption text in double quotes"},
		{"missing end", `@caption("x", 0)` + "\nreturn sphere(1);", `expected ","`},
		{"too many args", `@caption("x", 0, 2, 0.1, 4)` + "\nreturn sphere(1);", `expected ")"`},
		{"semicolon", `@caption("x", 0, 2);` + "\nreturn sphere(1);", "must not end with a semicolon"},
		{"in a function", "f(x) {\n" + `@caption("x", 0, 2)` + "\nreturn x; }\nreturn sphere(f(1));", "cannot appear inside a function"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := parser2.Parse([]string{tc.src})
			if err == nil {
				t.Fatalf("expected an error containing %q", tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.want)
			}
		})
	}
}

func TestParseNoCaptions(t *testing.T) {
	pkg, _, err := parser2.Parse([]string{"@scene(20)\nreturn sphere(1);"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkg.Captions != nil {
		t.Errorf("got captions %v, want none", pkg.Captions)
	}
}
