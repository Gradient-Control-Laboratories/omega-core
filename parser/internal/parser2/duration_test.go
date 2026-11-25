package parser2_test

import (
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/parser/internal/parser2"
)

func TestParseDuration(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want float64
		line int32
	}{
		{"under the scene", "@scene(20)\n@duration(42.5)\n@caption(\"Opening\", 0, 4)\nimport \"std/math\";\nreturn sphere(1);", 42.5, 2},
		{"above the scene", "@duration(8)\n@scene(20, 2)\nreturn sphere(1);", 8, 1},
		{"with no scene", "@duration(8)\nr = 1;\nreturn sphere(r);", 8, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			pkg, _, err := parser2.Parse([]string{c.src})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if pkg.Duration == nil {
				t.Fatal("got no duration")
			}
			if pkg.Duration.Seconds != c.want {
				t.Errorf("duration = %v s, want %v", pkg.Duration.Seconds, c.want)
			}
			if got := pkg.Duration.Start.GetLine(); got != c.line {
				t.Errorf("duration starts on line %d, want %d", got, c.line)
			}
		})
	}
}

func TestParseDurationLeavesTheRest(t *testing.T) {
	pkg, _, err := parser2.Parse([]string{"@scene(20, 2)\n@duration(8)\n@caption(\"Opening\", 0, 4)\nr = 1;\nreturn sphere(r);"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkg.Scene == nil || pkg.Scene.Size != 20 || pkg.Scene.Grid != 2 {
		t.Errorf("scene = %+v, want size 20 grid 2", pkg.Scene)
	}
	if len(pkg.Captions) != 1 || pkg.Captions[0].Text != "Opening" {
		t.Errorf("captions = %+v, want the one caption", pkg.Captions)
	}
	if len(pkg.Files[0].Assignments) != 1 {
		t.Errorf("got %d assignments, want 1", len(pkg.Files[0].Assignments))
	}
}

func TestParseNoDuration(t *testing.T) {
	pkg, _, err := parser2.Parse([]string{"@scene(20)\nreturn sphere(1);"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkg.Duration != nil {
		t.Errorf("duration = %+v, want none", pkg.Duration)
	}
}

func TestParseDurationErrors(t *testing.T) {
	cases := []struct {
		name string
		srcs []string
		want string
	}{
		{"zero", []string{"@duration(0)\nreturn sphere(1);"}, "duration must be greater than zero"},
		{"negative", []string{"@duration(-5)\nreturn sphere(1);"}, "expected a number of seconds"},
		{"empty", []string{"@duration()\nreturn sphere(1);"}, "expected a number of seconds"},
		{"a name", []string{"t = 5;\n@duration(t)\nreturn sphere(1);"}, "duration"},
		{"a name at the top", []string{"@duration(t)\nreturn sphere(1);"}, "expected a number of seconds"},
		{"two numbers", []string{"@duration(5, 6)\nreturn sphere(1);"}, `expected ")"`},
		{"semicolon", []string{"@duration(5);\nreturn sphere(1);"}, "@duration directive must not end with a semicolon"},
		{"twice", []string{"@duration(5)\n@duration(6)\nreturn sphere(1);"}, "duration"},
		{"after a declaration", []string{"r = 1;\n@duration(5)\nreturn sphere(r);"}, "duration"},
		{"second file", []string{"r = 1;", "@duration(5)\nreturn sphere(r);"}, "@duration directive is only allowed in the first source file"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, _, err := parser2.Parse(c.srcs)
			if err == nil {
				t.Fatalf("expected an error containing %q, got none", c.want)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), c.want)
			}
		})
	}
}
