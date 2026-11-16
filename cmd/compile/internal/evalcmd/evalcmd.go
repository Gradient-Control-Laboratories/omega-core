// Package evalcmd is `compile eval`, which `gsl eval` runs: questions asked of a GSL document's field -- its values
// and gradients at points, an extremum, a sound proof over a region, its bounds, points on its surface, its volume,
// a picture of it -- as JSON requests, each answered by the eval package.
//
// One request is read from a file or stdin and answered on stdout; or, with --stdio, a session answers one request
// per line for as long as stdin is open, keeping one evaluator (and its query cache) for each set of source paths.
package evalcmd

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/eval"
	"gradientcontrol.com/omega/core/parser"
)

// Exit codes of one request: answered (a disproven claim is an answer), the query failed, the request is malformed.
const (
	exitAnswered  = 0
	exitFailed    = 1
	exitMalformed = 2
)

// sessionCache is how many prepared queries a session's evaluator keeps (eval.Evaluator.CacheQueries).
const sessionCache = 64

const usage = `Usage of compile eval (gsl eval):
  gsl eval [-I DIR]... REQUEST.json    answer one JSON request (- reads it from stdin)
  gsl eval [-I DIR]... --stdio         a session: one JSON request per line in, one JSON response per line out

Flags:
  -I, --source-path string
        Root directory for package resolution (may be specified multiple times; default: the document's folder)
      --stdio
        Answer requests line by line until stdin closes
  -h, --help
        Display usage instructions

A request names its op and its document, by path ("source") or text ("code"):
  {"id": 1, "op": "eval", "source": "part.gsl", "exprs": ["body", "body.dp"], "points": [[0, 0, 5]]}

  op         fields beyond id, source|code, defs
  eval       exprs, points                            values and gradients at each point
  extremum   expr, region, mode, tolerance, lipschitz  min and/or max; the bound assumes the Lipschitz constant
  certify    expr, region, threshold, tolerance,      a sound proof that expr >= threshold over the region
             params | overSliders                     (and over the parameter ranges)
  bounds     expr, region, resolution                 a box proven to hold all of expr < 0 ("proven"); else a
                                                      fit on a grid that can miss thin walls, and "reason"
  surface    expr, points, maxIters, tolerance        each point moved onto expr = 0
  volume     expr, region, relativeTolerance,         the volume where expr < 0, bracketed; "converged" says the
             timeLimit                                bracket met the tolerance, "reason" why it did not; it
                                                      stops after timeLimit seconds, 10 unless named
  describe                                            the sliders, top-level names and returns
  render     out, expr, views, width, height, light,   a PNG at out, as compile -t png draws it; expr defaults
             extent, time                              to the returns, each in its own colour
  mesh       out, region, subdivision, algorithm,      a .glb at out: a node per return, a normal per vertex from
             time                                      the field's gradient; and each body's counts and volume
  shader     target, outDir                            the field as a library a host's shader calls, in "wgsl" or
                                                       "glsl-es300": gsl_field and gsl_march, a uniform block
                                                       for time and the sliders, and its textures as files

"defs" are GSL statements in the document's scope; one naming a declared name replaces it ("skin = 1.6;").
"expr" defaults to the union of the document's returns, and eval's "exprs" to each of them.
"region" is [x0, y0, z0, x1, y1, z1] in mm.
render's "views" is compile's --views ("iso,top,-1,1,0.6"), its size 640 by 480 unless named.
mesh's "region" defaults to the box the field fills (compile's --bbox auto), "subdivision" to 9 and "algorithm"
to "mdc" (or "marching-cubes"); the response's "region" is the box that was meshed.
shader's "code" has no entry point: a host adds its own, binds a buffer of "blockBytes" to "uniformBlock" and
writes each of "uniforms" and "sliders" at its "offset", and loads each of "textures" from its "path" (raw texels,
written under "outDir"). "region" is the box to march: the @scene box, or the box the field fills.
Each response is {"id", "ok": true, ...} or {"id", "ok": false, "error"}. A value that is not finite is null.
Exit codes (one request): 0 answered, 1 the query failed, 2 a malformed request.
`

// Run is `compile eval ARGS`; it returns the process's exit code.
func Run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	var paths sourcePaths
	flags := flag.NewFlagSet("eval", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.Var(&paths, "I", "")
	flags.Var(&paths, "source-path", "")
	stdio := flags.Bool("stdio", false, "")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprint(stdout, usage)
			return exitAnswered
		}
		fmt.Fprintf(stderr, "eval: %v\n\n%s", err, usage)
		return exitMalformed
	}
	s := &server{paths: paths, cache: *stdio, evaluators: map[string]*eval.Evaluator{}}
	if *stdio {
		if flags.NArg() != 0 {
			fmt.Fprintf(stderr, "eval: --stdio reads its requests from stdin, not %q\n", flags.Arg(0))
			return exitMalformed
		}
		return s.session(ctx, stdin, stdout, stderr)
	}
	if flags.NArg() != 1 {
		fmt.Fprintf(stderr, "eval: name one request file, or - to read it from stdin\n\n%s", usage)
		return exitMalformed
	}
	var data []byte
	var err error
	if flags.Arg(0) == "-" {
		data, err = io.ReadAll(stdin)
	} else {
		data, err = os.ReadFile(flags.Arg(0))
	}
	if err != nil {
		fmt.Fprintf(stderr, "eval: %v\n", err)
		return exitMalformed
	}
	response, code := s.answer(ctx, data)
	writeLine(stdout, response)
	return code
}

type sourcePaths []string

func (p *sourcePaths) String() string { return strings.Join(*p, ",") }

func (p *sourcePaths) Set(value string) error {
	*p = append(*p, value)
	return nil
}

// server answers requests with one evaluator per set of source paths and mesh folder.
type server struct {
	paths      []string
	cache      bool
	evaluators map[string]*eval.Evaluator
}

// evaluator is the evaluator for documents whose packages resolve against paths and whose importMesh files sit in
// docDir: a mesh is read from docDir, else from under a source path.
func (s *server) evaluator(paths []string, docDir string) *eval.Evaluator {
	key := strings.Join(append([]string{docDir}, paths...), "\x00")
	if e, ok := s.evaluators[key]; ok {
		return e
	}
	var opts []assembler.Option
	if len(paths) > 0 {
		opts = append(opts, assembler.SourcePaths(paths...))
	}
	e := eval.New(nil, nil, opts...)
	meshDirs := []string{docDir}
	for _, p := range paths {
		if filepath.Clean(p) != filepath.Clean(docDir) {
			meshDirs = append(meshDirs, p)
		}
	}
	e.LoadMeshesFrom(meshDirs...)
	if s.cache {
		e.CacheQueries(sessionCache)
	}
	s.evaluators[key] = e
	return e
}

// session answers each line of in on out until in closes. A request that fails is answered, never fatal.
func (s *server) session(ctx context.Context, in io.Reader, out, stderr io.Writer) int {
	reader := bufio.NewReaderSize(in, 1<<20)
	writer := bufio.NewWriter(out)
	for {
		line, err := reader.ReadBytes('\n')
		if len(bytes.TrimSpace(line)) > 0 {
			response, _ := s.answer(ctx, line)
			writeLine(writer, response)
			if flushErr := writer.Flush(); flushErr != nil {
				fmt.Fprintf(stderr, "eval: %v\n", flushErr)
				return exitFailed
			}
		}
		if err == io.EOF {
			return exitAnswered
		}
		if err != nil {
			fmt.Fprintf(stderr, "eval: reading requests: %v\n", err)
			return exitFailed
		}
	}
}

// request is every field any op reads; an op ignores the ones it does not use.
type request struct {
	ID                json.RawMessage       `json:"id"`
	Op                string                `json:"op"`
	Source            string                `json:"source"`
	Code              *string               `json:"code"`
	Defs              string                `json:"defs"`
	Expr              string                `json:"expr"`
	Exprs             []string              `json:"exprs"`
	Points            [][3]float64          `json:"points"`
	Region            *[6]float64           `json:"region"`
	Mode              string                `json:"mode"`
	Tolerance         float64               `json:"tolerance"`
	Lipschitz         float64               `json:"lipschitz"`
	Threshold         *float64              `json:"threshold"`
	Params            map[string][2]float64 `json:"params"`
	OverSliders       bool                  `json:"overSliders"`
	Resolution        int                   `json:"resolution"`
	MaxIters          int                   `json:"maxIters"`
	RelativeTolerance float64               `json:"relativeTolerance"`
	TimeLimit         float64               `json:"timeLimit"`
	Out               string                `json:"out"`
	Views             string                `json:"views"`
	Width             int                   `json:"width"`
	Height            int                   `json:"height"`
	Light             bool                  `json:"light"`
	Extent            float64               `json:"extent"`
	Time              float64               `json:"time"`
	Subdivision       int                   `json:"subdivision"`
	Algorithm         string                `json:"algorithm"`
	Target            string                `json:"target"`
	OutDir            string                `json:"outDir"`
}

// malformed is a request the command cannot read or that lacks what its op needs: exit 2, not a query that failed.
type malformed struct{ msg string }

func (m malformed) Error() string { return m.msg }

func malformedf(format string, args ...any) error { return malformed{fmt.Sprintf(format, args...)} }

// answer is the response to one request and the exit code it earns.
func (s *server) answer(ctx context.Context, data []byte) (map[string]any, int) {
	var req request
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&req)
	if err == nil && decoder.More() {
		err = errors.New("one JSON object per request")
	}
	if err != nil {
		var id struct {
			ID json.RawMessage `json:"id"`
		}
		_ = json.Unmarshal(data, &id)
		return failure(id.ID, fmt.Sprintf("a malformed request: %v", err)), exitMalformed
	}
	fields, err := s.dispatch(ctx, &req)
	if err != nil {
		code := exitFailed
		var bad malformed
		if errors.As(err, &bad) {
			code = exitMalformed
		}
		return failure(req.ID, err.Error()), code
	}
	response := map[string]any{"id": req.ID, "ok": true}
	for k, v := range fields {
		response[k] = v
	}
	return response, exitAnswered
}

func failure(id json.RawMessage, msg string) map[string]any {
	return map[string]any{"id": id, "ok": false, "error": msg}
}

func (s *server) dispatch(ctx context.Context, req *request) (map[string]any, error) {
	switch req.Op {
	case "eval", "extremum", "certify", "bounds", "surface", "volume", "describe", "render", "mesh", "shader":
	case "":
		return nil, malformedf("name an op: eval, extremum, certify, bounds, surface, volume, describe, render, mesh or shader")
	default:
		return nil, malformedf("op %q: one of eval, extremum, certify, bounds, surface, volume, describe, render, mesh or shader", req.Op)
	}
	if req.Op != "eval" && len(req.Exprs) > 0 {
		return nil, malformedf("exprs is eval's; %s takes one expr", req.Op)
	}
	code, paths, docDir, err := s.document(req)
	if err != nil {
		return nil, err
	}
	e := s.evaluator(paths, docDir)
	switch req.Op {
	case "eval":
		return evalPoints(ctx, e, code, req)
	case "extremum":
		return extremum(ctx, e, code, req)
	case "certify":
		return certify(ctx, e, code, req)
	case "bounds":
		return bounds(ctx, e, code, req)
	case "surface":
		return surface(ctx, e, code, req)
	case "volume":
		return volume(ctx, e, code, req)
	case "render":
		return render(ctx, e, code, req)
	case "mesh":
		return meshFile(ctx, e, code, req)
	case "shader":
		return shaderLibrary(ctx, e, code, req)
	default:
		return describe(code)
	}
}

// document is the request's GSL text, the source paths its imports resolve against (-I when given, else the folder
// of the file it names) and the folder its importMesh files sit in (the file's, or the working directory for code).
func (s *server) document(req *request) (string, []string, string, error) {
	var code string
	paths := s.paths
	docDir := "."
	switch {
	case req.Source != "" && req.Code != nil:
		return "", nil, "", malformedf("name the document once: source or code, not both")
	case req.Code != nil:
		code = *req.Code
	case req.Source != "":
		data, err := os.ReadFile(req.Source)
		if err != nil {
			return "", nil, "", fmt.Errorf("reading %s: %w", req.Source, err)
		}
		code = string(data)
		docDir = filepath.Dir(req.Source)
		if len(paths) == 0 {
			paths = []string{docDir}
		}
	default:
		return "", nil, "", malformedf("name the document: source (a .gsl path) or code (its text)")
	}
	return code, paths, docDir, nil
}

// expr is the request's expression, or the union of the document's returns.
func expr(req *request, code string) (string, error) {
	if req.Expr != "" {
		return req.Expr, nil
	}
	returns := eval.TopLevelReturnExprs(code)
	switch len(returns) {
	case 0:
		return "", errors.New("the document returns nothing: name an expr")
	case 1:
		return returns[0], nil
	default:
		return "min(" + strings.Join(returns, ", ") + ")", nil
	}
}

// region is the request's region, which must be given and have extent on every axis.
func region(req *request) ([6]float64, error) {
	if req.Region == nil {
		return [6]float64{}, malformedf("%s names its region, [x0, y0, z0, x1, y1, z1] in mm: it is part of the answer", req.Op)
	}
	return checkRegion(*req.Region)
}

func checkRegion(r [6]float64) ([6]float64, error) {
	for k := 0; k < 3; k++ {
		if !(r[k+3] > r[k]) {
			return r, malformedf("region %v: its max must exceed its min on every axis", r)
		}
	}
	return r, nil
}

func nonNegative(name string, v float64) error {
	if v < 0 || math.IsNaN(v) {
		return malformedf("%s %g: zero for the default, or more", name, v)
	}
	return nil
}

func evalPoints(ctx context.Context, e *eval.Evaluator, code string, req *request) (map[string]any, error) {
	if req.Expr != "" {
		return nil, malformedf("eval takes exprs, a list, not expr")
	}
	exprs := req.Exprs
	if len(exprs) == 0 {
		// As expr defaults to the returns for the other ops; here each is its own group.
		if exprs = eval.TopLevelReturnExprs(code); len(exprs) == 0 {
			return nil, errors.New("the document returns nothing: name its exprs")
		}
	}
	if len(req.Points) == 0 {
		return nil, malformedf("eval names its points, [[x, y, z], ...] in mm")
	}
	groups, err := e.EvalGsl(ctx, nil, code, req.Defs, exprs, req.Points)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(groups))
	for i, g := range groups {
		values := make([][]f32, len(g.Samples))
		gradients := make([][][3]f32, len(g.Samples))
		for c, component := range g.Samples {
			for _, sample := range component {
				values[c] = append(values[c], f32(sample.Value))
				gradients[c] = append(gradients[c], [3]f32{f32(sample.Dx), f32(sample.Dy), f32(sample.Dz)})
			}
		}
		group := map[string]any{"type": g.TypeName, "values": values, "gradients": gradients}
		// A call returning several values gives several groups for one expr: then a group is known by its place only.
		if len(groups) == len(exprs) {
			group["expr"] = exprs[i]
		}
		out = append(out, group)
	}
	return map[string]any{"groups": out}, nil
}

func extremum(ctx context.Context, e *eval.Evaluator, code string, req *request) (map[string]any, error) {
	switch req.Mode {
	case "", "min", "max", "both":
	default:
		return nil, malformedf("mode %q: min, max or both", req.Mode)
	}
	if err := errors.Join(nonNegative("tolerance", req.Tolerance), nonNegative("lipschitz", req.Lipschitz)); err != nil {
		return nil, err
	}
	box, err := region(req)
	if err != nil {
		return nil, err
	}
	x, err := expr(req, code)
	if err != nil {
		return nil, err
	}
	r, err := e.Extremum(ctx, nil, code, req.Defs, x, box, req.Mode, req.Tolerance, req.Lipschitz)
	if err != nil {
		return nil, err
	}
	out := map[string]any{
		"expr": x, "mode": r.Mode, "tolerance": f64(r.Tolerance), "evals": r.Evals,
		// The bound holds only if |grad expr| <= lipschitz everywhere in the region; the sample is a true value.
		"assumes": map[string]any{"lipschitz": f64(r.Lipschitz)},
	}
	if r.Mode == "min" || r.Mode == "both" {
		out["min"], out["minAt"], out["minLower"], out["minConverged"] = f64(r.Min), vec(r.MinAt[:]), f64(r.MinLower), r.MinConverged
	}
	if r.Mode == "max" || r.Mode == "both" {
		out["max"], out["maxAt"], out["maxUpper"], out["maxConverged"] = f64(r.Max), vec(r.MaxAt[:]), f64(r.MaxUpper), r.MaxConverged
	}
	return out, nil
}

func certify(ctx context.Context, e *eval.Evaluator, code string, req *request) (map[string]any, error) {
	if req.Threshold == nil {
		return nil, malformedf("certify names its threshold: it proves expr >= threshold")
	}
	if req.OverSliders && len(req.Params) > 0 {
		return nil, malformedf("certify over the document's sliders (overSliders) or over params, not both")
	}
	if err := nonNegative("tolerance", req.Tolerance); err != nil {
		return nil, err
	}
	box, err := region(req)
	if err != nil {
		return nil, err
	}
	x, err := expr(req, code)
	if err != nil {
		return nil, err
	}
	var r *eval.CertifyResult
	switch {
	case req.OverSliders:
		r, err = e.CertifyOverSliders(ctx, nil, code, req.Defs, x, box, *req.Threshold, req.Tolerance)
	case len(req.Params) > 0:
		r, err = e.CertifyParams(ctx, nil, code, req.Defs, x, box, *req.Threshold, req.Tolerance, req.Params)
	default:
		r, err = e.Certify(ctx, nil, code, req.Defs, x, box, *req.Threshold, req.Tolerance)
	}
	if err != nil {
		return nil, err
	}
	out := map[string]any{
		"expr": x, "threshold": f64(r.Threshold), "tolerance": f64(r.Tolerance), "evals": r.Evals,
		"regionLower": f64(r.RegionLower), "regionUpper": f64(r.RegionUpper),
	}
	if r.Params != nil {
		out["params"] = ranges(r.Params)
	}
	switch {
	case r.Proven:
		out["verdict"] = "proven"
	case r.Disproven:
		out["verdict"] = "disproven"
		out["counterBox"], out["counterUpper"] = vec(r.CounterBox[:]), f64(r.CounterUpper)
		if r.CounterParams != nil {
			out["counterParams"] = ranges(r.CounterParams)
		}
	default:
		out["verdict"], out["reason"] = "indeterminate", r.Reason
		out["worstBox"], out["worstLower"] = vec(r.WorstBox[:]), f64(r.WorstLower)
		if r.WorstParams != nil {
			out["worstParams"] = ranges(r.WorstParams)
		}
	}
	return out, nil
}

func bounds(ctx context.Context, e *eval.Evaluator, code string, req *request) (map[string]any, error) {
	var box [6]float64 // zero: the evaluator's default search region
	if req.Region != nil {
		var err error
		if box, err = checkRegion(*req.Region); err != nil {
			return nil, err
		}
	}
	if req.Resolution < 0 {
		return nil, malformedf("resolution %d: zero for the default, or more", req.Resolution)
	}
	x, err := expr(req, code)
	if err != nil {
		return nil, err
	}
	r, err := e.Enclose(ctx, nil, code, req.Defs, x, box, req.Resolution)
	if err != nil {
		return nil, err
	}
	out := map[string]any{
		"expr": x, "found": r.Found, "bboxMin": vec(r.BboxMin[:]), "bboxMax": vec(r.BboxMax[:]),
		"cellSize": vec(r.CellSize[:]), "resolution": r.Resolution, "minValue": f32(r.MinValue), "minAt": vec(r.MinAt[:]),
		"proven": r.Proven, "evals": r.Evals,
	}
	if !r.Proven {
		out["reason"] = r.Reason
	}
	return out, nil
}

func surface(ctx context.Context, e *eval.Evaluator, code string, req *request) (map[string]any, error) {
	if len(req.Points) == 0 {
		return nil, malformedf("surface names its points, [[x, y, z], ...] in mm")
	}
	if req.MaxIters < 0 {
		return nil, malformedf("maxIters %d: zero for the default, or more", req.MaxIters)
	}
	if err := nonNegative("tolerance", req.Tolerance); err != nil {
		return nil, err
	}
	x, err := expr(req, code)
	if err != nil {
		return nil, err
	}
	samples, err := e.SurfacePoint(ctx, nil, code, req.Defs, x, req.Points, req.MaxIters, req.Tolerance)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, len(samples))
	for i, sample := range samples {
		out[i] = map[string]any{"point": vec(sample.Point[:]), "normal": vec(sample.Normal[:]),
			"value": f64(sample.Value), "converged": sample.Converged}
	}
	return map[string]any{"expr": x, "samples": out}, nil
}

func volume(ctx context.Context, e *eval.Evaluator, code string, req *request) (map[string]any, error) {
	if err := nonNegative("relativeTolerance", req.RelativeTolerance); err != nil {
		return nil, err
	}
	if err := nonNegative("timeLimit", req.TimeLimit); err != nil {
		return nil, err
	}
	box, err := region(req)
	if err != nil {
		return nil, err
	}
	x, err := expr(req, code)
	if err != nil {
		return nil, err
	}
	limit := time.Duration(req.TimeLimit * float64(time.Second))
	if req.TimeLimit > 0 && limit <= 0 {
		limit = 1 // a limit too short for a nanosecond is the shortest one, not the default
	}
	r, err := e.VolumeWithin(ctx, nil, code, req.Defs, x, box, req.RelativeTolerance, limit)
	if err != nil {
		return nil, err
	}
	out := map[string]any{"expr": x, "lowerMm3": f64(r.Lower), "upperMm3": f64(r.Upper), "evals": r.Evals,
		"converged": r.Converged}
	if !r.Converged {
		out["reason"] = r.Reason
	}
	return out, nil
}

// The panel size render draws when the request names none: compile -t png's.
const (
	renderWidth  = 640
	renderHeight = 480
)

// render draws expr, or the document's returns each in its own colour, into the PNG at req.Out: compile -t png's
// picture, with defs and expr.
func render(ctx context.Context, e *eval.Evaluator, code string, req *request) (map[string]any, error) {
	if req.Out == "" {
		return nil, malformedf("render names out, the PNG to write")
	}
	views := req.Views
	if views == "" {
		views = "iso"
	}
	cameras, err := eval.ParseViews(views)
	if err != nil {
		return nil, malformedf("views: %v", err)
	}
	width, height := req.Width, req.Height
	if width == 0 {
		width = renderWidth
	}
	if height == 0 {
		height = renderHeight
	}
	if width < 1 || height < 1 {
		return nil, malformedf("a panel of %dx%d pixels: each side is at least 1", width, height)
	}
	if err := nonNegative("extent", req.Extent); err != nil {
		return nil, err
	}
	for i := range cameras {
		cameras[i].Light = req.Light
	}
	data, err := e.RenderViews(ctx, nil, code, req.Defs, req.Expr, cameras, width, height, req.Extent, req.Time)
	if err != nil {
		return nil, err
	}
	out, err := filepath.Abs(req.Out)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(out, data, 0o644); err != nil {
		return nil, err
	}
	sheetWidth, sheetHeight := pngSize(data)
	return map[string]any{"out": out, "width": sheetWidth, "height": sheetHeight, "bytes": len(data)}, nil
}

// meshFile writes the document, with defs, as a .glb at req.Out, and answers with the box meshed and what each
// body's triangles say of it. The file is written only once the whole mesh is made.
func meshFile(ctx context.Context, e *eval.Evaluator, code string, req *request) (map[string]any, error) {
	if req.Out == "" {
		return nil, malformedf("mesh names out, the .glb to write")
	}
	if req.Expr != "" {
		return nil, malformedf("mesh takes no expr: it meshes the document's returns, each a body")
	}
	if req.Subdivision < 0 {
		return nil, malformedf("subdivision %d: zero for the default, or more", req.Subdivision)
	}
	switch req.Algorithm {
	case "", "mdc", "marching-cubes":
	default:
		return nil, malformedf("algorithm %q: mdc or marching-cubes", req.Algorithm)
	}
	opt := eval.MeshOptions{Subdivision: req.Subdivision, Algorithm: req.Algorithm, Time: req.Time}
	if req.Region != nil {
		box, err := checkRegion(*req.Region)
		if err != nil {
			return nil, err
		}
		opt.Region = &box
	}
	var buf bytes.Buffer
	r, err := e.Mesh(ctx, nil, code, req.Defs, &buf, opt)
	if err != nil {
		return nil, err
	}
	out, err := filepath.Abs(req.Out)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(out, buf.Bytes(), 0o644); err != nil {
		return nil, err
	}
	bodies := make([]map[string]any, len(r.Bodies))
	for i, b := range r.Bodies {
		bodies[i] = map[string]any{"name": b.Name, "index": b.Index, "vertices": b.Vertices, "triangles": b.Triangles,
			"openEdges": b.OpenEdges, "nonManifoldEdges": b.NonManifoldEdges, "volumeMm3": f64(b.VolumeMm3)}
	}
	region := make([]any, len(r.Region))
	for k, v := range r.Region {
		region[k] = f64(v)
	}
	return map[string]any{"out": out, "bytes": buf.Len(), "region": region, "bodies": bodies}, nil
}

// shaderLibrary answers with the document's field, with defs, as a library in req.Target's language, its textures
// written under req.OutDir.
func shaderLibrary(ctx context.Context, e *eval.Evaluator, code string, req *request) (map[string]any, error) {
	if req.Target != eval.ShaderWGSL && req.Target != eval.ShaderGLSLES300 {
		return nil, malformedf("shader names a target: %q or %q, got %q", eval.ShaderWGSL, eval.ShaderGLSLES300, req.Target)
	}
	if req.Expr != "" {
		return nil, malformedf("shader takes no expr: it is the document's returns, each a body")
	}
	outDir := req.OutDir
	if outDir != "" {
		abs, err := filepath.Abs(outDir)
		if err != nil {
			return nil, err
		}
		outDir = abs
	}
	r, err := e.Shader(ctx, nil, code, req.Defs, eval.ShaderOptions{Target: req.Target, OutDir: outDir})
	if err != nil {
		return nil, err
	}
	numbers := func(values []float64) []any {
		out := make([]any, len(values))
		for k, v := range values {
			out[k] = f64(v)
		}
		return out
	}
	uniforms := make([]map[string]any, len(r.Uniforms))
	for i, u := range r.Uniforms {
		uniforms[i] = map[string]any{"name": u.Name, "type": u.Type, "offset": u.Offset, "value": numbers(u.Value)}
	}
	sliders := make([]map[string]any, len(r.Sliders))
	for i, s := range r.Sliders {
		sliders[i] = map[string]any{"name": s.Name, "uniform": s.Uniform, "type": s.Type, "offset": s.Offset,
			"value": f64(s.Value), "min": f64(s.Min), "max": f64(s.Max)}
	}
	textures := make([]map[string]any, len(r.Textures))
	for i, t := range r.Textures {
		tex := map[string]any{"uniform": t.Uniform, "kind": t.Kind, "format": t.Format, "filter": t.Filter,
			"size": t.Size, "path": t.Path}
		// WGSL binds by number, and a texture read through a sampler has the sampler's too; GLSL binds by name.
		if req.Target == eval.ShaderWGSL {
			tex["binding"] = t.Binding
			if t.SamplerUniform != "" {
				tex["samplerUniform"], tex["samplerBinding"] = t.SamplerUniform, t.SamplerBinding
			}
		}
		textures[i] = tex
	}
	bodies := make([]map[string]any, len(r.Bodies))
	for i, b := range r.Bodies {
		bodies[i] = map[string]any{"name": b.Name, "index": b.Index}
	}
	return map[string]any{
		"target": r.Target, "code": r.Code, "entry": map[string]any{"field": r.FieldFn, "march": r.MarchFn},
		"uniformBlock": r.UniformBlock, "blockBytes": r.BlockSize, "uniforms": uniforms, "sliders": sliders,
		"textures": textures, "bodies": bodies, "region": numbers(r.Region[:]),
	}, nil
}

// pngSize is a PNG's width and height, read from its header.
func pngSize(data []byte) (int, int) {
	if len(data) < 24 {
		return 0, 0
	}
	return int(binary.BigEndian.Uint32(data[16:20])), int(binary.BigEndian.Uint32(data[20:24]))
}

func describe(code string) (map[string]any, error) {
	sliders, err := eval.SliderRanges(code)
	if err != nil {
		// SliderRanges refuses a document that declares none (it serves certify over the sliders); here that is
		// an answer, and only a document that does not parse is an error.
		if _, parseErr := parser.ParsePackage([]string{code}); parseErr != nil {
			return nil, err
		}
		sliders = map[string][2]float64{}
	}
	names := make([]string, 0)
	for name := range eval.TopLevelNames(code) {
		names = append(names, name)
	}
	sort.Strings(names)
	returns := eval.TopLevelReturnExprs(code)
	if returns == nil {
		returns = []string{}
	}
	return map[string]any{"sliders": ranges(sliders), "names": names, "returns": returns}, nil
}

// f32 and f64 write a number, or null when it is not finite (encoding/json refuses NaN and infinities).
type (
	f32 float32
	f64 float64
)

func (v f32) MarshalJSON() ([]byte, error) { return number(float64(v), 32), nil }
func (v f64) MarshalJSON() ([]byte, error) { return number(float64(v), 64), nil }

func number(v float64, bits int) []byte {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return []byte("null")
	}
	return strconv.AppendFloat(nil, v, 'g', -1, bits)
}

func vec(xs []float64) []f64 {
	out := make([]f64, len(xs))
	for i, x := range xs {
		out[i] = f64(x)
	}
	return out
}

func ranges(m map[string][2]float64) map[string][2]f64 {
	out := make(map[string][2]f64, len(m))
	for name, r := range m {
		out[name] = [2]f64{f64(r[0]), f64(r[1])}
	}
	return out
}

func writeLine(w io.Writer, response map[string]any) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(response); err != nil {
		buf.Reset()
		_ = json.NewEncoder(&buf).Encode(failure(nil, fmt.Sprintf("writing the response: %v", err)))
	}
	_, _ = w.Write(buf.Bytes())
}
