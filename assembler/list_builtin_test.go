package assembler_test

import (
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/assembler"
)

// foldedReturn assembles src and returns its first return, folded.
func foldedReturn(t *testing.T, src string) string {
	t.Helper()
	res, err := assembler.Assemble([]string{src})
	if err != nil {
		t.Fatalf("%v\n%s", err, src)
	}
	text := res.Program.PrettyPrint(0)
	i := strings.Index(text, "return ")
	if i < 0 {
		t.Fatalf("no return in:\n%s", text)
	}
	line, _, _ := strings.Cut(text[i:], "\n")
	return line
}

// list(n) is the indices, list(n, fn) maps fn over them, and n may reach a
// function's body as a constant parameter.
func TestListBuiltin(t *testing.T) {
	for _, c := range []struct{ src, want string }{
		{"v = list(4);\nreturn v[3];", "return 3;"},
		{"k = 2 * 3;\nreturn list(k)[5];", "return 5;"},
		{"return list(3, (i) { return i * 2; })[2];", "return 4;"},
		{"f(n) { return list(n, (i) { return i + 1; }); }\nreturn f(5)[4] + f(2)[1];", "return 7;"},
		{"f(n) { return list(n, (i) { return i + 1; }); }\ng(k) { return f(k + 1); }\nreturn g(3)[3];", "return 4;"},
		{"ident(n) { return list(n, (j) { return list(n, (i) { return 1 - min(abs(i - j), 1); }); }); }\nreturn ident(3)[1][1] + ident(4)[0][3] + ident(4)[2][2];", "return 2;"},
	} {
		if got := foldedReturn(t, c.src); got != c.want {
			t.Errorf("%s\n  = %s, want %s", c.src, got, c.want)
		}
	}
}

func TestListBuiltinNeedsAConstantLength(t *testing.T) {
	for _, c := range []struct{ src, want string }{
		{"f(n) { return list(n); }\nreturn f(p.x)[0];", "f needs a constant for its parameter n, which decides the length of a list"},
		{"return list(p.x)[0];", "list's length must be known when the document compiles"},
		{"return list(2.5)[0];", "list's length must be a whole number of at least 1"},
	} {
		if msg := assembleErr(t, c.src); !strings.Contains(msg, c.want) {
			t.Errorf("missing %q in:\n%s", c.want, msg)
		}
	}
}
