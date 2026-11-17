package compile_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/internal/compile"
	"gradientcontrol.com/omega/core/program"
)

func TestCompile_StandardSDFWithSlidersAndParts(t *testing.T) {
	src := `
r = 10;
r.min = 2;
r.max = 20;
body = sphere(r);
return body;
`
	var stages []string
	resp := compile.Process(compile.Request{
		ID:     "req-1",
		Source: src,
	}, nil, func(stage, label string, completed, total uint32) {
		stages = append(stages, stage+":"+label)
	})

	if resp.ID != "req-1" {
		t.Errorf("expected ID 'req-1', got %q", resp.ID)
	}
	if len(resp.Errors) > 0 {
		t.Fatalf("unexpected compile errors: %+v", resp.Errors)
	}
	if !strings.Contains(resp.WGSL, "fn sdShape") {
		t.Errorf("expected WGSL to contain 'fn sdShape', got:\n%s", resp.WGSL)
	}
	if resp.ShapeCount != 1 {
		t.Errorf("expected ShapeCount 1, got %d", resp.ShapeCount)
	}
	if len(resp.Sliders) != 1 || resp.Sliders[0].VariableName != "r" {
		t.Errorf("expected 1 slider 'r', got %+v", resp.Sliders)
	}
	if len(stages) == 0 {
		t.Errorf("expected progress callbacks during Compile")
	}
}

func TestCompile_WithAudioBlock(t *testing.T) {
	src := `
return sphere(5), audio(sin(440 * 2 * pi * time), 2);
`
	resp := compile.Process(compile.Request{
		ID:     "req-audio",
		Source: src,
	}, nil, nil)

	if len(resp.Errors) > 0 {
		t.Fatalf("unexpected compile errors: %+v", resp.Errors)
	}
	if resp.ShapeCount != 1 {
		t.Errorf("expected ShapeCount 1, got %d", resp.ShapeCount)
	}
	if resp.Audio == nil {
		t.Fatalf("expected non-nil Audio result")
	}
	if resp.Audio.JSCode == "" {
		t.Errorf("expected non-empty Audio.JSCode")
	}
	if resp.Audio.Duration <= 0 {
		t.Errorf("expected positive Audio.Duration, got %d", resp.Audio.Duration)
	}
}

func TestCompile_StructuredDiagnosticsOnError(t *testing.T) {
	src := `
x = 10;
return unknownFunc(x);
`
	resp := compile.Process(compile.Request{
		ID:     "req-err",
		Source: src,
	}, nil, nil)

	if resp.ID != "req-err" {
		t.Errorf("expected ID 'req-err', got %q", resp.ID)
	}
	if len(resp.Errors) == 0 {
		t.Fatalf("expected structured errors on undefined function")
	}
	errItem := resp.Errors[0]
	if errItem.LineNum != 3 {
		t.Errorf("expected LineNum 3, got %d", errItem.LineNum)
	}
	if !strings.Contains(errItem.Line, "unknownFunc") {
		t.Errorf("expected Line context containing 'unknownFunc', got %q", errItem.Line)
	}
	if errItem.Message == "" {
		t.Errorf("expected non-empty error Message")
	}
}

func TestCompile_EmptySourceError(t *testing.T) {
	resp := compile.Process(compile.Request{
		ID:     "req-empty",
		Source: "   \n  ",
	}, nil, nil)
	if len(resp.Errors) == 0 {
		t.Fatalf("expected compile error for empty source with no return statement")
	}
}

func TestCompile_AudioOnlyWithSliders(t *testing.T) {
	src := `
freq = 440;
freq.min = 100;
freq.max = 1000;
return audio(sin(freq * 2 * pi * time));
`
	resp := compile.Process(compile.Request{
		ID:     "req-audio-only",
		Source: src,
	}, nil, nil)

	if len(resp.Errors) > 0 {
		t.Fatalf("unexpected compile errors: %+v", resp.Errors)
	}
	if resp.WGSL != "" {
		t.Errorf("expected empty WGSL for audio-only document, got:\n%s", resp.WGSL)
	}
	if resp.ShapeCount != 0 {
		t.Errorf("expected ShapeCount 0, got %d", resp.ShapeCount)
	}
	if resp.Audio == nil || resp.Audio.JSCode == "" {
		t.Fatalf("expected non-nil Audio with JSCode")
	}
	if len(resp.Sliders) != 1 || resp.Sliders[0].VariableName != "freq" {
		t.Errorf("expected 1 slider 'freq', got %+v", resp.Sliders)
	}
}

func TestCompile_AudioCodegenErrorAndStereo(t *testing.T) {
	t.Run("unsupported_audio_builtin_reports_error", func(t *testing.T) {
		resp := compile.Process(compile.Request{
			ID:     "req-audio-err",
			Source: "return audio(worley([time, 0, 0]));",
		}, nil, nil)
		if len(resp.Errors) == 0 {
			t.Fatalf("expected error when audio expression uses unsupported builtin, got %+v", resp)
		}
		if resp.Audio != nil {
			t.Errorf("expected nil Audio on codegen failure, got %+v", resp.Audio)
		}
	})

	t.Run("stereo_audio_channels", func(t *testing.T) {
		resp := compile.Process(compile.Request{
			ID:     "req-stereo",
			Source: "return audio(sin(440 * 2 * pi * time), cos(440 * 2 * pi * time));",
		}, nil, nil)
		if len(resp.Errors) > 0 {
			t.Fatalf("unexpected compile errors for stereo audio: %+v", resp.Errors)
		}
		if resp.Audio == nil || resp.Audio.JSCode == "" {
			t.Fatalf("expected stereo Audio JSCode, got %+v", resp.Audio)
		}
	})
}

func TestCompile_WithSourcePaths(t *testing.T) {
	root := t.TempDir()
	pkgDir := filepath.Join(root, "mypkg")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "mypkg.gsl"), []byte("rad = 4;\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	src := "import \"mypkg\";\nreturn sphere(mypkg.rad);\n"
	resp := compile.Process(compile.Request{
		ID:     "req-import",
		Source: src,
	}, []string{root}, nil)
	if len(resp.Errors) > 0 {
		t.Fatalf("unexpected compile errors with sourcePaths: %+v", resp.Errors)
	}
	if resp.ShapeCount != 1 || !strings.Contains(resp.WGSL, "fn sdShape") {
		t.Errorf("expected compiled WGSL for imported package, got ShapeCount=%d", resp.ShapeCount)
	}
}

func TestUnwrapProgramErrors(t *testing.T) {
	if got := compile.UnwrapProgramErrors("return sphere(1);", nil); len(got) != 0 {
		t.Fatalf("expected empty slice for nil error, got %+v", got)
	}

	src := "a = 1 + ;\nb = bad.call(1);\n"
	joined := errors.Join(
		&program.Error{LineNum: 1, Column: 7, Message: "unexpected semicolon"},
		&program.Error{LineNum: 2, Column: 5, Message: "undefined symbol"},
		errors.New("plain fallback failure"),
	)
	got := compile.UnwrapProgramErrors(src, joined)
	if len(got) != 3 {
		t.Fatalf("expected 3 unwrapped errors, got %d: %+v", len(got), got)
	}
	// '+' at column 7 is a non-word character -> Length 1.
	if got[0].Line != "a = 1 + ;" || got[0].Length != 1 {
		t.Errorf("got[0] = %+v, want Line=%q Length=1", got[0], "a = 1 + ;")
	}
	// 'bad.call' at column 5 spans 8 characters including '.'.
	if got[1].Line != "b = bad.call(1);" || got[1].Length != len("bad.call") {
		t.Errorf("got[1] = %+v, want Line=%q Length=%d", got[1], "b = bad.call(1);", len("bad.call"))
	}
	// Plain error falls back to 1:1 with length 1.
	if got[2].LineNum != 1 || got[2].Column != 1 || got[2].Length != 1 || got[2].Message != "plain fallback failure" {
		t.Errorf("got[2] = %+v, want 1:1 plain fallback failure", got[2])
	}
}

func TestCompile_WarningsDoNotBlockWGSLOrAudio(t *testing.T) {
	t.Run("primary_package_builtin_shadow_warning", func(t *testing.T) {
		src := "min(radius, fallback) {\n    return radius;\n}\nreturn sphere(min(5.0, 2.0));\n"
		resp := compile.Process(compile.Request{
			ID:     "req-warn-shadow",
			Source: src,
		}, nil, nil)

		if len(resp.Errors) > 0 {
			t.Fatalf("expected no compile errors when warnings are present, got %+v", resp.Errors)
		}
		if resp.ShapeCount != 1 || !strings.Contains(resp.WGSL, "fn sdShape") {
			t.Fatalf("expected WGSL shader alongside warnings, got ShapeCount=%d WGSL=%q", resp.ShapeCount, resp.WGSL)
		}
		if len(resp.Warnings) != 1 {
			t.Fatalf("expected 1 warning, got %+v", resp.Warnings)
		}
		w := resp.Warnings[0]
		if w.PackagePath != "" {
			t.Errorf("expected empty PackagePath for primary package warning, got %q", w.PackagePath)
		}
		if w.LineNum != 1 || w.Column != 5 || w.Length != len("radius") {
			t.Errorf("unexpected warning span: got %d:%d len=%d, want 1:5 len=%d", w.LineNum, w.Column, w.Length, len("radius"))
		}
		if w.Line != "min(radius, fallback) {" {
			t.Errorf("unexpected warning Line %q", w.Line)
		}
		if !strings.Contains(w.Message, `shadows implicitly imported builtin function`) {
			t.Errorf("unexpected warning Message %q", w.Message)
		}
	})

	t.Run("dot_import_collision_in_primary_and_dependency", func(t *testing.T) {
		root := t.TempDir()
		for pkg, body := range map[string]string{
			"pkgA": "x = 3.0;\n",
			"pkgB": "x = 4.0;\n",
			"dep":  "// dep header\nimport . \"pkgA\";\nimport . \"pkgB\";\nr = pkgA.x + pkgB.x;\n",
		} {
			dir := filepath.Join(root, pkg)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, pkg+".gsl"), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}

		primarySrc := "import . \"pkgA\";\nimport . \"pkgB\";\nreturn sphere(pkgA.x + pkgB.x);\n"
		primaryResp := compile.Process(compile.Request{
			ID:     "req-warn-dot",
			Source: primarySrc,
		}, []string{root}, nil)
		if len(primaryResp.Errors) > 0 {
			t.Fatalf("unexpected errors: %+v", primaryResp.Errors)
		}
		if len(primaryResp.Warnings) != 1 {
			t.Fatalf("expected 1 primary dot-import warning, got %+v", primaryResp.Warnings)
		}
		pw := primaryResp.Warnings[0]
		if pw.LineNum != 2 || pw.Column != 1 || pw.Length != len("import") || pw.Line != `import . "pkgB";` {
			t.Errorf("unexpected primary dot-import warning: %+v", pw)
		}

		depSrc := "import \"dep\";\nreturn sphere(dep.r);\n"
		depResp := compile.Process(compile.Request{
			ID:     "req-warn-dep",
			Source: depSrc,
		}, []string{root}, nil)
		if len(depResp.Errors) > 0 {
			t.Fatalf("unexpected errors: %+v", depResp.Errors)
		}
		if len(depResp.Warnings) != 1 {
			t.Fatalf("expected 1 dependency warning, got %+v", depResp.Warnings)
		}
		dw := depResp.Warnings[0]
		if dw.PackagePath != "dep" {
			t.Errorf("expected PackagePath 'dep', got %q", dw.PackagePath)
		}
		if dw.LineNum != 3 || dw.Line != `import . "pkgB";` {
			t.Errorf("expected dependency warning Line from dep.gsl line 3, got %+v", dw)
		}
	})

	t.Run("audio_only_with_warning_emits_audio", func(t *testing.T) {
		src := "min(a, b) {\n    return a;\n}\nreturn audio(sin(min(440.0, 220.0) * 2 * pi * time));\n"
		resp := compile.Process(compile.Request{
			ID:     "req-warn-audio",
			Source: src,
		}, nil, nil)
		if len(resp.Errors) > 0 {
			t.Fatalf("unexpected errors: %+v", resp.Errors)
		}
		if resp.Audio == nil || resp.Audio.JSCode == "" {
			t.Fatalf("expected Audio JSCode alongside warnings, got %+v", resp.Audio)
		}
		if len(resp.Warnings) != 1 {
			t.Fatalf("expected 1 warning, got %+v", resp.Warnings)
		}
	})

	t.Run("clean_source_omits_warnings_in_json", func(t *testing.T) {
		resp := compile.Process(compile.Request{
			ID:     "req-clean",
			Source: "return sphere(5);\n",
		}, nil, nil)
		if len(resp.Warnings) != 0 {
			t.Fatalf("expected no warnings on clean source, got %+v", resp.Warnings)
		}
		raw, err := json.Marshal(resp)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(raw), `"warnings"`) {
			t.Errorf("expected clean response JSON to omit warnings, got %s", string(raw))
		}
	})

	t.Run("legacy_scene_comment_parser_warning_flows_to_bundle", func(t *testing.T) {
		resp := compile.Process(compile.Request{
			ID:     "req-legacy-scene",
			Source: "// Scene Size: 40\nreturn sphere(5);\n",
		}, nil, nil)
		if len(resp.Errors) > 0 {
			t.Fatalf("unexpected errors: %+v", resp.Errors)
		}
		if resp.Scene == nil || resp.Scene.Size != 40 {
			t.Fatalf("expected Scene.Size 40, got %+v", resp.Scene)
		}
		if len(resp.Warnings) != 1 {
			t.Fatalf("expected 1 legacy scene warning, got %+v", resp.Warnings)
		}
		w := resp.Warnings[0]
		if w.LineNum != 1 || w.Column != 1 || w.Length != len("// Scene Size: 40") {
			t.Errorf("unexpected legacy scene warning span: %+v", w)
		}
	})
}
