package spinodoid

import (
	"fmt"
	"math"
)

// hodge's Gabor spinodoid (hodge-release spinodoid.py, spinodoid_gabor)
// exactly: kernels jittered inside the cells of its grid, its untapered
// envelope, its SplitMix64 hash, and its drivers read once at every kernel
// centre. A point agrees with hodge's Python to rounding (hodge_parity_test),
// so a part baked from this is a part hodge can rebuild from the
// same arguments — which is the whole reason to pay for it. The lattice form
// beside it (kernels.go) is the same model, cheaper, and a different
// realization.
//
// It shares Params, Level, the pool and the domain rules with the lattice
// baker, but not its layout. A stationary field is the periodic window hodge's
// own grid describes, 32^3 cells about the world origin. A driven field with a
// Domain is baked over the cells covering it, hashed on true cell indices, so
// it is hodge's field everywhere inside it; one without a Domain falls back to
// the periodic window, and its drivers repeat with it.

// HodgeTableCells is the table's edge in kernel cells: 32^3 cells, about 116
// wavelengths at the default packet width, before the field repeats.
const HodgeTableCells = 32

// hodgeTableOrigin is the lowest cell index the table holds on each axis,
// centring it on the world origin.
const hodgeTableOrigin = -HodgeTableCells / 2

// gaborOffsets is hodge's _GABOR_OFFSETS: the 27-cell neighbourhood with the
// LAST axis fastest, the order hodge accumulates kernels in.
var gaborOffsets = func() [27][3]int64 {
	var o [27][3]int64
	n := 0
	for i := int64(-1); i <= 1; i++ {
		for j := int64(-1); j <= 1; j++ {
			for k := int64(-1); k <= 1; k++ {
				o[n] = [3]int64{i, j, k}
				n++
			}
		}
	}
	return o
}()

// HodgeConstants are a hodge table's scalar constants, which depend only on
// the period, packet width and kernel density.
type HodgeConstants struct {
	// CellSize is the kernel cell's edge, the envelope's reach to the cutoff.
	CellSize float64
	// A is the envelope parameter 1 / (PacketWidth * Period), the same for every
	// kernel: hodge decouples the envelope from the local wavelength.
	A float64
	// Sigma is the unit-variance normalization of the kernel sum.
	Sigma float64
	// BaseBeta is the wavenumber of Period, the fallback where no kernel reaches.
	BaseBeta float64
}

// HodgeTableConstants computes hodge's kernel-cell size (the envelope's reach
// to its cutoff), envelope parameter, unit-variance normalization and base
// wavenumber.
func HodgeTableConstants(period, packetWidth float64, impulses int) HodgeConstants {
	a0 := 1.0 / (packetWidth * period)
	cellSize := math.Sqrt(-math.Log(envelopeCutoff)/pi) / a0
	rhoK := float64(impulses) / (cellSize * cellSize * cellSize)
	return HodgeConstants{
		CellSize: cellSize,
		A:        a0,
		Sigma:    math.Sqrt(0.5 * rhoK * math.Pow(2.0*a0*a0, -1.5)),
		BaseBeta: 2.0 * pi / period,
	}
}

// Kernel is one baked Gabor kernel.
type Kernel struct {
	// Jitter places the centre inside its cell: centre = (cell + Jitter) * CellSize.
	Jitter [3]float64
	// Dir is the unit wave direction, after Orientation.
	Dir [3]float64
	// Beta is the wavenumber 2 pi / (Period * scale).
	Beta float64
	// Gamma is the phase.
	Gamma float64
}

// HodgeTable is a baked Gabor spinodoid, hodge's exactly.
type HodgeTable struct {
	HodgeConstants
	// Layout is which cells the table holds: hodge's periodic 32^3 window for
	// a stationary field, or the cells covering a driven field's domain.
	Layout
	Impulses int
	// Kernels are indexed ((z*Size[1] + y)*Size[0] + x)*Impulses + k, for cell
	// (x, y, z) - Origin.
	Kernels []Kernel
}

// newHodgeLayout is NewLayout for hodge's grid: cells rather than lattice
// sites, and a one-cell pad rather than a Reach, because a query reads exactly
// the 27 cells around its own (hodge's _GABOR_OFFSETS).
//
// Unlike the lattice's, a bounded hodge table hashes the TRUE cell index --
// hodge's realization IS the hash of the true cell, so wrapping it to keep a
// world-anchored realization (as the lattice does) would stop the field being
// hodge's. The consequence is the right one: inside the domain this table is
// hodge's everywhere, with nothing repeating.
func newHodgeLayout(c HodgeConstants, domain *[6]float64, periodic bool) (Layout, error) {
	if periodic || domain == nil {
		return Layout{
			Origin:   [3]int{hodgeTableOrigin, hodgeTableOrigin, hodgeTableOrigin},
			Size:     [3]int{HodgeTableCells, HodgeTableCells, HodgeTableCells},
			Periodic: true,
		}, nil
	}
	l := Layout{Bounds: *domain}
	for a := range 3 {
		lo, hi := domain[a], domain[a+3]
		if math.IsNaN(lo) || math.IsNaN(hi) || math.IsInf(lo, 0) || math.IsInf(hi, 0) || hi <= lo {
			return Layout{}, fmt.Errorf("spinodoid domain must have finite increasing bounds on every axis")
		}
		first := math.Floor(lo/c.CellSize) - 1
		last := math.Floor(hi/c.CellSize) + 1
		if math.Abs(first) > 1<<29 || math.Abs(last) > 1<<29 || last-first+1 > MaxTableSites {
			return Layout{}, fmt.Errorf("graded spinodoid domain is too large; reduce the scene/domain extent or increase the base period")
		}
		l.Origin[a] = int(first)
		l.Size[a] = int(last-first) + 1
	}
	return l, nil
}

// HodgeBake builds hodge's table for p.
func HodgeBake(p Params) (*HodgeTable, error) {
	if !(p.Period > 0) {
		return nil, fmt.Errorf("period must be > 0, got %v", p.Period)
	}
	if !(p.PacketWidth > 0) {
		return nil, fmt.Errorf("packet_width must be > 0, got %v", p.PacketWidth)
	}
	if p.Impulses < 1 {
		return nil, fmt.Errorf("impulses must be >= 1, got %d", p.Impulses)
	}
	if p.Seed < 0 {
		return nil, fmt.Errorf("seed must be >= 0, got %d", p.Seed)
	}
	pool := p.Pool
	if pool == 0 {
		pool = defaultPool
	}
	// The pool is hodge's draw, in radians as hodge takes them. Allow pi/2 to
	// arrive a float32 rounding high, and clamp it back.
	theta := make([]float64, len(p.Theta))
	for i, t := range p.Theta {
		if !(t >= 0 && t <= pi/2+1e-6) {
			return nil, fmt.Errorf("theta must lie in [0, pi/2] radians, got %v", t)
		}
		theta[i] = min(t, pi/2)
	}
	baseDirs, err := Waves(theta, pool, uint64(p.Seed), p.Poles)
	if err != nil {
		return nil, err
	}

	twoPi := 2.0 * pi
	constants := HodgeTableConstants(p.Period, p.PacketWidth, p.Impulses)
	layout, err := newHodgeLayout(constants, p.Domain, p.Periodic || (p.Scale == nil && p.Orientation == nil))
	if err != nil {
		return nil, err
	}
	if layout.Count() > MaxTableSites/p.Impulses {
		return nil, fmt.Errorf("a graded hodge spinodoid over this domain needs %d kernels, past the %d limit; "+
			"reduce the domain extent, raise the base period, or lower impulses", layout.Count()*p.Impulses, MaxTableSites)
	}
	t := &HodgeTable{
		HodgeConstants: constants,
		Layout:         layout,
		Impulses:       p.Impulses,
		Kernels:        make([]Kernel, layout.Count()*p.Impulses),
	}
	cellSize := t.CellSize

	// First the hashed part of every kernel, in table order: jitter, pool
	// direction and phase.
	n := len(t.Kernels)
	centers := make([][3]float64, n)
	dirs := make([][3]float64, n)
	for z := range t.Size[2] {
		for y := range t.Size[1] {
			for x := range t.Size[0] {
				cell := [3]int64{int64(x + t.Origin[0]), int64(y + t.Origin[1]), int64(z + t.Origin[2])}
				base := ((z*t.Size[1]+y)*t.Size[0] + x) * p.Impulses
				for k := range p.Impulses {
					stream := 8 * k
					kn := &t.Kernels[base+k]
					for a := range 3 {
						kn.Jitter[a] = cellUniform(cell, stream+a, p.Seed)
						centers[base+k][a] = (float64(cell[a]) + kn.Jitter[a]) * cellSize
					}
					idx := int64(cellUniform(cell, stream+3, p.Seed) * float64(len(baseDirs)))
					idx = min(max(idx, 0), int64(len(baseDirs)-1))
					kn.Gamma = cellUniform(cell, stream+4, p.Seed) * twoPi
					dirs[base+k] = baseDirs[idx]
				}
			}
		}
	}

	// Then the drivers, once per kernel centre, as hodge calls them.
	var rots [][3][3]float64
	if p.Orientation != nil {
		if rots, err = p.Orientation(centers); err != nil {
			return nil, err
		}
		if len(rots) != n {
			return nil, fmt.Errorf("orientation returned %d rotations for %d kernel centres", len(rots), n)
		}
	}
	var scales []float64
	if p.Scale != nil {
		if scales, err = p.Scale(centers); err != nil {
			return nil, err
		}
		if len(scales) != n {
			return nil, fmt.Errorf("scale returned %d values for %d kernel centres", len(scales), n)
		}
	}
	for i := range t.Kernels {
		kn := &t.Kernels[i]
		dir := dirs[i]
		if rots != nil {
			r := rots[i]
			dir = [3]float64{dot3(r[0], dir), dot3(r[1], dir), dot3(r[2], dir)}
		}
		kn.Dir = normalize(dir)
		s := 1.0
		if scales != nil {
			s = scales[i]
			if !(s > 0) {
				return nil, fmt.Errorf("scale must be > 0, got %v at %v", s, centers[i])
			}
		}
		kn.Beta = twoPi / (p.Period * s)
	}
	return t, nil
}

// KernelStride is how many float64s Flat writes per kernel.
const KernelStride = 8

// Flat is the table's kernels as consecutive float64s, KernelStride per
// kernel in Kernels order: jitter xyz, direction xyz, wavenumber, phase.
func (t *HodgeTable) Flat() []float64 {
	out := make([]float64, 0, len(t.Kernels)*KernelStride)
	for _, k := range t.Kernels {
		out = append(out, k.Jitter[0], k.Jitter[1], k.Jitter[2],
			k.Dir[0], k.Dir[1], k.Dir[2], k.Beta, k.Gamma)
	}
	return out
}

// wrapCell maps an unwrapped cell index to the periodic window's
// [0, HodgeTableCells).
func wrapCell(c int64) int {
	m := (c - hodgeTableOrigin) % HodgeTableCells
	if m < 0 {
		m += HodgeTableCells
	}
	return int(m)
}

// kernelsOf is the kernels of cell c, or false where a bounded table does not
// hold it. A periodic table wraps, which is what makes a stationary field tile.
func (t *HodgeTable) kernelsOf(c [3]int64) ([]Kernel, bool) {
	var idx [3]int
	for a := range 3 {
		if t.Periodic {
			idx[a] = wrapCell(c[a])
			continue
		}
		i := int(c[a]) - t.Origin[a]
		if i < 0 || i >= t.Size[a] {
			return nil, false
		}
		idx[a] = i
	}
	base := ((idx[2]*t.Size[1]+idx[1])*t.Size[0] + idx[0]) * t.Impulses
	return t.Kernels[base : base+t.Impulses], true
}

// Sample is the raw Gabor sum at q: field is the kernel sum divided by Sigma
// (unit variance), and beta is the envelope-energy-weighted wavenumber the
// field carries there (hodge's normalize="local"), BaseBeta where no kernel
// reaches.
func (t *HodgeTable) Sample(q [3]float64) (field, beta float64) {
	// Outside a bounded domain the field has no kernels, so carry the distance
	// to the domain instead: the same conservative extension the lattice uses,
	// which lets a ray marcher reach the part rather than step over it.
	if !t.Contains(q) {
		d, _ := t.Outside(q)
		return d * t.BaseBeta, t.BaseBeta
	}
	var cell [3]int64
	for a := range 3 {
		cell[a] = int64(math.Floor(q[a] / t.CellSize))
	}
	a2 := t.A * t.A
	var acc, wAcc, bwAcc float64
	for _, off := range gaborOffsets {
		c := [3]int64{cell[0] + off[0], cell[1] + off[1], cell[2] + off[2]}
		kernels, ok := t.kernelsOf(c)
		if !ok {
			continue
		}
		for k := range kernels {
			kn := &kernels[k]
			var r [3]float64
			for a := range 3 {
				r[a] = q[a] - (float64(c[a])+kn.Jitter[a])*t.CellSize
			}
			env := math.Exp(-pi * a2 * dot3(r, r))
			phase := kn.Beta*dot3(r, kn.Dir) + kn.Gamma
			acc += env * math.Cos(phase)
			w := env * env
			wAcc += w
			bwAcc += w * kn.Beta
		}
	}
	beta = bwAcc / math.Max(wAcc, 1e-300)
	if !(beta > 0) {
		beta = t.BaseBeta
	}
	return acc / t.Sigma, beta
}

// Distance is spinodoid_gabor's result without wall_thickness: the field
// leveled at level and divided by the local wavenumber (local true) or by
// BaseBeta (hodge's normalize="base"). Negative in the solid phase. Outside a
// bounded domain it is the distance to the domain, unleveled, as the lattice
// table's Distance and the oracle are.
func (t *HodgeTable) Distance(q [3]float64, level float64, local bool) float64 {
	if !t.Contains(q) {
		d, _ := t.Outside(q)
		return d
	}
	field, beta := t.Sample(q)
	if !local {
		beta = t.BaseBeta
	}
	return (field - level) / beta
}
