package assembler_test

import (
	"fmt"
	"strings"
	"testing"
)

// The size-generic matrix operations match the sized closed forms, and the
// pivoting handles a zero on the diagonal.
func TestMatrixGenericMatchesSized(t *testing.T) {
	m2 := "[[2, 1], [0.3, 3]]"
	m3 := "[[2, 1, 0.5], [0.3, 3, 1], [1, 0.2, 4]]"
	m4 := "[[2, 1, 0.5, 0], [0.3, 3, 1, 0.1], [1, 0.2, 4, 0.7], [0.5, 0, 0.3, 5]]"
	quarter := "[[0, 1, 0], [-1, 0, 0], [0, 0, 1]]"
	affine := "matrix.compose([1, 2, 3], [0.3, 0.2, 0.1])"
	diff := func(a, b string, n int) string {
		var terms []string
		for c := 0; c < n; c++ {
			for r := 0; r < n; r++ {
				terms = append(terms, fmt.Sprintf("abs(%s[%d][%d] - %s[%d][%d])", a, c, r, b, c, r))
			}
		}
		return strings.Join(terms, " + ")
	}
	pt := [][3]float64{{0, 0, 0}}
	for _, c := range []struct{ a, b string }{
		{diff("matrix.transpose("+m2+")", "matrix.transpose2("+m2+")", 2), "0"},
		{diff("matrix.transpose("+m3+")", "matrix.transpose3("+m3+")", 3), "0"},
		{diff("matrix.transpose("+m4+")", "matrix.transpose4("+m4+")", 4), "0"},
		{diff("matrix.identity(3)", "matrix.identity3()", 3), "0"},
		{diff("matrix.identity(4)", "matrix.identity4()", 4), "0"},
		{"matrix.determinant(" + m2 + ")", "matrix.determinant2(" + m2 + ")"},
		{"matrix.determinant(" + m3 + ")", "matrix.determinant3(" + m3 + ")"},
		{"matrix.determinant(" + m4 + ")", "matrix.determinant4(" + m4 + ")"},
		{"matrix.determinant(" + quarter + ")", "1"},
		{"matrix.determinant([[0, 0, 1], [0, 1, 0], [1, 0, 0]])", "-1"},
		{diff("matrix.inverse("+m2+")", "matrix.inverse2("+m2+")", 2), "0"},
		{diff("matrix.inverse("+m3+")", "matrix.inverse3("+m3+")", 3), "0"},
		{diff("matrix.inverse("+m4+")", "matrix.inverse4("+m4+")", 4), "0"},
		{diff("matrix.inverse("+quarter+")", "matrix.transpose3("+quarter+")", 3), "0"},
		{diff("matrix.inverse("+affine+")", "matrix.inverseAffine("+affine+")", 4), "0"},
	} {
		if d := differenceAt(t, c.a, c.b, pt); d > 1e-9 {
			t.Errorf("%.80s... differs from %s by %g", c.a, c.b, d)
		}
	}
}

// A matrix that varies with p takes the specialized, unrolled path; evaluated
// at sample points it matches the closed forms, a zero pivot included.
func TestMatrixGenericVaryingMatchesSized(t *testing.T) {
	m := "[[p.x * 0.1, 1 + p.y * 0.05, 0.2], [1, p.x * 0.02, 0.3], [0.1, 0.4, 1 + p.z * 0.03]]"
	if d := differenceAt(t, "matrix.determinant("+m+")", "matrix.determinant3("+m+")", samplePoints); d > 1e-6 {
		t.Errorf("determinant differs by %g", d)
	}
	for _, e := range []string{"[0][0]", "[1][2]", "[2][1]"} {
		if d := differenceAt(t, "matrix.inverse("+m+")"+e, "matrix.inverse3("+m+")"+e, samplePoints); d > 1e-6 {
			t.Errorf("inverse%s differs by %g", e, d)
		}
	}
}
