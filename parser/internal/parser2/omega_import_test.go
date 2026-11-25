package parser2_test

import (
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/parser/internal/parser2"
	"github.com/Gradient-Control-Laboratories/omega-core/program"
)

func TestParseOmegaDotImportWarning(t *testing.T) {
	t.Run("standalone_top_of_file", func(t *testing.T) {
		src := "import . \"omega\";\nreturn sphere(1);\n"
		pkg, diags, err := parser2.Parse([]string{src})
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}
		if len(pkg.Files) != 1 || len(pkg.Files[0].Imports) != 1 {
			t.Fatalf("expected 1 rewritten import, got %+v", pkg.Files)
		}
		imp := pkg.Files[0].Imports[0]
		if imp.Path != "vector" || imp.Prefix != "." {
			t.Errorf("compatibility rewrite = (%q, %q), want (\".\", \"vector\")", imp.Prefix, imp.Path)
		}
		if len(diags) != 1 {
			t.Fatalf("len(diags) = %d, want 1 (%+v)", len(diags), diags)
		}
		d := diags[0]
		if d.Severity != program.SeverityWarning {
			t.Errorf("Severity = %v, want SeverityWarning", d.Severity)
		}
		if d.LineNum != 1 || d.Column != 1 {
			t.Errorf("location = %d:%d, want 1:1", d.LineNum, d.Column)
		}
		if d.Line != `import . "omega";` {
			t.Errorf("Line = %q, want %q", d.Line, `import . "omega";`)
		}
		if d.Length != len(`import . "omega";`) {
			t.Errorf("Length = %d, want %d", d.Length, len(`import . "omega";`))
		}
		for _, sub := range []string{`import . "omega"`, "builtin", "gsl fix"} {
			if !strings.Contains(d.Message, sub) {
				t.Errorf("Message %q missing substring %q", d.Message, sub)
			}
		}
	})

	t.Run("indented_with_inline_comment_and_crlf", func(t *testing.T) {
		src := "// Header\r\n  import . \"omega\"; // legacy\r\nreturn sphere(1);\r\n"
		_, diags, err := parser2.Parse([]string{src})
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}
		if len(diags) != 1 {
			t.Fatalf("len(diags) = %d, want 1 (%+v)", len(diags), diags)
		}
		d := diags[0]
		if d.Severity != program.SeverityWarning {
			t.Errorf("Severity = %v, want SeverityWarning", d.Severity)
		}
		if d.LineNum != 2 || d.Column != 3 {
			t.Errorf("location = %d:%d, want 2:3", d.LineNum, d.Column)
		}
		if d.Line != `  import . "omega"; // legacy` {
			t.Errorf("Line = %q, want %q", d.Line, `  import . "omega"; // legacy`)
		}
		if d.Length != len(`import . "omega";`) {
			t.Errorf("Length = %d, want %d", d.Length, len(`import . "omega";`))
		}
	})

	t.Run("multiline_import_span_clamped_to_first_line", func(t *testing.T) {
		src := "import .\n  \"omega\";\nreturn sphere(1);\n"
		_, diags, err := parser2.Parse([]string{src})
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}
		if len(diags) != 1 {
			t.Fatalf("len(diags) = %d, want 1 (%+v)", len(diags), diags)
		}
		d := diags[0]
		if d.LineNum != 1 || d.Column != 1 {
			t.Errorf("location = %d:%d, want 1:1", d.LineNum, d.Column)
		}
		if d.Line != "import ." {
			t.Errorf("Line = %q, want %q", d.Line, "import .")
		}
		if d.Length != len("import .") {
			t.Errorf("Length = %d, want %d", d.Length, len("import ."))
		}
	})

	t.Run("combined_with_legacy_scene_and_multi_file_sources", func(t *testing.T) {
		file0 := "// Scene Size: 20\nimport . \"omega\";\nx = 2;\n"
		file1 := "import . \"omega\";\nreturn sphere(x);\n"
		_, diags, err := parser2.Parse([]string{file0, file1})
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}
		if len(diags) != 3 {
			t.Fatalf("len(diags) = %d, want 3 (%+v)", len(diags), diags)
		}
		if !strings.Contains(diags[0].Message, "@scene") {
			t.Errorf("diags[0] = %+v, expected legacy scene warning", diags[0])
		}
		if !strings.Contains(diags[1].Message, `import . "omega"`) || diags[1].LineNum != 2 {
			t.Errorf("diags[1] = %+v, expected file0 omega import warning at line 2", diags[1])
		}
		if !strings.Contains(diags[2].Message, `import . "omega"`) || diags[2].LineNum != 1 {
			t.Errorf("diags[2] = %+v, expected file1 omega import warning at line 1", diags[2])
		}
	})

	t.Run("clean_imports_emit_no_diagnostics", func(t *testing.T) {
		cases := []struct {
			name string
			src  string
		}{
			{name: "qualified_vector", src: "import \"vector\";\nreturn sphere(1);\n"},
			{name: "dot_vector", src: "import . \"vector\";\nreturn sphere(1);\n"},
			{name: "non_dot_omega", src: "import \"omega\";\nreturn sphere(1);\n"},
			{name: "aliased_omega", src: "import o \"omega\";\nreturn sphere(1);\n"},
			{name: "dot_builtin", src: "import . \"builtin\";\nreturn sphere(1);\n"},
			{name: "comment_and_string_literal", src: "// import . \"omega\";\ns = [\"import . \\\"omega\\\";\"];\nreturn sphere(1);\n"},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				_, diags, err := parser2.Parse([]string{tc.src})
				if err != nil {
					t.Fatalf("unexpected parse error: %v", err)
				}
				if len(diags) != 0 {
					t.Fatalf("expected 0 diagnostics, got %+v", diags)
				}
			})
		}
	})
}
