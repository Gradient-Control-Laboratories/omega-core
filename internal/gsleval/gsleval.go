// Package gsleval evaluates ad-hoc GSL expressions in the scope of an
// existing document, backing the AI chat's evalGsl tool. The query is
// synthesized as a single source: the document with its top-level return
// statements removed, followed by `return e1, e2, ...;`. Names resolve at
// parse time, so parse and build errors read exactly like ordinary GSL
// errors and can be relayed to the model verbatim.
//
// Documents that reference mesh assets (importStep / importMesh / bvhSdf /
// vdbSdf) additionally need the canvas agent's per-client asset registry:
// the optional AssetFetcher pulls the import-slot table and oracle payloads
// for the module, mirroring what the canvas packs into a fidget MeshRequest.
package gsleval

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"

	"google.golang.org/protobuf/proto"
	assembler "gradientcontrol.com/omega/core/internal/assembler"
	fidgetapi "gradientcontrol.com/omega/core/internal/fidget"
	"gradientcontrol.com/omega/core/internal/fidget/lib"
	"gradientcontrol.com/omega/core/internal/meshfold"
	noise "gradientcontrol.com/omega/core/internal/noise"
	"gradientcontrol.com/omega/core/internal/source"
	parser "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
	gsllibs "gradientcontrol.com/omega/core/stdlib"
	"log"
)

// AssetFetcher returns the mesh-asset context (import slots + oracle
// payloads) for a module. digestOnly requests skip the heavy payloads
// and return just slots + digests, for prompt context.
type AssetFetcher func(ctx context.Context, manifest map[string]string, module []byte, digestOnly bool) (*fidgetapi.EvalAssetsResponse, error)

// Evaluator evaluates GSL query expressions against a document via fidget.
type Evaluator struct {
	srcMgr          source.Manager
	fetchAssets     AssetFetcher
	assembleOptions []assembler.Option

	// certifyBudget overrides certify's per-proof interval-eval budget
	// when positive; see SetCertifyBudget.
	certifyBudget int

	// queryCacheSize, when positive, is how many prepared queries
	// queryCache keeps, most recently used last; see CacheQueries.
	queryCacheSize int
	queryCache     []preparedQuery
	queryCacheMu   sync.Mutex
}

// preparedQuery is one cached query: the source it was assembled from and
// the program.
type preparedQuery struct {
	source string
	prog   *program.Program
}

// CacheQueries keeps the n most recently prepared queries, so a query asked
// again of the same document is not assembled again. A sweep proving a value
// over each side of its box asks the same query once a side, and assembling
// the document and its libraries costs far more than a proof that decides in
// one interval evaluation. The cache does not watch the source manager, so a
// caller whose libraries can change under it must leave it off, as it is by
// default. A query that references mesh assets is never cached: its oracles
// are registered for that query. Zero or negative turns it off.
func (e *Evaluator) CacheQueries(n int) {
	e.queryCacheMu.Lock()
	defer e.queryCacheMu.Unlock()
	e.queryCacheSize = n
	e.queryCache = nil
}

// cachedQuery returns the program prepared from source, or nil.
func (e *Evaluator) cachedQuery(source string) *program.Program {
	e.queryCacheMu.Lock()
	defer e.queryCacheMu.Unlock()
	for i, q := range e.queryCache {
		if q.source == source {
			e.queryCache = append(append(e.queryCache[:i:i], e.queryCache[i+1:]...), q)
			return q.prog
		}
	}
	return nil
}

// cacheQuery keeps prog as the program prepared from source, dropping the
// least recently used past the cache's size.
func (e *Evaluator) cacheQuery(source string, prog *program.Program) {
	e.queryCacheMu.Lock()
	defer e.queryCacheMu.Unlock()
	if e.queryCacheSize <= 0 {
		return
	}
	e.queryCache = append(e.queryCache, preparedQuery{source: source, prog: prog})
	if over := len(e.queryCache) - e.queryCacheSize; over > 0 {
		e.queryCache = append(e.queryCache[:0:0], e.queryCache[over:]...)
	}
}

// SetCertifyBudget raises (or lowers) certify's interval-evaluation budget
// from its default. Interval evaluations cost microseconds, so a caller
// prepared to wait seconds can afford tens of millions — a joint proof over
// space, a motion range, and slider ranges may genuinely need them. The
// chat tool loop deliberately leaves this unset: the default is its
// runaway guard. Zero or negative restores the default.
func (e *Evaluator) SetCertifyBudget(n int) {
	e.certifyBudget = n
}

var noiseTileOnce sync.Once

// Queries retain exclusive ownership of the process-wide oracle slots until evaluation finishes.
var querySlot = make(chan struct{}, 1)
var registeredOracles = make(map[uint32]string)

func startQuery(ctx context.Context) error {
	select {
	case querySlot <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func endQuery() { <-querySlot }

// New returns an Evaluator. `srcMgr` resolves GSL library
// imports in evaluated documents; pass nil if imports are not supported.
// `fetchAssets` supplies mesh-asset payloads for documents that reference
// them; pass nil if mesh assets are not supported.
func New(srcMgr source.Manager, fetchAssets AssetFetcher, options ...assembler.Option) *Evaluator {
	// The noiseHash3 oracle samples a fixed baked tile that survives
	// fidget.Clear(); register it once so noise fields evaluate.
	noiseTileOnce.Do(func() {
		if err := lib.RegisterNoiseTile(noise.Bake()); err != nil {
			log.Printf("gsleval: register noise tile: %v", err)
		}
	})
	if srcMgr == nil {
		srcMgr = source.NewStdLibManager()
	} else {
		srcMgr = source.NewMultiSourceManager(srcMgr, source.NewStdLibManager())
	}
	cp := parser.NewCachingParser()
	opts := append([]assembler.Option{assembler.Parser(cp.ParsePackage)}, options...)
	return &Evaluator{
		srcMgr:          srcMgr,
		fetchAssets:     fetchAssets,
		assembleOptions: opts,
	}
}

// EvalGsl evaluates each expression at every point in the scope of the
// document's top-level declarations. Each return value — scalar, vector, or
// matrix — becomes one GradGroup carrying its GSL type name and one sample
// slice per scalar component, in order (a multi-value function call yields
// one group per value). `defs` is optional GSL statements evaluated in the
// document's scope before the query — scratch variables for multi-step
// computations. manifest maps asset filenames to content hashes backing mesh
// assets; it may be nil for documents without mesh assets.
func (e *Evaluator) EvalGsl(ctx context.Context, manifest map[string]string, code, defs string, exprs []string, points [][3]float64) ([]fidgetapi.GradGroup, error) {
	if err := startQuery(ctx); err != nil {
		return nil, err
	}
	defer endQuery()
	prog, err := e.prepareQuery(ctx, manifest, code, defs, exprs)
	if err != nil {
		return nil, err
	}

	xs := make([]float32, len(points))
	ys := make([]float32, len(points))
	zs := make([]float32, len(points))
	for i, pt := range points {
		xs[i] = float32(pt[0])
		ys[i] = float32(pt[1])
		zs[i] = float32(pt[2])
	}

	return lib.EvalGradGroups(prog, xs, ys, zs)
}

// prepareQuery parses the document with the given expressions substituted
// as its returns, and applies mesh-asset context when the source needs it.
// `defs` is optional GSL statements appended after the document, in its
// scope, before the query return.
func (e *Evaluator) prepareQuery(ctx context.Context, manifest map[string]string, code, defs string, exprs []string) (*program.Program, error) {
	if len(exprs) == 0 {
		return nil, fmt.Errorf("no expressions to evaluate")
	}

	// Clean the defs: drop import statements (all libraries are injected
	// automatically) and dedup repeated definitions of the same name,
	// keeping the last — so a caller may redefine a name across accumulated
	// calls without an SSA collision.
	defStmts := dedupDefs(defs)
	body := stripTopLevelReturns(code)
	// The document body may carry redundant (or malformed) imports of the
	// auto-injected built-ins — a returned program from a model commonly does.
	// Drop those so injectLibraryImports below can supply the correct form;
	// omega and product-library imports are preserved.
	body = stripRedundantLeadingImports(body)

	// A def whose name the document already declares overrides it in place
	// (GSL resolution is order-dependent, so the override must sit where the
	// document declared the name); a def with a new name is appended after
	// the body so it can still reference document names.
	docNames := topLevelNames(body)
	overrides := make(map[string]string)
	var appended strings.Builder
	for _, s := range defStmts {
		if s.name != "" && docNames[s.name] {
			overrides[s.name] = s.text
		} else {
			appended.WriteString(s.text)
			appended.WriteString("\n")
		}
	}
	if len(overrides) > 0 {
		body = substituteTopLevelDecls(body, overrides)
	}

	gslSource := body + "\n" + appended.String() + "\nreturn " + strings.Join(exprs, ", ") + ";\n"

	meshAssets := referencesMeshAssets(gslSource)
	if !meshAssets {
		if prog := e.cachedQuery(gslSource); prog != nil {
			return prog, nil
		}
	}

	res, err := assembler.Assemble([]string{gslSource}, e.srcMgr, e.assembleOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to assemble query: %w", err)
	}
	prog := res.Program
	if !meshAssets {
		e.cacheQuery(gslSource, prog)
	}

	// Documents referencing mesh assets need the canvas's slot table and
	// oracle payloads before they can build. The textual filter errs open:
	// a false positive costs one round trip, a mention inside a comment is
	// harmless.
	if meshAssets {
		if e.fetchAssets == nil {
			return nil, fmt.Errorf("this document references mesh assets (importStep/importMesh/bvhSdf/vdbSdf), which evalGsl cannot evaluate here")
		}
		prog, err = e.applyMeshAssets(ctx, manifest, prog)
		if err != nil {
			return nil, err
		}
	}

	return prog, nil
}

// BoundsResult reports the axis-aligned box of the region where an
// expression is negative (inside), fit on a resolution^3 grid, padded by
// up to one cell.
type BoundsResult struct {
	Found      bool
	BboxMin    [3]float64
	BboxMax    [3]float64
	CellSize   [3]float64
	Resolution int
	// MinValue is the smallest sampled value and MinAt its location —
	// reported even when nothing was inside, so a caller can tell how far
	// off (and in which direction) the nearest geometry is.
	MinValue float32
	MinAt    [3]float64
}

const (
	defaultBoundsResolution = 32
	maxBoundsResolution     = 64
	// defaultBoundsRegion is the half-width of the coarse first-pass region
	// when the caller gives none.
	defaultBoundsRegion = 1024.0
)

// Bounds fits an axis-aligned bounding box to the region where expr is
// negative: a coarse pass over `region` (or a large default) finds occupied
// cells, then a second pass at the same resolution over the occupied box
// tightens it. Cells count as occupied when the sampled value is below the
// cell half-diagonal, so distance-like fields smaller than one coarse cell
// are still caught.
func (e *Evaluator) Bounds(ctx context.Context, manifest map[string]string, code, defs, expr string, region [6]float64, resolution int) (*BoundsResult, error) {
	if err := startQuery(ctx); err != nil {
		return nil, err
	}
	defer endQuery()
	if resolution <= 0 {
		resolution = defaultBoundsResolution
	}
	if resolution > maxBoundsResolution {
		resolution = maxBoundsResolution
	}
	if region == ([6]float64{}) {
		region = [6]float64{
			-defaultBoundsRegion, -defaultBoundsRegion, -defaultBoundsRegion,
			defaultBoundsRegion, defaultBoundsRegion, defaultBoundsRegion,
		}
	}
	for k := 0; k < 3; k++ {
		if region[k+3] <= region[k] {
			return nil, fmt.Errorf("bounds region max must exceed min on every axis")
		}
	}

	prog, err := e.prepareQuery(ctx, manifest, code, defs, []string{expr})
	if err != nil {
		return nil, err
	}

	pass1, err := gridScan(prog, region, resolution)
	if err != nil {
		return nil, err
	}
	result := &BoundsResult{
		Resolution: resolution,
		CellSize:   pass1.cellSize,
		MinValue:   pass1.minValue,
		MinAt:      pass1.minAt,
	}
	if !pass1.found {
		return result, nil
	}

	// Refine: rescan the occupied box (inflated by one coarse cell) at the
	// same resolution.
	var refined [6]float64
	for k := 0; k < 3; k++ {
		refined[k] = pass1.bboxMin[k] - pass1.cellSize[k]
		refined[k+3] = pass1.bboxMax[k] + pass1.cellSize[k]
	}
	pass2, err := gridScan(prog, refined, resolution)
	if err != nil {
		return nil, err
	}
	if !pass2.found {
		// The coarse hit fell between fine samples; report the coarse box.
		result.Found = true
		result.BboxMin = pass1.bboxMin
		result.BboxMax = pass1.bboxMax
		return result, nil
	}
	result.Found = true
	result.BboxMin = pass2.bboxMin
	result.BboxMax = pass2.bboxMax
	result.CellSize = pass2.cellSize
	if pass2.minValue < result.MinValue {
		result.MinValue = pass2.minValue
		result.MinAt = pass2.minAt
	}
	return result, nil
}

type gridScanResult struct {
	found            bool
	bboxMin, bboxMax [3]float64
	cellSize         [3]float64
	minValue         float32
	minAt            [3]float64
}

// gridScan samples the expression at resolution^3 cell centers over region
// and returns the union box of occupied cells. A cell is occupied when its
// value is negative; only when NO cell is negative does the scan fall back
// to the cell half-diagonal threshold (to catch a shape thinner than a
// cell). The fallback must not be the primary test: clamped fields like
// vdbSdf sit at a small constant background away from the surface, and a
// half-diagonal larger than that background would mark every cell occupied.
func gridScan(prog *program.Program, region [6]float64, resolution int) (*gridScanResult, error) {
	r := &gridScanResult{minValue: float32(math.Inf(1))}
	var cellSize [3]float64
	for k := 0; k < 3; k++ {
		cellSize[k] = (region[k+3] - region[k]) / float64(resolution)
	}
	r.cellSize = cellSize
	halfDiag := 0.5 * math.Sqrt(
		cellSize[0]*cellSize[0]+cellSize[1]*cellSize[1]+cellSize[2]*cellSize[2])

	n := resolution * resolution * resolution
	xs := make([]float32, 0, n)
	ys := make([]float32, 0, n)
	zs := make([]float32, 0, n)
	for iz := 0; iz < resolution; iz++ {
		z := region[2] + (float64(iz)+0.5)*r.cellSize[2]
		for iy := 0; iy < resolution; iy++ {
			y := region[1] + (float64(iy)+0.5)*r.cellSize[1]
			for ix := 0; ix < resolution; ix++ {
				x := region[0] + (float64(ix)+0.5)*r.cellSize[0]
				xs = append(xs, float32(x))
				ys = append(ys, float32(y))
				zs = append(zs, float32(z))
			}
		}
	}

	results, err := lib.EvalGrad(prog, xs, ys, zs)
	if err != nil {
		return nil, err
	}
	if len(results) != 1 {
		return nil, fmt.Errorf("bounds requires a single scalar expression, got %d values", len(results))
	}

	type box struct {
		found    bool
		min, max [3]float64
	}
	newBox := func() box {
		return box{
			min: [3]float64{math.Inf(1), math.Inf(1), math.Inf(1)},
			max: [3]float64{math.Inf(-1), math.Inf(-1), math.Inf(-1)},
		}
	}
	grow := func(b *box, c [3]float64) {
		b.found = true
		for k := 0; k < 3; k++ {
			b.min[k] = math.Min(b.min[k], c[k]-0.5*cellSize[k])
			b.max[k] = math.Max(b.max[k], c[k]+0.5*cellSize[k])
		}
	}

	negative := newBox()
	nearBand := newBox()
	nearCount := 0
	samples := results[0]
	for i, s := range samples {
		center := [3]float64{float64(xs[i]), float64(ys[i]), float64(zs[i])}
		if s.Value < r.minValue {
			r.minValue = s.Value
			r.minAt = center
		}
		if s.Value < 0 {
			grow(&negative, center)
		} else if float64(s.Value) < halfDiag {
			grow(&nearBand, center)
			nearCount++
		}
	}

	switch {
	case negative.found:
		r.found = true
		r.bboxMin, r.bboxMax = negative.min, negative.max
	case nearBand.found && nearCount < len(samples)*4/5:
		// The fallback is only meaningful when it discriminates: a clamped
		// field whose constant background sits below the half-diagonal
		// marks (nearly) every cell, and reporting the whole region as the
		// bbox would be worse than admitting nothing was found.
		r.found = true
		r.bboxMin, r.bboxMax = nearBand.min, nearBand.max
	}
	return r, nil
}

// DescribeAssets returns a compact JSON digest of every mesh asset the
// document references — exact bbox, centroid, and principal axes computed
// from the imported triangles — or "" when the document references none.
// Failures also return "": the digest is advisory prompt context and must
// never block a chat turn.
func (e *Evaluator) DescribeAssets(ctx context.Context, manifest map[string]string, code string) (string, error) {
	if e.fetchAssets == nil || !referencesMeshAssets(code) {
		return "", nil
	}
	res := parser.ParsePackage([]string{code})
	if res.Error() != nil || res == nil || res.Package == nil {
		// A document that doesn't parse gets no digest; the chat proceeds.
		return "", nil
	}
	moduleBytes, err := proto.Marshal(res.Package)
	if err != nil {
		return "", err
	}
	assets, err := e.fetchAssets(ctx, manifest, moduleBytes, true)
	if err != nil {
		return "", err
	}
	if assets.Error != "" || len(assets.Digests) == 0 {
		return "", nil
	}

	type mapRangeJSON struct {
		Min      float64    `json:"min"`
		MinAt    [3]float64 `json:"minAt"`
		Max      float64    `json:"max"`
		MaxAt    [3]float64 `json:"maxAt"`
		Coverage float64    `json:"coverage"`
	}
	roundRange := func(r *fidgetapi.MapRangeDigest) *mapRangeJSON {
		if r == nil {
			return nil
		}
		round := func(v float64) float64 { return math.Round(v*1e4) / 1e4 }
		return &mapRangeJSON{
			Min:      round(r.Min),
			MinAt:    roundVec(r.MinAt),
			Max:      round(r.Max),
			MaxAt:    roundVec(r.MaxAt),
			Coverage: round(r.Coverage),
		}
	}
	type digestJSON struct {
		Slot          int           `json:"slot"`
		ContentHash   string        `json:"contentHash,omitempty"`
		File          string        `json:"file,omitempty"`
		BboxMin       [3]float64    `json:"bboxMin"`
		BboxMax       [3]float64    `json:"bboxMax"`
		Centroid      [3]float64    `json:"centroid"`
		PrincipalAxes [3][3]float64 `json:"principalAxes"`
		HalfExtents   [3]float64    `json:"halfExtents"`
		// Ready-to-paste 4x4 transforms in GSL's column-list convention.
		// toLocal maps world points into the part's principal frame
		// (centroid at origin, longest axis = x); fromLocal is its inverse.
		ToLocal   [4][4]float64 `json:"toLocal"`
		FromLocal [4][4]float64 `json:"fromLocal"`
		// Whole-mesh summaries of hodge's surface-analysis maps (mm),
		// present on annotated STEP imports; probe locally via meshMaps.
		Thickness *mapRangeJSON `json:"thickness,omitempty"`
		Clearance *mapRangeJSON `json:"clearance,omitempty"`
	}
	out := make([]digestJSON, len(assets.Digests))
	for i, d := range assets.Digests {
		toLocal, fromLocal := frameMatrices(d.PrincipalAxes, d.Centroid)
		out[i] = digestJSON{
			Slot:          d.Slot,
			ContentHash:   d.ContentHash,
			File:          d.Filename,
			BboxMin:       roundVec(d.BboxMin),
			BboxMax:       roundVec(d.BboxMax),
			Centroid:      roundVec(d.Centroid),
			PrincipalAxes: [3][3]float64{roundVec(d.PrincipalAxes[0]), roundVec(d.PrincipalAxes[1]), roundVec(d.PrincipalAxes[2])},
			HalfExtents:   roundVec(d.HalfExtents),
			ToLocal:       toLocal,
			FromLocal:     fromLocal,
			Thickness:     roundRange(d.Thickness),
			Clearance:     roundRange(d.Clearance),
		}
	}
	payload, err := json.Marshal(map[string]any{"assets": out})
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

// roundVec trims digest values to 4 decimals — plenty for placement
// context, and it keeps the prompt free of 17-digit float noise.
func roundVec(v [3]float64) [3]float64 {
	for k := 0; k < 3; k++ {
		v[k] = math.Round(v[k]*1e4) / 1e4
	}
	return v
}

// frameMatrices builds the part-frame transforms from the digest's
// principal axes (rows of the rotation R) and centroid c, in GSL's
// column-list convention (vector.multiply combines columns):
//
//	toLocal · [p, 1] = R(p - c)   — world point into the principal frame
//	fromLocal · [l, 1] = Rᵀl + c  — principal-frame point back to world
//
// Rounded to 6 decimals: rotation entries compound, so they get more
// precision than the positional digest fields.
func frameMatrices(axes [3][3]float64, c [3]float64) (toLocal, fromLocal [4][4]float64) {
	round := func(v float64) float64 { return math.Round(v*1e6) / 1e6 }
	dot := func(a, b [3]float64) float64 { return a[0]*b[0] + a[1]*b[1] + a[2]*b[2] }
	for j := 0; j < 3; j++ {
		for i := 0; i < 3; i++ {
			// toLocal's linear block is R (rows = axes): column j holds the
			// j-th entry of every row.
			toLocal[j][i] = round(axes[i][j])
			// fromLocal's linear block is Rᵀ: column j is axis j itself.
			fromLocal[j][i] = round(axes[j][i])
		}
	}
	for i := 0; i < 3; i++ {
		toLocal[3][i] = round(-dot(axes[i], c))
		fromLocal[3][i] = round(c[i])
	}
	toLocal[3][3] = 1
	fromLocal[3][3] = 1
	return toLocal, fromLocal
}

// applyMeshAssets fetches the module's mesh-asset context, registers the
// bvhSdf/vdbSdf oracles, and folds importStep/importMesh expressions to
// their literal slots — the same preparation the fidget agent does for a
// MeshRequest.
func (e *Evaluator) applyMeshAssets(ctx context.Context, manifest map[string]string, prog *program.Program) (*program.Program, error) {
	moduleBytes, err := proto.Marshal(prog.Primary)
	if err != nil {
		return nil, fmt.Errorf("marshal module for asset fetch: %w", err)
	}
	assets, err := e.fetchAssets(ctx, manifest, moduleBytes, false)
	if err != nil {
		return nil, fmt.Errorf("fetch mesh assets: %w", err)
	}
	if assets.Error != "" {
		return nil, fmt.Errorf("fetch mesh assets: %s", assets.Error)
	}
	if len(assets.SpinodoidBounds) == 6 {
		if prog.Primary.Scene == nil {
			prog.Primary.Scene = &program.Scene{}
		}
		prog.Primary.Scene.SpinodoidBounds = append([]float64(nil), assets.SpinodoidBounds...)
	}

	for _, payload := range assets.MeshOracles {
		if registeredOracles[payload.AssetID] == "mesh:"+payload.VerticesB64 {
			continue
		}
		verts, tris, err := decodePackedTriangles(payload.VerticesB64)
		if err != nil {
			return nil, fmt.Errorf("mesh oracle %d decode: %w", payload.AssetID, err)
		}
		lib.Lock()
		err = lib.RegisterMeshOracle(payload.AssetID, verts, tris)
		lib.Unlock()
		if err != nil {
			return nil, fmt.Errorf("mesh oracle %d register: %w", payload.AssetID, err)
		}
		registeredOracles[payload.AssetID] = "mesh:" + payload.VerticesB64
	}
	for _, payload := range assets.VdbOracles {
		key := fmt.Sprintf("vdb:%v:%v:", payload.VoxelSize, payload.HalfBandVoxels) + payload.VerticesB64
		if registeredOracles[payload.AssetID] == key {
			continue
		}
		verts, tris, err := decodePackedTriangles(payload.VerticesB64)
		if err != nil {
			return nil, fmt.Errorf("vdb oracle %d decode: %w", payload.AssetID, err)
		}
		lib.Lock()
		err = lib.RegisterVdbOracle(payload.AssetID, verts, tris, payload.VoxelSize, payload.HalfBandVoxels)
		lib.Unlock()
		if err != nil {
			return nil, fmt.Errorf("vdb oracle %d register: %w", payload.AssetID, err)
		}
		registeredOracles[payload.AssetID] = key
	}

	// Fold expression-form imports to literal slots (matching the fidget
	// agent's foldImportSlots) so build never sees the importStep builtin.
	if len(assets.ImportSlots) == 0 {
		return prog, nil
	}
	type key struct{ pkgPath, kind, path string }
	bySlot := make(map[key]int, len(assets.ImportSlots))
	for _, s := range assets.ImportSlots {
		bySlot[key{s.PkgPath, s.Kind, s.Path}] = s.Slot
	}
	resolver := func(pkgPath, kind, path string) (int, bool) {
		v, ok := bySlot[key{pkgPath, kind, path}]
		return v, ok
	}
	folded, err := meshfold.FoldProgramMeshIndices(prog, resolver)
	if err != nil {
		return nil, err
	}
	return folded, nil
}

// isAutoImportable reports whether the compiler linker can re-add an import
// for lib automatically, so a redundant explicit import of it in query source
// is safe to drop. omega is excluded on purpose: it is dot-imported and cannot
// be auto-injected, so an omega import must survive.
func isAutoImportable(lib string) bool {
	if lib == "builtin" || lib == "omega" {
		return false
	}
	for _, info := range gsllibs.All() {
		if info.ImportPath == lib {
			return true
		}
	}
	return false
}

func isIdentByte(c byte) bool {
	return c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
}

// stripRedundantLeadingImports removes leading top-level `import "<lib>"`
// statements whose library the compiler linker re-adds automatically, so a
// model's redundant — or slightly malformed, e.g. missing the trailing
// semicolon — import of a built-in does not break the query parse. It delimits
// each import by its quoted string (plus an optional `;`), not by the statement
// terminator, so a semicolon-less `import "matrix"` no longer swallows the
// statement after it. Imports of omega or any non-built-in package are kept —
// they need real resolution. Scanning stops at the first non-import token:
// imports sit at the file top by convention, and a stray one lower down is
// left for the parser to reject.
func stripRedundantLeadingImports(source string) string {
	i, n := 0, len(source)
	var kept strings.Builder
	for i < n {
		// Advance j past whitespace and comments (copied through below).
		j := i
		for {
			for j < n && (source[j] == ' ' || source[j] == '\t' || source[j] == '\n' || source[j] == '\r') {
				j++
			}
			if j+1 < n && source[j] == '/' && source[j+1] == '/' {
				if e := strings.IndexByte(source[j:], '\n'); e < 0 {
					j = n
				} else {
					j += e
				}
				continue
			}
			if j+1 < n && source[j] == '/' && source[j+1] == '*' {
				if e := strings.Index(source[j+2:], "*/"); e < 0 {
					j = n
				} else {
					j = j + 2 + e + 2
				}
				continue
			}
			break
		}
		// Require the `import` keyword at a word boundary.
		if !strings.HasPrefix(source[j:], "import") {
			break
		}
		after := j + len("import")
		if after < n && isIdentByte(source[after]) {
			break // an identifier like `importScale`, not the keyword
		}
		// import <ws> "lib"
		k := after
		for k < n && (source[k] == ' ' || source[k] == '\t' || source[k] == '\n' || source[k] == '\r') {
			k++
		}
		if k >= n || source[k] != '"' {
			break // not a string import; leave it for the parser
		}
		q := k + 1
		for q < n && source[q] != '"' {
			q++
		}
		if q >= n {
			break // unterminated string
		}
		lib := source[k+1 : q]
		end := q + 1
		for end < n && (source[end] == ' ' || source[end] == '\t') {
			end++
		}
		if end < n && source[end] == ';' {
			end++
		}
		if isAutoImportable(lib) {
			kept.WriteString(source[i:j]) // keep leading ws/comments, drop the import
			i = end
			continue
		}
		kept.WriteString(source[i:end]) // keep ws/comments and the import verbatim
		i = end
	}
	kept.WriteString(source[i:])
	return kept.String()
}

// referencesMeshAssets reports whether the source mentions any mesh-asset
// construct that needs the canvas's asset registry to build.
func referencesMeshAssets(source string) bool {
	for _, needle := range []string{"importStep", "importMesh", "bvhSdf", "vdbSdf"} {
		if strings.Contains(source, needle) {
			return true
		}
	}
	return false
}

// decodePackedTriangles unpacks the canvas agent's base64+f32 format into
// the flat (verts, indices) layout the Rust oracle FFIs expect — a copy of
// the fidget agent's decoder (internal/agent/packed_triangles.go). Vertices
// come 9 floats per triangle (no shared indexing); indices are the trivial
// 0,1,2,3,... sequence one-per-vertex.
func decodePackedTriangles(verticesB64 string) ([]float32, []uint32, error) {
	raw, err := base64.StdEncoding.DecodeString(verticesB64)
	if err != nil {
		return nil, nil, fmt.Errorf("decode verticesB64: %w", err)
	}
	if len(raw)%4 != 0 {
		return nil, nil, fmt.Errorf("vertex blob length %d not a multiple of 4 (f32)", len(raw))
	}
	floats := make([]float32, len(raw)/4)
	for i := range floats {
		bits := binary.LittleEndian.Uint32(raw[i*4 : i*4+4])
		floats[i] = math.Float32frombits(bits)
	}
	if len(floats)%9 != 0 {
		return nil, nil, fmt.Errorf("vertex count %d not a multiple of 9 floats (3 verts xyz per triangle)", len(floats))
	}
	vertCount := uint32(len(floats) / 3)
	indices := make([]uint32, vertCount)
	for i := range indices {
		indices[i] = uint32(i)
	}
	return floats, indices, nil
}

// gslStmt is one top-level GSL statement plus its leading identifier.
type gslStmt struct {
	text     string // raw text, including leading whitespace and terminator
	name     string // leading identifier ("" if none)
	isImport bool   // leading keyword is "import"
}

// topLevelStatements splits GSL source into top-level statements. A
// statement ends at a top-level ';' (assignments, returns) or at the '}'
// that closes a top-level '{...}' block (function declarations). Comments
// and strings are opaque. Concatenating the statement texts reproduces the
// source exactly, so callers can drop statements and re-emit the rest
// verbatim.
func topLevelStatements(source string) []gslStmt {
	var stmts []gslStmt
	var buf strings.Builder
	flush := func() {
		text := buf.String()
		buf.Reset()
		if strings.TrimSpace(text) == "" {
			// Pure whitespace/comments between statements: attach to the
			// previous statement so nothing is lost on re-emit.
			if text != "" && len(stmts) > 0 {
				stmts[len(stmts)-1].text += text
			} else if text != "" {
				stmts = append(stmts, gslStmt{text: text})
			}
			return
		}
		name, isImport := leadingIdentifier(text)
		stmts = append(stmts, gslStmt{text: text, name: name, isImport: isImport})
	}

	depth := 0
	i, n := 0, len(source)
	for i < n {
		c := source[i]
		switch {
		case c == '/' && i+1 < n && source[i+1] == '/':
			end := strings.IndexByte(source[i:], '\n')
			if end < 0 {
				end = n - i
			}
			buf.WriteString(source[i : i+end])
			i += end
		case c == '/' && i+1 < n && source[i+1] == '*':
			end := strings.Index(source[i+2:], "*/")
			stop := n
			if end >= 0 {
				stop = i + 2 + end + 2
			}
			buf.WriteString(source[i:stop])
			i = stop
		case c == '"':
			j := i + 1
			for j < n && source[j] != '"' {
				if source[j] == '\\' {
					j++
				}
				j++
			}
			if j < n {
				j++
			}
			buf.WriteString(source[i:j])
			i = j
		case c == '{' || c == '[' || c == '(':
			depth++
			buf.WriteByte(c)
			i++
		case c == '}' || c == ']' || c == ')':
			if depth > 0 {
				depth--
			}
			buf.WriteByte(c)
			i++
			if c == '}' && depth == 0 {
				flush()
			}
		case c == ';' && depth == 0:
			buf.WriteByte(c)
			i++
			flush()
		default:
			buf.WriteByte(c)
			i++
		}
	}
	flush()
	return stmts
}

// leadingIdentifier returns the first identifier in a statement (skipping
// leading whitespace and comments) and whether it is the `import` keyword.
func leadingIdentifier(stmt string) (name string, isImport bool) {
	i, n := 0, len(stmt)
	for i < n {
		c := stmt[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i++
		case c == '/' && i+1 < n && stmt[i+1] == '/':
			e := strings.IndexByte(stmt[i:], '\n')
			if e < 0 {
				return "", false
			}
			i += e
		case c == '/' && i+1 < n && stmt[i+1] == '*':
			e := strings.Index(stmt[i+2:], "*/")
			if e < 0 {
				return "", false
			}
			i += 2 + e + 2
		default:
			goto ident
		}
	}
ident:
	start := i
	for i < n {
		c := stmt[i]
		if c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || (i > start && c >= '0' && c <= '9') {
			i++
		} else {
			break
		}
	}
	name = stmt[start:i]
	return name, name == "import"
}

// dedupDefs cleans a defs block for splicing: it drops `import` statements
// (every library is injected automatically) and, when a name is assigned
// more than once, keeps only the last assignment. Returns the surviving
// statements in source order.
func dedupDefs(defs string) []gslStmt {
	if strings.TrimSpace(defs) == "" {
		return nil
	}
	stmts := topLevelStatements(defs)
	lastIdx := make(map[string]int)
	for i, s := range stmts {
		if s.isImport || s.name == "" {
			continue
		}
		lastIdx[s.name] = i
	}
	var out []gslStmt
	for i, s := range stmts {
		if s.isImport {
			continue
		}
		if s.name != "" && lastIdx[s.name] != i {
			continue // superseded by a later assignment of the same name
		}
		if strings.TrimSpace(s.text) == "" {
			continue
		}
		out = append(out, s)
	}
	return out
}

// topLevelNames is the set of names declared at the top level of source.
func topLevelNames(source string) map[string]bool {
	names := make(map[string]bool)
	for _, s := range topLevelStatements(source) {
		if s.name != "" && !s.isImport {
			names[s.name] = true
		}
	}
	return names
}

// substituteTopLevelDecls replaces each top-level declaration whose name is
// in repl with repl[name], in place (preserving declaration order), and
// re-emits the rest verbatim. GSL top-level resolution is order-dependent,
// so an override must sit where the document declared the name rather than
// being appended after the code that uses it.
func substituteTopLevelDecls(source string, repl map[string]string) string {
	var out strings.Builder
	used := make(map[string]bool)
	for _, s := range topLevelStatements(source) {
		if s.name != "" && !used[s.name] {
			if r, ok := repl[s.name]; ok {
				out.WriteString(r)
				used[s.name] = true
				continue
			}
		}
		out.WriteString(s.text)
	}
	return out.String()
}

// stripTopLevelReturns removes top-level `return ...;` statements from GSL
// source so the query's return can take their place (a file may have only
// one return statement). Returns inside function bodies (within braces) are
// kept, as are occurrences inside comments and strings. A `;` inside a
// comment embedded in a return statement would end the strip early — an
// accepted limitation for tool queries.
func stripTopLevelReturns(source string) string {
	var out strings.Builder
	depth := 0
	i := 0
	n := len(source)
	for i < n {
		c := source[i]
		switch {
		case c == '/' && i+1 < n && source[i+1] == '/':
			end := strings.IndexByte(source[i:], '\n')
			if end < 0 {
				end = n - i
			}
			out.WriteString(source[i : i+end])
			i += end
		case c == '/' && i+1 < n && source[i+1] == '*':
			end := strings.Index(source[i+2:], "*/")
			stop := n
			if end >= 0 {
				stop = i + 2 + end + 2
			}
			out.WriteString(source[i:stop])
			i = stop
		case c == '"':
			j := i + 1
			for j < n && source[j] != '"' {
				if source[j] == '\\' {
					j++
				}
				j++
			}
			if j < n {
				j++
			}
			out.WriteString(source[i:j])
			i = j
		case c == '{':
			depth++
			out.WriteByte(c)
			i++
		case c == '}':
			depth--
			out.WriteByte(c)
			i++
		case depth == 0 && c == 'r' && strings.HasPrefix(source[i:], "return") &&
			isWordBoundary(source, i-1) && isWordBoundary(source, i+len("return")):
			end := strings.IndexByte(source[i:], ';')
			if end < 0 {
				i = n
			} else {
				i += end + 1
			}
		default:
			out.WriteByte(c)
			i++
		}
	}
	return out.String()
}

// isWordBoundary reports whether the byte at index i (out of range counts
// as a boundary) cannot be part of an identifier.
func isWordBoundary(source string, i int) bool {
	if i < 0 || i >= len(source) {
		return true
	}
	c := source[i]
	return !(c == '_' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z')
}
