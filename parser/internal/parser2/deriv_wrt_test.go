package parser2_test

import (
	"strings"
	"testing"

	lib "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

// derivOf unpacks a DERIV operand into its target and the name of the variable
// it differentiates with respect to.
func derivOf(t *testing.T, op *program.Operand) (*program.Operand, string) {
	t.Helper()
	e, ok := op.T.(*program.Operand_Expression)
	if !ok || e.Expression.Operation != program.Operation_DERIV {
		t.Fatalf("expected a DERIV operand, got %s", op.PrettyPrint(0))
	}
	if len(e.Expression.Operands) != 2 {
		t.Fatalf("DERIV with %d operands", len(e.Expression.Operands))
	}
	wrt, ok := e.Expression.Operands[1].T.(*program.Operand_Reference)
	if !ok || wrt.Reference.Prefix != "" {
		t.Fatalf("DERIV wrt is not a plain reference: %s", e.Expression.Operands[1].PrettyPrint(0))
	}
	return e.Expression.Operands[0], wrt.Reference.Name
}

// TestParseDerivWrtVariable parses the two spellings of a derivative with
// respect to a named variable. `f.d(r)` reaches the build as a qualified call
// (prefix f, function d); `sphere(r).d(r)` and `(f + 1).d(p)` as a postfix
// accessor with arguments. Both lower to the DERIV node .dp and .dtime use,
// with the variable as the second operand, so `.d(p)` is `.dp`.
func TestParseDerivWrtVariable(t *testing.T) {
	src := "r = 2;\nf = sphere(r);\ng = f.d(r);\nh = sphere(r).d(r);\nk = (f + 1).d(p);\nreturn g, h, k;\n"
	pkg, err := lib.ParsePackage([]string{src})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	assigns := pkg.Files[0].Assignments
	if len(assigns) != 5 {
		t.Fatalf("expected 5 assignments, got %d", len(assigns))
	}

	// g = f.d(r): qualified-call spelling, target is the reference f.
	target, wrt := derivOf(t, assigns[2].Expression)
	if wrt != "r" {
		t.Errorf("g: wrt %q, want r", wrt)
	}
	if ref, ok := target.T.(*program.Operand_Reference); !ok || ref.Reference.Name != "f" || ref.Reference.Prefix != "" {
		t.Errorf("g: target %s, want reference f", target.PrettyPrint(0))
	}

	// h = sphere(r).d(r): postfix on a call.
	target, wrt = derivOf(t, assigns[3].Expression)
	if wrt != "r" {
		t.Errorf("h: wrt %q, want r", wrt)
	}
	if call, ok := target.T.(*program.Operand_FnCall); !ok || call.FnCall.Reference.Name != "sphere" {
		t.Errorf("h: target %s, want call sphere", target.PrettyPrint(0))
	}

	// k = (f + 1).d(p): postfix on a parenthesised expression, wrt p.
	target, wrt = derivOf(t, assigns[4].Expression)
	if wrt != program.PVarName {
		t.Errorf("k: wrt %q, want p", wrt)
	}
	if e, ok := target.T.(*program.Operand_Expression); !ok || e.Expression.Operation != program.Operation_ADD {
		t.Errorf("k: target %s, want f + 1", target.PrettyPrint(0))
	}
}

// TestParseDerivWrtVariableErrors rejects the malformed spellings: a
// non-variable argument, the wrong count, no arguments, and arguments on an
// accessor that is not .d.
func TestParseDerivWrtVariableErrors(t *testing.T) {
	cases := []struct {
		name, src, want string
	}{
		{"literal", "f = 1;\ng = f.d(2);\nreturn g;\n", "name of a variable"},
		{"expression", "r = 1;\nf = 1;\ng = f.d(r + 1);\nreturn g;\n", "name of a variable"},
		{"two_args", "r = 1;\nf = 1;\ng = f.d(r, r);\nreturn g;\n", "one argument"},
		{"no_args_postfix", "r = 1;\ng = (r + 1).d;\nreturn g;\n", "one argument"},
		{"args_on_dp", "r = 1;\ng = (r + 1).dp(r);\nreturn g;\n", "takes no arguments"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := lib.ParsePackage([]string{c.src})
			if err == nil {
				t.Fatalf("expected an error containing %q", c.want)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("error %q does not contain %q", err.Error(), c.want)
			}
		})
	}
}
