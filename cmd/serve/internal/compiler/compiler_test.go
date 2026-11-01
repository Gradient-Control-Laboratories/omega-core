package compiler_test

import (
	"encoding/binary"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/cmd/serve/internal/compiler"
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
	resp := compiler.Compile(compiler.CompileRequest{
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
	resp := compiler.Compile(compiler.CompileRequest{
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
	resp := compiler.Compile(compiler.CompileRequest{
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
	resp := compiler.Compile(compiler.CompileRequest{
		ID:     "req-empty",
		Source: "   \n  ",
	}, nil, nil)
	if len(resp.Errors) == 0 {
		t.Fatalf("expected compile error for empty source with no return statement")
	}
}

func TestMesh_BinarySTLAndProgress(t *testing.T) {
	src := `return sphere(5);`
	var meshStages []string
	stlBytes, err := compiler.Mesh(compiler.MeshRequest{
		ID:          "mesh-1",
		Source:      src,
		Subdivision: 2,
		Bbox:        [6]float32{-8, -8, -8, 8, 8, 8},
		HasBbox:     true,
		Algo:        "marching-cubes",
		Time:        0,
	}, []string{"."}, func(stage, label string, completed, total uint32) {
		meshStages = append(meshStages, stage)
	})
	if err != nil {
		t.Fatalf("unexpected Mesh error: %v", err)
	}
	if len(stlBytes) < 84 {
		t.Fatalf("expected binary STL >= 84 bytes, got %d", len(stlBytes))
	}
	triCount := binary.LittleEndian.Uint32(stlBytes[80:84])
	if triCount == 0 {
		t.Errorf("expected non-zero triangle count in binary STL")
	}
	expectedLen := 84 + int(triCount)*50
	if len(stlBytes) != expectedLen {
		t.Errorf("expected STL byte length %d for %d triangles, got %d", expectedLen, triCount, len(stlBytes))
	}
	if len(meshStages) == 0 {
		t.Errorf("expected progress callbacks during Mesh")
	}
}

func TestMesh_ErrorOnInvalidSource(t *testing.T) {
	_, err := compiler.Mesh(compiler.MeshRequest{
		ID:     "mesh-err",
		Source: "return nonexistent(1);",
	}, nil, nil)
	if err == nil {
		t.Fatalf("expected error from Mesh on invalid GSL source")
	}
}
