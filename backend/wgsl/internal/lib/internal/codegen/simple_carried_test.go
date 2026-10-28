package codegen_test

import (
	"strings"
	"testing"

	assembler "gradientcontrol.com/omega/core/internal/assembler"
)

// A part's stand-in declared inside the function that builds the part comes
// back with it (property/carry.go), placed as the part is: the viewer switches
// to it inside the part's guard, and its colour joins the part's.

// carriedWidget builds a box whose stand-in is a striped red sphere. The
// stripes keep the colour from being a constant, which the fold would take
// for the part's (as it does a top-level stand-in's).
const carriedWidget = `
widget(r) {
  s = box([-r, -r, -r], [r, r, r]);
  s.box = [[-r, -r, -r], [r, r, r]];
  stand = sphere(r);
  stand.color = [1, 0.5 + 0.5 * sin(10 * p.x), 0];
  s.simple = stand;
  return s;
}
`

// renderCarried assembles and renders src, failing the test on an error.
func renderCarried(t *testing.T, src string) string {
	t.Helper()
	res, err := assembler.Assemble([]string{src}, mockMgr)
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	out, err := renderHelper(src, res.Program)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	return out.WGSL
}

// fnBody is the body of the WGSL function name, or "" when there is none.
func fnBody(wgsl, name string) string {
	at := strings.Index(wgsl, "fn "+name+"(")
	if at < 0 {
		return ""
	}
	body := wgsl[at:]
	return body[:strings.Index(body, "\n}\n")]
}

func TestSimpleCarriedOutOfFunction(t *testing.T) {
	for name, decl := range map[string]string{
		"bound to the call": "q = widget(1.0);",
		"placed":            "q = transform.translate(widget(1.0), [3, 0, 0]);",
	} {
		wgsl := renderCarried(t, carriedWidget+decl+"\nreturn q;")
		if n := strings.Count(fnBody(wgsl, "sdShape"), "gsl_simple("); n != 1 {
			t.Errorf("%s: %d switches in sdShape, want the part's one", name, n)
		}
		if !strings.Contains(fnBody(wgsl, "gsl_colorField"), "gsl_simple(") {
			t.Errorf("%s: the colour field does not blend to the stand-in's red", name)
		}
	}
}

// A colour the document sets on the part is its stand-in's too, so there is
// no stand-in colour to blend to.
func TestSimpleCarriedTakesDocumentColour(t *testing.T) {
	wgsl := renderCarried(t, carriedWidget+`
q = transform.translate(widget(1.0), [3, 0, 0]);
q.color = [0, 0, 1];
return q;`)
	if n := strings.Count(fnBody(wgsl, "sdShape"), "gsl_simple("); n != 1 {
		t.Errorf("%d switches in sdShape, want the part's one", n)
	}
	if strings.Contains(fnBody(wgsl, "gsl_colorField"), "gsl_simple(") {
		t.Errorf("the colour field blends to the stand-in's own colour, want the document's blue on both")
	}
}

// A value read through a remap is not a part, so it has no stand-in to show,
// and one the document did not write is no error.
func TestSimpleCarriedDroppedOffNonPart(t *testing.T) {
	wgsl := renderCarried(t, carriedWidget+`
q = remap(widget(1.0), p - [3, 0, 0]);
return q;`)
	if strings.Contains(wgsl, "gsl_simple(") {
		t.Errorf("a stand-in was drawn for a value that is not a part")
	}
}

// A shape's stand-in on a local that sets nothing only a shape takes cannot
// be told from a colour's before types are known, so it does not leave the
// function, and is reported rather than lowered as a colour's.
func TestSimpleInFunctionNeedsShape(t *testing.T) {
	src := `
widget(r) {
  s = box([-r, -r, -r], [r, r, r]);
  s.simple = sphere(r);
  return s;
}
q = widget(1.0);
return q;`
	_, err := assembler.Assemble([]string{src}, mockMgr)
	if err == nil || !strings.Contains(err.Error(), "s.simple: a shape's stand-in leaves a function only") {
		t.Fatalf("got %v, want the stand-in reported", err)
	}
}

// Each way the fastener library draws a thread steps down one way of drawing
// on the screw's stand-in (mechanical/fastener, Stand-ins): geometry and
// relief are painted on it, paint gives way to plain steel, and a plain screw
// has a plain stand-in. The colour field draws the stand-in's colour first and
// blends to the part's inside the branch the switch's weight opens. The screws
// stand at the origin, so their colours are written in the field itself
// rather than in the remap a placement reads them through.
func TestFastenerStandIns(t *testing.T) {
	const branch = "if (gsl_simpleWeight > 0.0)"
	for _, c := range []struct {
		name, expr      string
		standPaint      bool // the stand-in's colour is thread.paint
		partColor       string
		standColourRung bool // thread.paint goes flat on the Textures rung
	}{
		{"plain", "fastener.shcs(5.0, 16.0)", false, "", false},
		{"geometry", "fastener.shcsThreaded(5.0, 16.0)", true, "", true},
		{"relief", "fastener.shcs(5.0, 16.0, steelMat, fastener.threadIso)", true, "thread_reliefColor", true},
		{"paint", "fastener.shcs(5.0, 16.0, steelMat, fastener.threadPaint)", false, "thread_paint", false},
		{"nut geometry", "fastener.nutThreaded(5.0)", true, "", true},
		{"nut relief", "fastener.nut(5.0, steelMat, fastener.threadSin)", true, "thread_reliefColor", true},
	} {
		wgsl := renderCarried(t, "steelMat, steelColor = metal.steel();\nq = "+c.expr+";\nreturn q;")
		if n := strings.Count(fnBody(wgsl, "sdShape"), "gsl_simple("); n != 1 {
			t.Errorf("%s: %d switches in sdShape, want the screw's one", c.name, n)
		}
		color := fnBody(wgsl, "gsl_colorField")
		if c.name == "plain" {
			if color != "" {
				t.Errorf("plain: a colour field, want none")
			}
			continue
		}
		at := strings.Index(color, branch)
		if at < 0 {
			t.Errorf("%s: the colour field does not switch to the stand-in's:\n%s", c.name, color)
			continue
		}
		stand, part := color[:at], color[at:]
		if got := strings.Contains(stand, "thread_paint("); got != c.standPaint {
			t.Errorf("%s: stand-in painted = %v, want %v:\n%s", c.name, got, c.standPaint, stand)
		}
		if c.partColor != "" && !strings.Contains(part, c.partColor+"(") {
			t.Errorf("%s: the part's colour does not read %s:\n%s", c.name, c.partColor, part)
		}
		if strings.Contains(stand, "thread_reliefColor(") {
			t.Errorf("%s: the stand-in paints with reliefColor, which fades where the part is displaced", c.name)
		}
		if c.standColourRung && !strings.Contains(fnBody(wgsl, "thread_paint"), "gsl_simpleDetail") {
			t.Errorf("%s: the stand-in's paint does not go flat on the Textures rung", c.name)
		}
	}
}
