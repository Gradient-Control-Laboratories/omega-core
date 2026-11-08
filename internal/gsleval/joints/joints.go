// Package joints finds the joints a GSL document makes and proves each one.
//
// A joint (mechanical/joint) builds the matching features on every part it
// joins from one frame, so its fit holds by construction. What a design gets
// wrong is the CONTEXT: a grip that is not the parts' grip, a thread with no
// wall around it, a screw that runs into a part the joint was not told about.
// This package reads the document for joint calls and writes, for each, the
// claims that say it is right, as ordinary certify expressions in the
// document's scope, over a region that is the joint's own extent. Proving them
// is the caller's: the gsl CLI, the verifyJoints chat tool and the final
// re-check of a returned program all run the same claims.
//
// Discovery is textual and covers top-level declarations of the form
//
//	screw, clamped, threaded = joint.bolted(clampedBody, threadedBody, seat, shankDir, d, grip, fit);
//
// and the same for joint.countersunk, which takes the same arguments, and for
// a fastener inserted into a declared interface,
//
//	screw, clamped, threaded = joint.insert(iface, head, clampedBody, threadedBody, grip);
//
// which is proven as the joint its head makes it, with the interface's
// accessors standing for the seat, direction, size and fit.
//
// A joint call anywhere else — inside a function body, or not destructured
// into its three parts — is reported as skipped rather than silently passed.
package joints

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

// Joint is one joint call found in a document.
type Joint struct {
	Kind  string   `json:"kind"`
	Names []string `json:"names"`
	Args  []string `json:"args"`
	Line  int      `json:"line"`
	// Insert is the joint.insert call this joint was resolved from, as
	// written; empty for a joint.bolted or joint.countersunk call.
	Insert string `json:"insert,omitempty"`
}

// Call is the joint's call as written, normalized to one line.
func (j Joint) Call() string {
	if j.Insert != "" {
		return j.Insert
	}
	return strings.Join(j.Names, ", ") + " = joint." + j.Kind + "(" + strings.Join(j.Args, ", ") + ")"
}

// Claim is one thing a joint has to be true: Expr >= 0 over the joint's region.
// Hint is the repair a failure of it usually calls for.
type Claim struct {
	Name string `json:"name"`
	Expr string `json:"expr"`
	Hint string `json:"hint,omitempty"`
}

// boltedArgs is joint.bolted's parameter list, in order, and
// joint.countersunk's.
var boltedArgs = []string{"clamped", "threaded", "seatPoint", "shankDir", "nominalDia", "grip", "fit"}

// insertArgs is joint.insert's parameter list, in order.
var insertArgs = []string{"iface", "head", "clamped", "threaded", "grip"}

// heads are joint.insert's head constants and the joint each makes.
var heads = map[float64]string{1: "bolted", 2: "countersunk"}

// kindHelpers are the joint library's names for one kind of joint: its
// length, the head's diameter, the seat and engagement allowances, its grip
// zone and that zone's claim and repair hint. Every kind shares
// boltedThreadZone.
type kindHelpers struct {
	length, headDia, seat, engaged, gripZone string
	// seatTakesFit: the seat allowance takes the fit (bolted's counterbore
	// is over the clearance; a countersink's cone is not).
	seatTakesFit       bool
	gripName, gripHint string
}

var kinds = map[string]kindHelpers{
	"bolted": {"joint.boltedLength", "catalog.counterbore", "joint.boltedSeat", "joint.boltedEngaged", "joint.boltedGripZone", true,
		"the clamped part fills the grip under the head",
		"the clamped part is not solid under the head for the whole grip: the grip is longer than " +
			"the part's thickness along the shank at the seat, the seat is not on the part, or the " +
			"head overhangs its edge"},
	"countersunk": {"joint.countersunkLength", "catalog.cskHeadDia", "joint.countersunkSeat", "joint.countersunkEngaged", "joint.countersunkGripZone", false,
		"the clamped part holds the head all round its cone",
		"the clamped part is not solid round the countersink: the countersink breaks out of an " +
			"edge, or the seat is not on the clamped part's outer face"},
}

// Find returns the joint calls a document declares at top level, the
// expressions its top-level return lists, and a note for each joint call it
// cannot verify.
func Find(src string) (found []Joint, returns []string, skipped []string) {
	text := stripComments(src)
	for _, st := range statements(text) {
		body := strings.TrimSpace(st.text)
		if st.inBody {
			if jointCall.MatchString(body) {
				skipped = append(skipped, fmt.Sprintf("line %d: a joint call inside a function body is not verified; declare the joint at top level", st.line))
			}
			continue
		}
		if m := returnStmt.FindStringSubmatch(body); m != nil {
			returns = splitTop(m[1])
			continue
		}
		loc := jointCall.FindStringSubmatchIndex(body)
		if loc == nil {
			continue
		}
		kind := body[loc[2]:loc[3]]
		m := destructure.FindStringSubmatch(body)
		if m == nil {
			skipped = append(skipped, fmt.Sprintf("line %d: joint.%s is not destructured as `screw, clamped, threaded = joint.%s(...)`; not verified", st.line, kind, kind))
			continue
		}
		open := loc[1] - 1 // the call's opening parenthesis
		end := matchParen(body, open)
		if end < 0 {
			skipped = append(skipped, fmt.Sprintf("line %d: unbalanced joint.%s call; not verified", st.line, kind))
			continue
		}
		args := splitTop(body[open+1 : end])
		params := boltedArgs
		if kind == "insert" {
			params = insertArgs
		}
		if len(args) != len(params) {
			skipped = append(skipped, fmt.Sprintf("line %d: joint.%s takes %d arguments (%s), got %d; not verified",
				st.line, kind, len(params), strings.Join(params, ", "), len(args)))
			continue
		}
		found = append(found, Joint{Kind: kind, Names: []string{m[1], m[2], m[3]}, Args: args, Line: st.line})
	}
	return found, returns, skipped
}

var (
	// jointCall matches a call that BUILDS a joint — joint.boltedLength and
	// the claim zones are helpers, not joints.
	jointCall   = regexp.MustCompile(`\bjoint\s*\.\s*(bolted|countersunk|insert)\s*\(`)
	destructure = regexp.MustCompile(`^([A-Za-z_]\w*)\s*,\s*([A-Za-z_]\w*)\s*,\s*([A-Za-z_]\w*)\s*=\s*joint\s*\.\s*(?:bolted|countersunk|insert)\s*\(`)
	returnStmt  = regexp.MustCompile(`(?s)^return\b(.*)$`)
)

// Others are the returned parts the screw is held clear of: each expression of
// the top-level return, a min(...) of parts split into its parts, less the
// joint's own three names and anything that names the screw. Run then drops any
// part that is solid on the screw's own axis (see Run).
func Others(j Joint, returns []string) []string {
	screw, held, tapped := j.Names[0], j.Names[1], j.Names[2]
	own := regexp.MustCompile(`\b` + regexp.QuoteMeta(screw) + `\b`)
	var out []string
	for _, r := range splitMins(returns) {
		if r == screw || r == held || r == tapped || own.MatchString(r) {
			continue
		}
		out = append(out, r)
	}
	return out
}

// splitMins replaces each min(a, b, ...) among exprs by its arguments,
// recursively: a display union is the parts it unions.
func splitMins(exprs []string) []string {
	var out []string
	for _, e := range exprs {
		e = strings.TrimSpace(e)
		if m := minCall.FindStringIndex(e); m != nil && m[0] == 0 {
			if end := matchParen(e, m[1]-1); end == len(e)-1 {
				out = append(out, splitMins(splitTop(e[m[1]:end]))...)
				continue
			}
		}
		out = append(out, e)
	}
	return out
}

var minCall = regexp.MustCompile(`^min\s*\(`)

// Claims are the joint's claims: its own fit (the screw clears the clamped
// part and overlaps the threaded one only in the engagement), its context (the
// clamped part fills the grip under the head, the threaded part holds 1.5 d of
// thread with a wall round it), and the screw against others, the rest of the
// assembly.
func Claims(j Joint, others []string) []Claim {
	a := make([]string, len(j.Args))
	for i, s := range j.Args {
		a[i] = "(" + s + ")"
	}
	clamped, threaded, seat, dir, d, grip, fit := a[0], a[1], a[2], a[3], a[4], a[5], a[6]
	screw, held, tapped := j.Names[0], j.Names[1], j.Names[2]
	k := kinds[j.Kind]
	seatOK := fmt.Sprintf("%s(%s, %s, %s)", k.seat, seat, dir, d)
	if k.seatTakesFit {
		seatOK = fmt.Sprintf("%s(%s, %s, %s, %s)", k.seat, seat, dir, d, fit)
	}
	engaged := fmt.Sprintf("%s(%s, %s, %s, %s)", k.engaged, seat, dir, d, grip)
	claims := []Claim{
		{"screw clears the clamped part", fmt.Sprintf("max(%s, %s, %s)", screw, held, seatOK),
			"the clamped part has material the joint's counterbore and clearance do not reach — " +
				"is the clamped body passed to the joint the one returned, and is the seat inside it?"},
		{"screw overlaps the threaded part only in its thread", fmt.Sprintf("max(%s, %s, %s)", screw, tapped, engaged),
			"the threaded part has material in the screw's way outside the thread — above the grip, " +
				"round the head — so it overlaps the clamped part's place; check the grip and the bodies"},
		{k.gripName, fmt.Sprintf("max(%s(%s, %s, %s, %s, %s), neg(%s))", k.gripZone, seat, dir, d, grip, fit, clamped),
			k.gripHint},
		{"the threaded part holds 1.5 d of thread with a wall round it", fmt.Sprintf("max(joint.boltedThreadZone(%s, %s, %s, %s), neg(%s))", seat, dir, d, grip, threaded),
			"the threaded part does not surround the first 1.5 d of thread past the grip with a " +
				"quarter-diameter wall: the grip is shorter than the clamped part's thickness, the " +
				"threaded part is too thin there, or the hole is too near its edge"},
	}
	if len(others) > 0 {
		rest := others[0]
		if len(others) > 1 {
			rest = "min(" + strings.Join(others, ", ") + ")"
		}
		claims = append(claims, Claim{"screw clears the rest of the assembly",
			fmt.Sprintf("max(%s, %s, %s, %s)", screw, rest, seatOK, engaged),
			"the screw runs into another returned part: move the joint, or the part, clear of it"})
	}
	return claims
}

// RegionExprs are the expressions whose values Region needs, in the order it
// takes them, and then the grip: the seat (vec3), the shank direction (vec3),
// the size, the screw's length, the head's pocket diameter (the counterbore,
// or a countersunk head's rim) and the grip.
func RegionExprs(j Joint) []string {
	seat, dir, d, grip := j.Args[2], j.Args[3], j.Args[4], j.Args[5]
	return []string{
		seat, dir, d,
		fmt.Sprintf("%s(%s, %s)", kinds[j.Kind].length, d, grip),
		fmt.Sprintf("%s(%s)", kinds[j.Kind].headDia, d),
		grip,
	}
}

// Region is the joint's extent, padded by 1: a box round the axis from the
// head's top (one diameter above the seat) to a diameter past the screw's
// tip, as wide as the counterbore or the thread's wall, whichever is wider.
func Region(seat, dir [3]float64, d, length, counterbore float64) ([6]float64, error) {
	n := math.Sqrt(dir[0]*dir[0] + dir[1]*dir[1] + dir[2]*dir[2])
	if !(n > 0) || !(d > 0) || !(length > 0) {
		return [6]float64{}, fmt.Errorf("the joint does not build: size %g, length %g, direction %v (is the size a catalog size, and the grip short enough for a catalog screw?)", d, length, dir)
	}
	radius := math.Max(counterbore/2, 0.75*d) + 1
	var box [6]float64
	for i := 0; i < 3; i++ {
		u := dir[i] / n
		head := seat[i] - u*(d+1)
		tip := seat[i] + u*(length+d+1)
		box[i] = math.Min(head, tip) - radius
		box[i+3] = math.Max(head, tip) + radius
	}
	return box, nil
}

// Verdict is one certify outcome.
type Verdict struct {
	Proven        bool       `json:"proven"`
	Disproven     bool       `json:"disproven,omitempty"`
	Indeterminate bool       `json:"indeterminate,omitempty"`
	Reason        string     `json:"reason,omitempty"`
	CounterBox    [6]float64 `json:"counterexampleBox,omitempty"`
	Evals         int        `json:"evals,omitempty"`
}

// Prover is what Run needs from an evaluator: the values of expressions at a
// point, in the document's scope — every component, for a vector — and a
// proof that an expression is never negative over a region.
type Prover struct {
	Eval    func(exprs []string, at [3]float64) ([][]float64, error)
	Certify func(expr string, region [6]float64) (Verdict, error)
}

// ClaimResult is one claim and how it came out.
type ClaimResult struct {
	Claim
	Verdict
	Error string `json:"error,omitempty"`
}

// JointReport is one joint's claims and their outcomes.
type JointReport struct {
	Joint
	Region *[6]float64   `json:"region,omitempty"`
	Claims []ClaimResult `json:"claims"`
	// Unchecked names the returned parts the screw is NOT proven clear of,
	// and why: each was solid on the screw's own axis inside the grip, where
	// only the screw can be, so it contains the screw — a display union —
	// or fills its clearance hole.
	Unchecked []string `json:"unchecked,omitempty"`
	Error     string   `json:"error,omitempty"`
}

// Report is every joint in a document, proven.
type Report struct {
	Joints  []JointReport `json:"joints"`
	Skipped []string      `json:"skipped,omitempty"`
	// AllProven is true when there is at least one joint and every claim of
	// every joint is proven.
	AllProven bool `json:"allProven"`
}

// Run finds src's joints and proves each one's claims with pv.
func Run(src string, pv Prover) Report {
	found, returns, skipped := Find(src)
	rep := Report{Skipped: skipped, AllProven: len(found) > 0}
	for _, j := range found {
		jr := JointReport{Joint: j}
		if j.Kind == "insert" {
			resolved, err := resolveInsert(j, pv)
			if err != nil {
				jr.Error = err.Error()
				rep.AllProven = false
				rep.Joints = append(rep.Joints, jr)
				continue
			}
			j = resolved
			jr.Joint = j
		}
		region, mid, err := jointRegion(j, pv)
		if err != nil {
			jr.Error = err.Error()
			rep.AllProven = false
			rep.Joints = append(rep.Joints, jr)
			continue
		}
		jr.Region = &region
		var others []string
		others, jr.Unchecked = clearOfAxis(Others(j, returns), mid, pv)
		for _, c := range Claims(j, others) {
			cr := ClaimResult{Claim: c}
			v, err := pv.Certify(c.Expr, region)
			if err != nil {
				cr.Error = err.Error()
			} else {
				cr.Verdict = v
			}
			if !cr.Proven {
				rep.AllProven = false
			}
			jr.Claims = append(jr.Claims, cr)
		}
		rep.Joints = append(rep.Joints, jr)
	}
	return rep
}

// resolveInsert is a joint.insert call as the joint its head makes: the head
// is evaluated, and the interface's accessors stand in for the seat, the
// direction, the size and the fit, so the claims are that joint's own.
func resolveInsert(j Joint, pv Prover) (Joint, error) {
	iface, head, clamped, threaded, grip := j.Args[0], j.Args[1], j.Args[2], j.Args[3], j.Args[4]
	vals, err := pv.Eval([]string{head}, [3]float64{})
	if err != nil {
		return j, err
	}
	if len(vals) != 1 || len(vals[0]) != 1 {
		return j, fmt.Errorf("the head %q is not one constant (joint.socketHead or joint.flatHead)", head)
	}
	kind, ok := heads[vals[0][0]]
	if !ok {
		return j, fmt.Errorf("the head %q is %g, not joint.socketHead or joint.flatHead", head, vals[0][0])
	}
	acc := func(name string) string { return "joint." + name + "(" + iface + ")" }
	return Joint{
		Kind:   kind,
		Names:  j.Names,
		Args:   []string{clamped, threaded, acc("seat"), acc("direction"), acc("diameter"), grip, acc("fit")},
		Line:   j.Line,
		Insert: strings.Join(j.Names, ", ") + " = joint.insert(" + strings.Join(j.Args, ", ") + ")",
	}, nil
}

// jointRegion evaluates the joint's placement and sizes, boxes it, and
// returns the point on its axis where only the screw can be: halfway through
// the grip, inside the clearance hole — or, for a countersunk joint, halfway
// through the thread's engagement, since a flat head can fill a thin grip and
// its socket is empty on the axis.
func jointRegion(j Joint, pv Prover) ([6]float64, [3]float64, error) {
	vals, err := pv.Eval(RegionExprs(j), [3]float64{})
	if err != nil {
		return [6]float64{}, [3]float64{}, err
	}
	want := []int{3, 3, 1, 1, 1, 1}
	if len(vals) != len(want) {
		return [6]float64{}, [3]float64{}, fmt.Errorf("evaluating the joint's placement: got %d values, want %d", len(vals), len(want))
	}
	for i, n := range want {
		if len(vals[i]) != n {
			return [6]float64{}, [3]float64{}, fmt.Errorf("evaluating the joint's placement: %q has %d components, want %d", RegionExprs(j)[i], len(vals[i]), n)
		}
	}
	seat := [3]float64{vals[0][0], vals[0][1], vals[0][2]}
	dir := [3]float64{vals[1][0], vals[1][1], vals[1][2]}
	box, err := Region(seat, dir, vals[2][0], vals[3][0], vals[4][0])
	if err != nil {
		return box, [3]float64{}, err
	}
	n := math.Sqrt(dir[0]*dir[0] + dir[1]*dir[1] + dir[2]*dir[2])
	along := vals[5][0] / 2
	if j.Kind == "countersunk" {
		along = vals[5][0] + 0.75*vals[2][0]
	}
	var mid [3]float64
	for i := range mid {
		mid[i] = seat[i] + dir[i]/n*along
	}
	return box, mid, nil
}

// clearOfAxis keeps the candidates that are empty at mid, the point on the
// screw's axis where only the screw can be (see jointRegion). A candidate solid there contains the screw — a display
// union of every part, say, which no text can see through a function call —
// and the claim against it would fail on the screw itself, so it is left out
// and named.
func clearOfAxis(candidates []string, mid [3]float64, pv Prover) (kept, unchecked []string) {
	if len(candidates) == 0 {
		return nil, nil
	}
	vals, err := pv.Eval(candidates, mid)
	if err != nil || len(vals) != len(candidates) {
		for _, c := range candidates {
			unchecked = append(unchecked, c+": could not be evaluated on the screw's axis")
		}
		return nil, unchecked
	}
	for i, c := range candidates {
		switch {
		case len(vals[i]) != 1:
			unchecked = append(unchecked, c+": not a single solid")
		case vals[i][0] <= 0:
			unchecked = append(unchecked, c+": solid on the screw's axis inside the grip — it contains the screw (a display union?) or fills its clearance hole")
		default:
			kept = append(kept, c)
		}
	}
	return kept, unchecked
}

// statement is one top-level statement, or one statement inside a body.
type statement struct {
	text   string
	line   int
	inBody bool
}

// statements splits src at semicolons outside brackets, and at the brace that
// closes a top-level function body. Statements inside a body are returned
// marked, so a joint called there can be reported.
func statements(src string) []statement {
	var out []statement
	depth, brace := 0, 0
	start, line, startLine := 0, 1, 1
	emit := func(end int, inBody bool) {
		if t := strings.TrimSpace(src[start:end]); t != "" {
			out = append(out, statement{text: t, line: startLine, inBody: inBody})
		}
		start = end + 1
		startLine = line
	}
	for i := 0; i < len(src); i++ {
		c := src[i]
		switch c {
		case '\n':
			line++
			if strings.TrimSpace(src[start:i]) == "" {
				startLine = line
			}
		case '(', '[':
			depth++
		case ')', ']':
			depth--
		case '{':
			if brace == 0 && depth == 0 {
				// A function body starts: its header is not a statement.
				start = i + 1
				startLine = line
			}
			brace++
		case '}':
			brace--
			if brace == 0 && depth == 0 {
				emit(i, true)
			}
		case ';':
			if depth == 0 {
				emit(i, brace > 0)
			}
		}
	}
	return out
}

// stripComments blanks // comments, keeping newlines so lines still count.
func stripComments(src string) string {
	var b strings.Builder
	for _, ln := range strings.SplitAfter(src, "\n") {
		if i := strings.Index(ln, "//"); i >= 0 {
			nl := ""
			if strings.HasSuffix(ln, "\n") {
				nl = "\n"
			}
			ln = ln[:i] + nl
		}
		b.WriteString(ln)
	}
	return b.String()
}

// matchParen is the index of the bracket closing the one at open, or -1.
func matchParen(s string, open int) int {
	depth := 0
	for i := open; i < len(s); i++ {
		switch s[i] {
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// splitTop splits s at commas outside brackets.
func splitTop(s string) []string {
	var out []string
	depth, start := 0, 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			depth--
		case ',':
			if depth == 0 {
				out = append(out, strings.TrimSpace(s[start:i]))
				start = i + 1
			}
		}
	}
	if t := strings.TrimSpace(s[start:]); t != "" {
		out = append(out, t)
	}
	return out
}
