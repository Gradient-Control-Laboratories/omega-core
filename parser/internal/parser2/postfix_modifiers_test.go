package parser2_test

import (
	"strings"
	"testing"

	lib "github.com/Gradient-Control-Laboratories/omega-core/parser"
	"github.com/Gradient-Control-Laboratories/omega-core/parser/internal/parser2"
	"github.com/Gradient-Control-Laboratories/omega-core/program"
)

func exprOp(t *testing.T, op *program.Operand, want program.Operation) []*program.Operand {
	t.Helper()
	if op == nil {
		t.Fatalf("nil operand, want %v", want)
	}
	e, ok := op.T.(*program.Operand_Expression)
	if !ok || e.Expression.Operation != want {
		t.Fatalf("got %s, want %v", op.PrettyPrint(0), want)
	}
	return e.Expression.Operands
}

func assertRefName(t *testing.T, op *program.Operand, wantPrefix, wantName string) {
	t.Helper()
	r, ok := op.T.(*program.Operand_Reference)
	if !ok || r.Reference.Prefix != wantPrefix || r.Reference.Name != wantName {
		t.Fatalf("got %s, want ref %s.%s", op.PrettyPrint(0), wantPrefix, wantName)
	}
}

func assertVal(t *testing.T, op *program.Operand, want float64) {
	t.Helper()
	v, ok := op.T.(*program.Operand_Value)
	if !ok || v.Value != want {
		t.Fatalf("got %s, want value %v", op.PrettyPrint(0), want)
	}
}

func TestPostfixModifierChainingASTAndLocations(t *testing.T) {
	src := strings.Join([]string{
		`import "pkg";`,
		`m = [[1, 2], [3, 4]];`,
		`pts = [[1, 2, 3], [4, 5, 6]];`,
		`flags = [0, 1];`,
		`t = 2;`,
		`v = [t, 2 * t, 3 * t];`,
		`body = sphere(1);`,
		`f(q) { return length(q) - 1; }`,
		`a = m[0][1];`,
		`b = pts[0].x;`,
		`c = pts[1].yz;`,
		`d = body.box[0];`,
		`e = body.box[1].z;`,
		`g = -m[0][1];`,
		`h = !flags[0];`,
		`j = (f(p)).dp.x;`,
		`k = v.d(t)[0];`,
		`l = pkg.m[0][1];`,
		`n = pkg.body.box[0];`,
		`q = pkg.v.d(t)[1];`,
		`s = ["abc"[0]];`,
		`return a;`,
	}, "\n") + "\n"

	res, err := lib.ParsePackage([]string{src})
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	file := res.Package.Files[0]
	byName := make(map[string]*program.Operand, len(file.Assignments))
	for _, assign := range file.Assignments {
		byName[assign.Variables[0].Name] = assign.Expression
	}

	// a = m[0][1] -> INDEX(INDEX(m, 0), 1)
	{
		outer := byName["a"]
		outerOps := exprOp(t, outer, program.Operation_INDEX)
		assertVal(t, outerOps[1], 1)
		inner := outerOps[0]
		innerOps := exprOp(t, inner, program.Operation_INDEX)
		assertRefName(t, innerOps[0], "", "m")
		assertVal(t, innerOps[1], 0)
		// Span check: `a = m[0][1];` -> `m` is at col 5, `m[0]` ends at col 9, `m[0][1]` ends at col 12.
		if inner.Start.Column != 5 || inner.End.Column != 9 {
			t.Errorf("m[0] span cols = [%d, %d), want [5, 9)", inner.Start.Column, inner.End.Column)
		}
		if outer.Start.Column != 5 || outer.End.Column != 12 {
			t.Errorf("m[0][1] span cols = [%d, %d), want [5, 12)", outer.Start.Column, outer.End.Column)
		}
	}

	// b = pts[0].x -> ACCESS(INDEX(pts, 0), x)
	{
		accOps := exprOp(t, byName["b"], program.Operation_ACCESS)
		assertRefName(t, accOps[1], "", "x")
		idxOps := exprOp(t, accOps[0], program.Operation_INDEX)
		assertRefName(t, idxOps[0], "", "pts")
		assertVal(t, idxOps[1], 0)
	}

	// c = pts[1].yz -> ACCESS(INDEX(pts, 1), yz)
	{
		accOps := exprOp(t, byName["c"], program.Operation_ACCESS)
		assertRefName(t, accOps[1], "", "yz")
		idxOps := exprOp(t, accOps[0], program.Operation_INDEX)
		assertRefName(t, idxOps[0], "", "pts")
		assertVal(t, idxOps[1], 1)
	}

	// d = body.box[0] -> INDEX(BOUNDINGBOX(body), 0)
	{
		idxOps := exprOp(t, byName["d"], program.Operation_INDEX)
		assertVal(t, idxOps[1], 0)
		boxOps := exprOp(t, idxOps[0], program.Operation_BOUNDINGBOX)
		assertRefName(t, boxOps[0], "", "body")
	}

	// e = body.box[1].z -> ACCESS(INDEX(BOUNDINGBOX(body), 1), z)
	{
		accOps := exprOp(t, byName["e"], program.Operation_ACCESS)
		assertRefName(t, accOps[1], "", "z")
		idxOps := exprOp(t, accOps[0], program.Operation_INDEX)
		assertVal(t, idxOps[1], 1)
		boxOps := exprOp(t, idxOps[0], program.Operation_BOUNDINGBOX)
		assertRefName(t, boxOps[0], "", "body")
	}

	// g = -m[0][1] -> NEG(INDEX(INDEX(m, 0), 1)): postfix binds tighter than unary `-`.
	{
		negOp := byName["g"]
		negOps := exprOp(t, negOp, program.Operation_NEG)
		idxOuter := negOps[0]
		outerOps := exprOp(t, idxOuter, program.Operation_INDEX)
		assertVal(t, outerOps[1], 1)
		innerOps := exprOp(t, outerOps[0], program.Operation_INDEX)
		assertRefName(t, innerOps[0], "", "m")
		assertVal(t, innerOps[1], 0)
		// `g = -m[0][1];`: `-` at col 5, `m` at col 6, end at col 13.
		if negOp.Start.Column != 5 || negOp.End.Column != 13 {
			t.Errorf("-m[0][1] span cols = [%d, %d), want [5, 13)", negOp.Start.Column, negOp.End.Column)
		}
		if idxOuter.Start.Column != 6 || idxOuter.End.Column != 13 {
			t.Errorf("inner m[0][1] span cols = [%d, %d), want [6, 13)", idxOuter.Start.Column, idxOuter.End.Column)
		}
	}

	// h = !flags[0] -> NOT(INDEX(flags, 0)): postfix binds tighter than unary `!`.
	{
		notOp := byName["h"]
		notOps := exprOp(t, notOp, program.Operation_NOT)
		idxOps := exprOp(t, notOps[0], program.Operation_INDEX)
		assertRefName(t, idxOps[0], "", "flags")
		assertVal(t, idxOps[1], 0)
		if notOp.Start.Column != 5 || notOps[0].Start.Column != 6 {
			t.Errorf("!flags[0] starts = (%d, %d), want (5, 6)", notOp.Start.Column, notOps[0].Start.Column)
		}
	}

	// j = (f(p)).dp.x -> ACCESS(DERIV(f(p), p), x)
	{
		accOps := exprOp(t, byName["j"], program.Operation_ACCESS)
		assertRefName(t, accOps[1], "", "x")
		target, wrt := derivOf(t, accOps[0])
		if wrt != program.PVarName {
			t.Errorf("j wrt = %q, want %q", wrt, program.PVarName)
		}
		if call, ok := target.T.(*program.Operand_FnCall); !ok || call.FnCall.Reference.Name != "f" {
			t.Errorf("j target = %s, want call f", target.PrettyPrint(0))
		}
	}

	// k = v.d(t)[0] -> INDEX(DERIV(v, t), 0)
	{
		idxOps := exprOp(t, byName["k"], program.Operation_INDEX)
		assertVal(t, idxOps[1], 0)
		target, wrt := derivOf(t, idxOps[0])
		if wrt != "t" {
			t.Errorf("k wrt = %q, want t", wrt)
		}
		assertRefName(t, target, "", "v")
	}

	// l = pkg.m[0][1] -> INDEX(INDEX(pkg.m, 0), 1)
	{
		outerOps := exprOp(t, byName["l"], program.Operation_INDEX)
		assertVal(t, outerOps[1], 1)
		innerOps := exprOp(t, outerOps[0], program.Operation_INDEX)
		assertRefName(t, innerOps[0], "pkg", "m")
		assertVal(t, innerOps[1], 0)
	}

	// n = pkg.body.box[0] -> INDEX(BOUNDINGBOX(pkg.body), 0)
	{
		idxOps := exprOp(t, byName["n"], program.Operation_INDEX)
		assertVal(t, idxOps[1], 0)
		boxOps := exprOp(t, idxOps[0], program.Operation_BOUNDINGBOX)
		assertRefName(t, boxOps[0], "pkg", "body")
	}

	// q = pkg.v.d(t)[1] -> INDEX(DERIV(pkg.v, t), 1)
	{
		idxOps := exprOp(t, byName["q"], program.Operation_INDEX)
		assertVal(t, idxOps[1], 1)
		target, wrt := derivOf(t, idxOps[0])
		if wrt != "t" {
			t.Errorf("q wrt = %q, want t", wrt)
		}
		assertRefName(t, target, "pkg", "v")
	}

	// s = ["abc"[0]] -> COLLECT(INDEX(COLLECT(97, 98, 99), 0)), not bare string expansion.
	{
		colOps := exprOp(t, byName["s"], program.Operation_COLLECT)
		if len(colOps) != 1 {
			t.Fatalf("s collection len = %d, want 1", len(colOps))
		}
		idxOps := exprOp(t, colOps[0], program.Operation_INDEX)
		assertVal(t, idxOps[1], 0)
		strOps := exprOp(t, idxOps[0], program.Operation_COLLECT)
		if len(strOps) != 3 {
			t.Fatalf("string runes len = %d, want 3", len(strOps))
		}
	}
}

func TestPostfixModifierErrors(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "empty_single_index",
			src:  "a = [1, 2];\nx = a[];\nreturn x;\n",
			want: "index expression cannot be empty",
		},
		{
			name: "empty_chained_index",
			src:  "m = [[1, 2]];\nx = m[0][];\nreturn x;\n",
			want: "index expression cannot be empty",
		},
		{
			name: "unclosed_chained_index_short",
			src:  "m = [[1, 2]];\nx = m[0][1;\nreturn x;\n",
			want: `unexpected token "["`,
		},
		{
			name: "unclosed_chained_index_expr",
			src:  "m = [[1, 2]];\nx = m[0][1 + 2 + 3;\nreturn x;\n",
			want: `expected "]"`,
		},
		{
			name: "trailing_dot_after_index",
			src:  "pts = [[1, 2, 3]];\nx = pts[0].;\nreturn x;\n",
			want: `unexpected token "."`,
		},
		{
			name: "invalid_accessor_after_index",
			src:  "pts = [[1, 2, 3]];\nx = pts[0].foo;\nreturn x;\n",
			want: "invalid accessor: foo",
		},
		{
			name: "overlong_swizzle_after_index",
			src:  "pts = [[1, 2, 3]];\nx = pts[0].xxxxx;\nreturn x;\n",
			want: "accessor must be 1 to 4 characters: xxxxx",
		},
		{
			name: "args_on_swizzle_in_chain",
			src:  "pts = [[1, 2, 3]];\nx = pts[0].x(1);\nreturn x;\n",
			want: "accessor .x takes no arguments",
		},
		{
			name: "args_on_box_in_chain",
			src:  "b = sphere(1);\nx = (b).box(1)[0];\nreturn x;\n",
			want: "accessor .box takes no arguments",
		},
		{
			name: "deriv_missing_args_in_chain",
			src:  "t = 1;\npts = [[t, 2]];\nx = pts[0].d;\nreturn x;\n",
			want: ".d takes one argument",
		},
		{
			name: "deriv_non_var_in_chain",
			src:  "t = 1;\npts = [[t, 2]];\nx = pts[0].d(1);\nreturn x;\n",
			want: ".d takes the name of a variable",
		},
		{
			name: "chained_index_in_return_missing_expr",
			src:  "m = [[1, 2]];\nreturn m[0][;\n",
			want: `unexpected token ";"`,
		},
		{
			name: "parameter_direction_anchor_rejects_indexed_ref",
			src:  "f(p0, @direction(p0[0]) d) { return p0; }\nreturn 1;\n",
			want: "@direction's anchor on a parameter must name another parameter",
		},
		{
			name: "invalid_expr_inside_index_modifier",
			src:  "m = [[1, 2]];\npts = [[1, 2, 3]];\nx = m[0][pts[0].foo];\nreturn x;\n",
			want: "invalid accessor: foo",
		},
		{
			name: "vdb_sdf_divisor_rejects_modified_int_literal",
			src:  "x = vdbSdf(0, 256[0]);\nreturn x;\n",
			want: "vdbSdf() second argument (voxelDivisor) must be a positive integer literal",
		},
		{
			name: "vdb_sdf_half_band_rejects_modified_float_literal",
			src:  "x = vdbSdf(0, 256, 3.0[0]);\nreturn x;\n",
			want: "vdbSdf() third argument (halfBandVoxels) must be a positive numeric literal",
		},
		{
			name: "u_then_second_rejected",
			src:  "s = sphere(1);\nreturn s.u.second;\n",
			want: "cannot apply .second after .u",
		},
		{
			name: "v_then_index_rejected",
			src:  "s = sphere(1);\nreturn s.v[0];\n",
			want: "cannot index after .v",
		},
		{
			name: "second_then_swizzle_rejected",
			src:  "s = sphere(1);\nreturn s.second.x;\n",
			want: "cannot apply .x after .second",
		},
		{
			name: "second_then_box_rejected",
			src:  "s = sphere(1);\nreturn s.second.box;\n",
			want: "cannot apply .box after .second",
		},
		{
			name: "box_then_swizzle_rejected",
			src:  "s = sphere(1);\nreturn s.box.x;\n",
			want: "cannot apply .x after .box",
		},
		{
			name: "box_then_box_rejected",
			src:  "s = sphere(1);\nreturn s.box.box;\n",
			want: "cannot apply .box after .box",
		},
		{
			name: "deriv_then_second_rejected",
			src:  "s = sphere(1);\nreturn s.dp.second;\n",
			want: "cannot apply .second after .dp",
		},
		{
			name: "deriv_then_box_rejected",
			src:  "s = sphere(1);\nreturn s.dtime.box;\n",
			want: "cannot apply .box after .dtime",
		},
		{
			name: "deriv_wrt_then_box_rejected",
			src:  "t = 1;\nv = [t, 2];\nreturn v.d(t).box;\n",
			want: "cannot apply .box after .d(...)",
		},
		{
			name: "swizzle_then_u_rejected",
			src:  "v = [1, 2, 3];\nreturn v.xy.u;\n",
			want: "cannot apply .u after .xy",
		},
		{
			name: "scalar_swizzle_then_index_rejected",
			src:  "v = [1, 2, 3];\nreturn v.x[0];\n",
			want: "cannot index after .x",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, diags, err := parser2.Parse([]string{tc.src})
			if err == nil {
				t.Fatalf("expected parse error containing %q, got nil", tc.want)
			}
			if len(diags) != 0 {
				t.Fatalf("unexpected diagnostics: %+v", diags)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.want)
			}
		})
	}
}

func TestPostfixModifierFieldPropertiesAndAnnotationGuards(t *testing.T) {
	src := strings.Join([]string{
		`shapes = [sphere(1), box([1, 1, 1])];`,
		`sec = shapes[0].second;`,
		`secu = shapes[0].second.u.dp;`,
		`bx = shapes[0].box[0].x;`,
		`su = shapes[0].u;`,
		`sv = shapes[1].v;`,
		`@spline sp1 = [[[0, 0, 0], [1, 1, 1]]][0];`,
		`@spline sp2 = [uiTransform(shapes)[0]];`,
		`return sec + secu.x + bx + su + sv;`,
	}, "\n") + "\n"

	res, err := lib.ParsePackage([]string{src})
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	byName := make(map[string]*program.Operand)
	for _, assign := range res.Package.Files[0].Assignments {
		byName[assign.Variables[0].Name] = assign.Expression
	}

	secOps := exprOp(t, byName["sec"], program.Operation_SECONDCLOSEST)
	exprOp(t, secOps[0], program.Operation_INDEX)

	suOps := exprOp(t, byName["su"], program.Operation_SURFACEU)
	exprOp(t, suOps[0], program.Operation_INDEX)

	svOps := exprOp(t, byName["sv"], program.Operation_SURFACEV)
	exprOp(t, svOps[0], program.Operation_INDEX)

	// sp1 has an outer `[0]` modifier, so its inner collection elements are not promoted to uiTransform.
	sp1Call, ok := byName["sp1"].T.(*program.Operand_FnCall)
	if !ok || sp1Call.FnCall.Reference.Name != "uiSpline" {
		t.Fatalf("sp1 = %s, want uiSpline call", byName["sp1"].PrettyPrint(0))
	}
	sp1Idx := exprOp(t, sp1Call.FnCall.Arguments[0], program.Operation_INDEX)
	sp1OuterCol := exprOp(t, sp1Idx[0], program.Operation_COLLECT)
	exprOp(t, sp1OuterCol[0], program.Operation_COLLECT)

	// sp2's element is `uiTransform(shapes)[0]`, which has a modifier so it gets wrapped in uiTransform.
	sp2Call, ok := byName["sp2"].T.(*program.Operand_FnCall)
	if !ok || sp2Call.FnCall.Reference.Name != "uiSpline" {
		t.Fatalf("sp2 = %s, want uiSpline call", byName["sp2"].PrettyPrint(0))
	}
	sp2Col := exprOp(t, sp2Call.FnCall.Arguments[0], program.Operation_COLLECT)
	elemCall, ok := sp2Col[0].T.(*program.Operand_FnCall)
	if !ok || elemCall.FnCall.Reference.Name != "uiTransform" {
		t.Fatalf("sp2[0] = %s, want wrapped uiTransform call", sp2Col[0].PrettyPrint(0))
	}
}
