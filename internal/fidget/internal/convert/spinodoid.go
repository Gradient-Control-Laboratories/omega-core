package convert

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"sync"

	"google.golang.org/protobuf/proto"

	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/internal/fidget/internal/fidget"
	tapepb "gradientcontrol.com/omega/core/internal/fidget/internal/tape"
	"gradientcontrol.com/omega/core/internal/noise/spinodoid"
	"gradientcontrol.com/omega/core/program"
)

// builtinGaborSpinodoid lowers gaborSpinodoid / gaborSpinodoidGradient and
// their fastGaborSpinodoid aliases to the spinodoid oracle leaves over a
// lattice kernel table baked here, in the fidget agent: the drivers are fields
// of p, and this is where a field can be evaluated at the kernel centres, once
// each, as hodge samples them.
//
// Both names are the lattice field, here as in the preview, so a part meshes
// as it renders. A tape marks every leaf fast; a reader that wants hodge's
// field swaps it in from the payload's arguments.
//
// Every argument but q, scale and orientation must be a constant by now
// (inlining has folded the library wrapper's parameters into the call); they
// are the table's key. A constant scale of 1 and a constant identity
// orientation are no driver at all, which is what hodge's None means, and bake
// the same table bit for bit.
//
// Slots are handed out per build: every tree is cleared when the build ends,
// so a slot never outlives the trees that read it.
func (d *Delegate) builtinGaborSpinodoid(
	op *program.Operand, call *program.FnCall, args []build.Expression,
) (build.Expression, error) {
	name := call.Reference.Name
	trees := make([][]fidget.TreeIndex, len(args))
	for i, a := range args {
		trees[i] = a.(*TreeNode).trees
	}
	q := trees[0]
	if len(q) != 3 {
		return nil, fmt.Errorf("%s: expected a 3-component point, got %d components", name, len(q))
	}

	constant := func(i int, what string) ([]float64, error) {
		vals := make([]float64, len(trees[i]))
		for j, t := range trees[i] {
			v, ok := fidget.ConstValue(t)
			if !ok {
				return nil, fmt.Errorf("%s: %s must be a constant: it keys the baked kernel table", name, what)
			}
			vals[j] = v
		}
		return vals, nil
	}
	scalar := func(i int, what string) (float64, error) {
		v, err := constant(i, what)
		if err != nil {
			return 0, err
		}
		if len(v) != 1 {
			return 0, fmt.Errorf("%s: %s must be a scalar", name, what)
		}
		return v[0], nil
	}
	integer := func(i int, what string) (int, error) {
		v, err := scalar(i, what)
		if err != nil {
			return 0, err
		}
		if v != math.Trunc(v) {
			return 0, fmt.Errorf("%s: %s must be a whole number, got %g", name, what, v)
		}
		return int(v), nil
	}

	var p spinodoid.Params
	var err error
	if p.Period, err = scalar(1, "period"); err != nil {
		return nil, err
	}
	if p.Theta, err = constant(2, "theta"); err != nil {
		return nil, err
	}
	poles, err := constant(3, "poles")
	if err != nil {
		return nil, err
	}
	if len(poles) != 3*len(p.Theta) {
		return nil, fmt.Errorf("%s: %d cone angles need %d pole components, got %d",
			name, len(p.Theta), 3*len(p.Theta), len(poles))
	}
	for i := 0; i < len(poles); i += 3 {
		p.Poles = append(p.Poles, [3]float64{poles[i], poles[i+1], poles[i+2]})
	}
	if p.PacketWidth, err = scalar(4, "packetWidth"); err != nil {
		return nil, err
	}
	if p.Impulses, err = integer(5, "impulses"); err != nil {
		return nil, err
	}
	if p.Pool, err = integer(6, "pool"); err != nil {
		return nil, err
	}
	seed, err := integer(7, "seed")
	if err != nil {
		return nil, err
	}
	p.Seed = int64(seed)
	// The bias is the level itself (spinodoid.CheckBias).
	bias, err := scalar(8, "bias")
	if err != nil {
		return nil, err
	}
	level := bias
	if err := spinodoid.CheckBias(level); err != nil {
		return nil, fmt.Errorf("%s: %v", name, err)
	}
	localFlag, err := scalar(9, "localNormalize")
	if err != nil {
		return nil, err
	}
	local := localFlag != 0

	// drivers collects the trees of the drivers that are fields, for the
	// fingerprint that lets a driven table be reused while they are unchanged.
	var drivers []fidget.TreeIndex
	scale, rot := 1.0, identity3
	// Whether each driver is a field rather than a constant: what decides
	// whether a tape can carry its value (see recordSpinodoidOracle). `scale`
	// and `rot` are only meaningful when the corresponding flag is false,
	// since the lookups below assign into them either way.
	var scaleIsField, rotIsField bool
	var ok bool
	if scale, ok = spinodoidScaleDriver(trees[10]); !ok {
		scaleIsField = true
		p.Scale = evalScalarDriver(trees[10][0])
		drivers = append(drivers, trees[10]...)
	} else if scale != 1 {
		p.Scale = func(centers [][3]float64) ([]float64, error) {
			out := make([]float64, len(centers))
			for i := range out {
				out[i] = scale
			}
			return out, nil
		}
	}
	if len(trees[11]) != 9 {
		return nil, fmt.Errorf("%s: orientation must be a 3x3 matrix, got %d components", name, len(trees[11]))
	}
	if rot, ok = spinodoidOrientationDriver(trees[11]); !ok {
		rotIsField = true
		p.Orientation = evalRotationDriver(trees[11])
		drivers = append(drivers, trees[11]...)
	} else if rot != identity3 {
		p.Orientation = func(centers [][3]float64) ([][3][3]float64, error) {
			out := make([][3][3]float64, len(centers))
			for i := range out {
				out[i] = rot
			}
			return out, nil
		}
	}

	// The key spells out everything the table depends on (p itself holds
	// closures, which do not print as their values): the constants, and for a
	// field driver a fingerprint of its tree, its serialized tape. So a driven
	// table is reused, not re-sampled, while its drivers are unchanged, and the
	// key names the table's content for a preview that keeps tables it has
	// already been sent.
	//
	// The same bytes are how a field driver travels in a tape export: the
	// trees are only live during this build, so they are serialized here and
	// spliced into the export later (see mesh.SpinodoidPayloads). Their outputs
	// are the scale tree, if a field, then the nine orientation trees, if a
	// field, in the column-major order the rotation's components arrive in.
	fingerprint := ""
	var driverTape []byte
	if len(drivers) > 0 {
		driverTape, err = fidget.SerializeTape(drivers)
		if err != nil {
			return nil, fmt.Errorf("%s: fingerprint drivers: %v", name, err)
		}
		sum := sha256.Sum256(driverTape)
		fingerprint = hex.EncodeToString(sum[:])
	}
	p.Periodic = len(drivers) == 0
	p.Domain = d.spinodoidDomain
	if !p.Periodic && p.Domain != nil {
		// A direct gaborSpinodoid(q, ...) call may translate/scale its query.
		// Bake the conservative q-space image of the known world-space domain.
		p.Domain, err = spinodoidQueryDomain(q, *p.Domain)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
	}
	layout, err := spinodoid.NewLayout(spinodoid.TableConstants(p.Period, p.PacketWidth, p.Impulses), p.Domain, p.Periodic)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	key := fmt.Sprintf("domain-v2|%v|%v|%v|%v|%d|%d|%d|%v|%v|%s|%v", p.Period, p.Theta, p.Poles,
		p.PacketWidth, p.Impulses, p.Pool, p.Seed, scale, rot, fingerprint, layout)
	table, err := bakeSpinodoid(key, p)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", name, err)
	}
	contentSum := sha256.Sum256([]byte(key))

	// Record the bake under the call's cross-backend key, for a WGSL preview
	// asking which table each of its calls reads (see BakeSpinodoids).
	callKey := spinodoid.Call{
		Period: p.Period, Theta: p.Theta, Poles: poles, PacketWidth: p.PacketWidth,
		Impulses: p.Impulses, Pool: p.Pool, Seed: p.Seed, Bias: bias, Local: local,
		Fast:  true,
		Chain: build.KeyedCallSite(d.GetScope(), op),
	}.Key()
	if d.spinodoidBakes == nil {
		d.spinodoidBakes = map[string]BakedSpinodoid{}
	}

	slotKey := fmt.Sprintf("%p/%v/%v", table, level, local)
	slot, ok := d.spinodoidSlots[slotKey]
	if !ok {
		if d.spinodoidSlots == nil {
			d.spinodoidSlots = map[string]uint32{}
		}
		slot = d.spinodoidNextSlot
		d.spinodoidNextSlot++
		if err := fidget.RegisterSpinodoid(slot, table, level, local); err != nil {
			return nil, fmt.Errorf("%s: %v", name, err)
		}
		d.spinodoidSlots[slotKey] = slot
		d.recordSpinodoidOracle(fidget.SpinodoidOracleID(slot, fidget.SpinodoidDistance),
			p, poles, bias, local, scale, scaleIsField, rot, rotIsField, fingerprint, driverTape)
	}
	d.spinodoidBakes[callKey] = BakedSpinodoid{Table: table, ContentID: hex.EncodeToString(contentSum[:16]),
		params: p, cacheKey: key, slot: slot, level: level, local: local}
	return d.spinodoidLeaves(op, name, slot, q, fidget.SpinodoidTree)
}

// spinodoidLeaves builds the oracle leaves for one call: the distance, or the
// three components of the gradient, remapped onto the query point.
func (d *Delegate) spinodoidLeaves(op *program.Operand, name string, slot uint32,
	q []fidget.TreeIndex, tree func(slot, comp uint32) fidget.TreeIndex) (build.Expression, error) {
	comps := []uint32{fidget.SpinodoidDistance}
	if program.IsGaborSpinodoidGradient(name) {
		comps = []uint32{fidget.SpinodoidGradX, fidget.SpinodoidGradX + 1, fidget.SpinodoidGradX + 2}
	}
	leaves := make([]fidget.TreeIndex, len(comps))
	for i, comp := range comps {
		leaf := tree(slot, comp)
		if leaf == ^fidget.TreeIndex(0) {
			return nil, fmt.Errorf("%s: spinodoid slot %d is not registered", name, slot)
		}
		leaves[i] = fidget.Remap(leaf, q[0], q[1], q[2])
	}
	return NewTreeNode(op, op.Start, leaves...)
}

var identity3 = [3][3]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}

// spinodoidScaleDriver reports the scale when it is a constant.
func spinodoidScaleDriver(trees []fidget.TreeIndex) (float64, bool) {
	if len(trees) != 1 {
		return 0, false
	}
	return fidget.ConstValue(trees[0])
}

// spinodoidOrientationDriver reports the rotation when every component is a
// constant. A GSL matrix is a list of column vectors, so component 3j+i is
// row i, column j.
func spinodoidOrientationDriver(trees []fidget.TreeIndex) ([3][3]float64, bool) {
	var r [3][3]float64
	for j := range 3 {
		for i := range 3 {
			v, ok := fidget.ConstValue(trees[3*j+i])
			if !ok {
				return r, false
			}
			r[i][j] = v
		}
	}
	return r, true
}

// evalField evaluates a field tree of the implicit (x, y, z) at points, in
// one batch: hodge calls its drivers with every kernel centre at once, and
// so does the bake.
func evalField(tree fidget.TreeIndex, xs, ys, zs []float32) ([]float64, error) {
	if v, ok := fidget.ConstValue(tree); ok {
		out := make([]float64, len(xs))
		for i := range out {
			out[i] = v
		}
		return out, nil
	}
	res, err := fidget.EvalGrad(fidget.BuildCanvas(tree), xs, ys, zs)
	if err != nil {
		return nil, err
	}
	out := make([]float64, len(res))
	for i, r := range res {
		out[i] = float64(r.Value)
	}
	return out, nil
}

func splitCenters(centers [][3]float64) (xs, ys, zs []float32) {
	xs = make([]float32, len(centers))
	ys = make([]float32, len(centers))
	zs = make([]float32, len(centers))
	for i, c := range centers {
		xs[i], ys[i], zs[i] = float32(c[0]), float32(c[1]), float32(c[2])
	}
	return xs, ys, zs
}

func evalScalarDriver(tree fidget.TreeIndex) func([][3]float64) ([]float64, error) {
	return func(centers [][3]float64) ([]float64, error) {
		xs, ys, zs := splitCenters(centers)
		return evalField(tree, xs, ys, zs)
	}
}

func evalRotationDriver(trees []fidget.TreeIndex) func([][3]float64) ([][3][3]float64, error) {
	return func(centers [][3]float64) ([][3][3]float64, error) {
		xs, ys, zs := splitCenters(centers)
		out := make([][3][3]float64, len(centers))
		for j := range 3 {
			for i := range 3 {
				vals, err := evalField(trees[3*j+i], xs, ys, zs)
				if err != nil {
					return nil, err
				}
				for k, v := range vals {
					out[k][i][j] = v
				}
			}
		}
		return out, nil
	}
}

// Tables are kept across builds by their key (parameters plus a fingerprint
// of any field driver). A handful is plenty: each is ~16 MB.
const spinodoidCacheSize = 4

var (
	spinodoidCacheMu sync.Mutex
	spinodoidCache   = map[string]*spinodoid.Table{}
)

func bakeSpinodoid(key string, p spinodoid.Params) (*spinodoid.Table, error) {
	spinodoidCacheMu.Lock()
	defer spinodoidCacheMu.Unlock()
	if t, ok := spinodoidCache[key]; ok {
		return t, nil
	}
	t, err := spinodoid.Bake(p)
	if err != nil {
		return nil, err
	}
	if len(spinodoidCache) >= spinodoidCacheSize {
		clear(spinodoidCache)
	}
	spinodoidCache[key] = t
	return t, nil
}

// recordSpinodoidOracle notes the arguments of a table just bound to a slot,
// for a tape export to carry. A constant driver travels as its value; a field
// driver leaves the value nil and travels as its trees, serialized in
// driverTape (the same bytes the fingerprint hashes).
func (d *Delegate) recordSpinodoidOracle(assetID uint32, p spinodoid.Params,
	poles []float64, bias float64, local bool,
	scale float64, scaleIsField bool, rot [3][3]float64, rotIsField bool,
	fingerprint string, driverTape []byte) {
	o := SpinodoidOracle{
		AssetID: assetID, Fast: true,
		Period: p.Period, Theta: p.Theta, Poles: poles,
		PacketWidth: p.PacketWidth, Impulses: uint32(p.Impulses), Pool: uint32(p.Pool),
		Seed: p.Seed, Bias: bias, LocalNormalize: local,
		DriverFingerprint: fingerprint, DriverTape: driverTape, Periodic: p.Periodic,
	}
	if !scaleIsField {
		o.Scale = []float64{scale}
	}
	if !rotIsField {
		o.Orientation = []float64{
			rot[0][0], rot[0][1], rot[0][2],
			rot[1][0], rot[1][1], rot[1][2],
			rot[2][0], rot[2][1], rot[2][2],
		}
	}
	if p.Domain != nil {
		o.Domain = p.Domain[:]
	}
	d.spinodoidOracles = append(d.spinodoidOracles, o)
}

// SpinodoidOracles returns the spinodoid tables bound this build, for a tape
// export to stamp into tape.oracles.
func (d *Delegate) SpinodoidOracles() []SpinodoidOracle { return d.spinodoidOracles }

// SpinodoidOracle describes one spinodoid table bound this build: the leaf id
// of its slot's distance component, and the arguments the field is a function
// of. A tape export stamps these into tape.oracles (tape.proto's
// SpinodoidParams) so a consumer can rebuild the field — the slot number alone
// means nothing in another process.
//
// The arguments travel, not the table: omega and hodge both compute the field
// from them and are pinned to each other, so a consumer rebuilding it gets a
// checkable agreement rather than an opaque blob.
type SpinodoidOracle struct {
	// AssetID is the leaf id of component 0 (distance). The slot's other three
	// components are the next three ids; a consumer registers the slot once and
	// resolves all four, so a gradient-only call is covered too.
	AssetID uint32
	// Fast marks the lattice field rather than hodge's. Every table omega
	// binds is the lattice, whichever builtin the document called.
	Fast bool

	Period         float64
	Theta          []float64 // radians
	Poles          []float64 // flat xyz, 3 per theta
	PacketWidth    float64
	Impulses       uint32
	Pool           uint32
	Seed           int64
	Bias           float64 // the level itself, in field standard deviations
	LocalNormalize bool

	// Constant drivers, nil when the driver is a field. Orientation is
	// row-major.
	Scale       []float64
	Orientation []float64
	// DriverFingerprint identifies the field drivers omega used, empty when
	// both are constants.
	DriverFingerprint string
	// DriverTape is the field drivers serialized as a tape of their own: one
	// output for the scale if Scale is nil, then nine for the orientation if
	// Orientation is nil, column-major (the order a GSL matrix's components
	// arrive in). mesh.SpinodoidPayloads splices it into the export.
	DriverTape []byte

	Periodic bool
	Domain   []float64 // 6 entries when bounded, nil otherwise
}

// BakeSpinodoids builds prog and returns the spinodoid table each of its
// gaborSpinodoid calls reads, by spinodoid.Call key: what a WGSL preview,
// which cannot sample drivers itself, needs to bind. The build is the same
// one meshing does, so the keys and tables are the ones the mesher uses.
// BakedSpinodoid is a baked table and the id of its content: tables with the
// same ContentID are identical, whichever call they serve.
type BakedSpinodoid struct {
	Table     *spinodoid.Table
	ContentID string
	params    spinodoid.Params
	cacheKey  string
	slot      uint32
	level     float64
	local     bool
}

func spinodoidQueryDomain(q []fidget.TreeIndex, world [6]float64) (*[6]float64, error) {
	var box [6]float32
	for i, v := range world {
		box[i] = float32(v)
		// Widen conversion instead of excluding points at a rounded boundary.
		if i < 3 {
			box[i] = math.Nextafter32(box[i], float32(math.Inf(-1)))
		} else {
			box[i] = math.Nextafter32(box[i], float32(math.Inf(1)))
		}
	}
	var domain [6]float64
	for axis := range 3 {
		values, err := fidget.EvalInterval(fidget.BuildCanvas(q[axis]), [][6]float32{box})
		if err != nil {
			return nil, fmt.Errorf("bound spinodoid query: %w", err)
		}
		lo, hi := float64(values[0].Lower), float64(values[0].Upper)
		if math.IsNaN(lo) || math.IsNaN(hi) || math.IsInf(lo, 0) || math.IsInf(hi, 0) {
			return nil, fmt.Errorf("spinodoid query has no finite domain; use a bounded query transform")
		}
		if lo == hi {
			lo, hi = lo-1e-6, hi+1e-6
		}
		domain[axis], domain[axis+3] = lo, hi
	}
	return &domain, nil
}

// Outer remap is applied after a field has been built; its full composed
// coordinate domain is not available when that field's kernels are baked.
// Preserve legacy behavior for only the actual spinodoid leaves being remapped,
// instead of clipping them or disabling finite grading for unrelated material
// library remaps. Imported default projects use the direct world-space field.
func (d *Delegate) preserveRemappedSpinodoids(trees []fidget.TreeIndex) error {
	if len(d.spinodoidBakes) == 0 {
		return nil
	}
	data, err := fidget.SerializeTape(trees)
	if err != nil {
		return err
	}
	var tape tapepb.Tape
	if err := proto.Unmarshal(data, &tape); err != nil {
		return err
	}
	const oracleBase = uint32(0xE05B1000) // tape-core spinodoid::SPINODOID_ORACLE_ID_BASE
	slots := map[uint32]bool{}
	for _, node := range tape.Nodes {
		if leaf, ok := node.N.(*tapepb.Node_MeshOracle); ok && leaf.MeshOracle >= oracleBase && leaf.MeshOracle < oracleBase+256*4 {
			slots[(leaf.MeshOracle-oracleBase)/4] = true
		}
	}
	for callKey, baked := range d.spinodoidBakes {
		if !slots[baked.slot] || baked.Table.Periodic {
			continue
		}
		baked.params.Periodic, baked.params.Domain = true, nil
		baked.cacheKey += "|outer-remap-legacy-periodic"
		baked.Table, err = bakeSpinodoid(baked.cacheKey, baked.params)
		if err != nil {
			return err
		}
		sum := sha256.Sum256([]byte(baked.cacheKey))
		baked.ContentID = hex.EncodeToString(sum[:16])
		if err := fidget.RegisterSpinodoid(baked.slot, baked.Table, baked.level, baked.local); err != nil {
			return err
		}
		d.spinodoidBakes[callKey] = baked
		for key, slot := range d.spinodoidSlots {
			if slot == baked.slot {
				delete(d.spinodoidSlots, key)
			}
		}
	}
	return nil
}

func BakeSpinodoids(prog *program.Program) (map[string]BakedSpinodoid, error) {
	if prog == nil || prog.Primary == nil {
		return nil, fmt.Errorf("program or primary package is nil")
	}
	var gslSource string
	if len(prog.Primary.Files) > 0 {
		gslSource = prog.Primary.Files[0].Source
	}
	d := NewDelegate(gslSource)

	fidget.Lock()
	defer fidget.Unlock()
	defer fidget.Clear()

	if err := build.Build(d, prog, true); err != nil {
		return nil, build.ErrWithPkgSrc(prog, err)
	}
	return d.spinodoidBakes, nil
}
