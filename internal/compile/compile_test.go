package compile_test

import (
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
