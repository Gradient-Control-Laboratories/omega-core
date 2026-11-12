package ui_test

import (
	"encoding/json"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/assembler/ui"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/program"
)

func TestExtractWithOperandsIncludesControlCopies(t *testing.T) {
	src := `return uiPoint([1,2,3]);`
	assembled, err := assembler.Assemble([]string{src}, assembler.SourceManager(&source.MockManager{}))
	if err != nil {
		t.Fatal(err)
	}
	// Inlining copies operand identities while retaining source spans.
	file := assembled.Program.Primary.Files[0]
	file.Returns = append(file.Returns, program.CloneOperand(file.Returns[0]))
	controls, operands := ui.ExtractWithOperands(assembled.Program.Primary, assembled.Program.GetImport)
	if len(controls) != 1 || len(controls[0].Params) != 3 {
		t.Fatalf("expected one editable point: %+v", controls)
	}
	if len(operands) < 6 {
		t.Fatalf("expected both inlined control copies, got %d operands", len(operands))
	}
	for _, param := range controls[0].Params {
		copies := 0
		for operand, span := range operands {
			if span != param.Span {
				continue
			}
			if operand.GetValue() != param.Value {
				t.Fatalf("parameter %s captured a different literal", param.ParamName)
			}
			copies++
		}
		if copies < 2 {
			t.Fatalf("parameter %s lost an inlined copy", param.ParamName)
		}
	}
	encoded, err := json.Marshal(controls[0].Params[0])
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	if len(fields) != 3 || fields["paramName"] == nil || fields["span"] == nil || fields["value"] == nil {
		t.Fatalf("operand binding changed the viewer parameter schema: %s", encoded)
	}
}
