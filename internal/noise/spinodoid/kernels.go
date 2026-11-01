// Package spinodoid bakes a lattice Gabor spinodoid for omega's backends.
//
// It began as an exact port of hodge's spinodoid_gabor (hodge-release
// spinodoid.py) and keeps its statistical model: wave packets of one
// wavelength under a Gaussian envelope of fixed physical size, directions
// drawn from hodge's cone orientation distribution (the same numpy-drawn
// pool), and scale and orientation drivers sampled once per kernel centre, so
// each packet is an honest plane wave. Two things differ, both for speed:
//
//   - Kernels sit on a regular lattice instead of jittered in hodge's cells, so
//     where they are is arithmetic: an evaluation visits only the sites within
//     reach and loads data only for those, instead of scanning 27 cells of
//     kernels to find the few that matter. Each site still has its own random
//     direction and phase, and at the lattice spacing used (1.5 envelope
//     standard deviations at hodge's density) the lattice does not print
//     through the variance.
//   - Each kernel's envelope is tapered to zero at hodge's cutoff radius,
//     exp(-pi a^2 r^2) (1 - r^2/R^2)^2, so a kernel enters and leaves the sum
//     smoothly. The field is smooth everywhere (hodge's has small seams at its
//     cell faces), and the reach is part of the definition, not a culling
//     approximation, so every backend computes the same field.
//
// The integer work (the hash, the direction pool) happens here, once, and the
// result is a table the backends read with float-only work. Stationary tables
// repeat their random kernel realization; driven tables sample actual world
// centres over a finite domain so scale and orientation never repeat by accident.
package spinodoid

import (
	"fmt"
	"math"
)

// TableSites is the table's edge in lattice sites. At hodge's density (8
// kernels per cutoff cell) that is 32 cutoff cells, about 116 wavelengths at
// the default packet width, before the field repeats.
const TableSites = 64

// tableOrigin is the lowest site index the table holds on each axis, centring
// it on the world origin.
const tableOrigin = -TableSites / 2

const (
	// envelopeCutoff is hodge's: the envelope value at the kernel's reach R.
	envelopeCutoff = 0.01
	defaultPool    = 4096
)

const pi = math.Pi

// Params are a spinodoid's arguments.
type Params struct {
	// Theta are the cone half-angles, one per pole, in radians.
	Theta []float64
	// Poles are the cone axes; nil is the Cartesian basis.
	Poles [][3]float64
	// Period is the reference wavelength; Scale multiplies it locally.
	Period float64
	// PacketWidth is the envelope width in wavelengths at scale 1 (hodge: 3).
	PacketWidth float64
	// Impulses is the kernel density, in kernels per cube of the reach's edge
	// (hodge's impulses_per_cell; 8 by default). The lattice spacing is
	// reach / cbrt(Impulses), and about 4.19 * Impulses kernels reach a point.
	Impulses int
	// Pool is the direction pool's size (hodge: 4096; 0 means that).
	Pool int
	// Seed seeds both the pool draw and the spatial hash.
	Seed int64
	// Orientation, if set, returns a rotation per kernel centre (row-major,
	// applied as R . n), called once with every centre.
	Orientation func(centers [][3]float64) ([][3][3]float64, error)
	// Scale, if set, returns the wavelength multiplier (> 0) per kernel
	// centre, called once with every centre.
	Scale func(centers [][3]float64) ([]float64, error)
	// Domain is the finite world-space region needed by a driven field. The
	// bake includes additional kernel centres out to Reach beyond every face.
	Domain *[6]float64
	// Periodic permits constant scale/orientation callbacks to retain the
	// stationary table. Spatially varying callbacks must leave this false.
	Periodic bool
}

// Site is one lattice site's kernel. Its centre is its index times Spacing.
type Site struct {
	// Dir is the unit wave direction, after Orientation.
	Dir [3]float64
	// Beta is the wavenumber 2 pi / (Period * scale).
	Beta float64
	// Gamma is the phase.
	Gamma float64
}

// Table is a baked spinodoid.
type Table struct {
	Constants
	Layout
	// Sites are indexed (z*Size[1] + y)*Size[0] + x, for local site
	// (x, y, z); world centre is (local + Origin) * Spacing.
	Sites []Site
}

// Bake builds the table for p.
func Bake(p Params) (*Table, error) {
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

	constants := TableConstants(p.Period, p.PacketWidth, p.Impulses)
	layout, err := NewLayout(constants, p.Domain, p.Periodic || (p.Scale == nil && p.Orientation == nil))
	if err != nil {
		return nil, err
	}
	t := &Table{Constants: constants, Layout: layout, Sites: make([]Site, layout.Count())}
	n := len(t.Sites)
	centers := make([][3]float64, n)
	dirs := make([][3]float64, n)
	for z := range t.Size[2] {
		for y := range t.Size[1] {
			for x := range t.Size[0] {
				site := [3]int64{int64(x + t.Origin[0]), int64(y + t.Origin[1]), int64(z + t.Origin[2])}
				i := (z*t.Size[1]+y)*t.Size[0] + x
				var randomSite [3]int64
				for a := range 3 {
					centers[i][a] = float64(site[a]) * t.Spacing
					// Preserve the original world-anchored random realization.
					// Only the driver sampling gains an unwrapped domain.
					randomSite[a] = int64(tableOrigin + wrapSite(site[a]))
				}
				idx := int64(cellUniform(randomSite, 0, p.Seed) * float64(len(baseDirs)))
				idx = min(max(idx, 0), int64(len(baseDirs)-1))
				dirs[i] = baseDirs[idx]
				t.Sites[i].Gamma = cellUniform(randomSite, 1, p.Seed) * 2 * pi
			}
		}
	}

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
	for i := range t.Sites {
		dir := dirs[i]
		if rots != nil {
			r := rots[i]
			dir = [3]float64{dot3(r[0], dir), dot3(r[1], dir), dot3(r[2], dir)}
		}
		t.Sites[i].Dir = normalize(dir)
		s := 1.0
		if scales != nil {
			s = scales[i]
			if !(s > 0) || math.IsInf(s, 0) {
				return nil, fmt.Errorf("scale must be > 0, got %v at %v", s, centers[i])
			}
		}
		t.Sites[i].Beta = 2 * pi / (p.Period * s)
	}
	return t, nil
}

// wrapSite maps an unwrapped site index to the table's [0, TableSites).
func wrapSite(c int64) int {
	m := (c - tableOrigin) % TableSites
	if m < 0 {
		m += TableSites
	}
	return int(m)
}

// Sample is the kernel sum at q, divided by Sigma (unit variance), and beta,
// the envelope-energy-weighted wavenumber the field carries there
// (BaseBeta where no kernel reaches), with both gradients.
func (t *Table) Sample(q [3]float64) (field float64, grad [3]float64, beta float64, gbeta [3]float64) {
	if !t.Contains(q) {
		d, g := t.Outside(q)
		for a := range 3 {
			g[a] *= t.BaseBeta
		}
		return d * t.BaseBeta, g, t.BaseBeta, [3]float64{}
	}
	a2 := t.A * t.A
	r2max := t.Reach * t.Reach
	var lo, hi [3]int64
	for a := range 3 {
		lo[a] = int64(math.Ceil((q[a] - t.Reach) / t.Spacing))
		hi[a] = int64(math.Floor((q[a] + t.Reach) / t.Spacing))
	}
	var acc, wAcc, bwAcc float64
	var gw, gbw [3]float64
	for z := lo[2]; z <= hi[2]; z++ {
		for y := lo[1]; y <= hi[1]; y++ {
			for x := lo[0]; x <= hi[0]; x++ {
				r := [3]float64{q[0] - float64(x)*t.Spacing, q[1] - float64(y)*t.Spacing, q[2] - float64(z)*t.Spacing}
				d2 := dot3(r, r)
				if d2 >= r2max {
					continue
				}
				index, ok := t.Layout.index(x, y, z)
				if !ok {
					continue
				}
				s := &t.Sites[index]
				g := math.Exp(-pi * a2 * d2)
				taper := 1 - d2/r2max
				k := g * taper * taper
				// d/dr of g taper^2, per unit r: g taper (-2 pi a^2 taper - 4 / R^2).
				dk := g * taper * (-2*pi*a2*taper - 4/r2max)
				phase := s.Beta*dot3(r, s.Dir) + s.Gamma
				sn, co := math.Sincos(phase)
				acc += k * co
				w := k * k
				wAcc += w
				bwAcc += w * s.Beta
				for ax := range 3 {
					grad[ax] += dk*r[ax]*co - k*sn*s.Beta*s.Dir[ax]
					dw := 2 * k * dk * r[ax]
					gw[ax] += dw
					gbw[ax] += dw * s.Beta
				}
			}
		}
	}
	field = acc / t.Sigma
	for ax := range 3 {
		grad[ax] /= t.Sigma
	}
	beta = t.BaseBeta
	if wAcc > 0 {
		if b := bwAcc / wAcc; b > 0 {
			beta = b
			for ax := range 3 {
				gbeta[ax] = (gbw[ax] - b*gw[ax]) / wAcc
			}
		}
	}
	return field, grad, beta, gbeta
}

// Level is the threshold that leaves fraction `solidFraction` of the
// (asymptotically standard normal) field below it: the bias a
// ...BySolidFraction spinodoid passes, and hodge's solid_fraction= path. It is
// an estimate, exact only where the kernel sum is Gaussian, which the median
// (fraction 0.5, level 0) always is.
func Level(solidFraction float64) (float64, error) {
	if !(solidFraction > 0 && solidFraction < 1) {
		return 0, fmt.Errorf("solid fraction must lie in (0, 1), got %v", solidFraction)
	}
	return math.Sqrt2 * math.Erfinv(2.0*solidFraction-1.0), nil
}

// CheckBias rejects a bias that cannot be a threshold. A bias is the level
// itself, in units of the field's standard deviation, as a TPMS level t is:
// 0 is the median, ±1 leaves about 16% / 84% below it, and any finite value
// is allowed.
func CheckBias(bias float64) error {
	if math.IsNaN(bias) || math.IsInf(bias, 0) {
		return fmt.Errorf("bias must be finite, got %v (a ...BySolidFraction fraction "+
			"outside (0, 1) gives this)", bias)
	}
	return nil
}

// Distance is the field leveled at level and divided by the local wavenumber
// (local true) or by BaseBeta, with its gradient. Negative in the solid phase.
func (t *Table) Distance(q [3]float64, level float64, local bool) (float64, [3]float64) {
	if !t.Contains(q) {
		return t.Outside(q)
	}
	field, grad, beta, gbeta := t.Sample(q)
	if !local {
		beta, gbeta = t.BaseBeta, [3]float64{}
	}
	num := field - level
	var g [3]float64
	for ax := range 3 {
		g[ax] = grad[ax]/beta - num*gbeta[ax]/(beta*beta)
	}
	return num / beta, g
}

// SiteStride is how many float64s Flat writes per site.
const SiteStride = 5

// Flat is the table's sites as consecutive float64s, SiteStride per site in
// Sites order: direction xyz, wavenumber, phase.
func (t *Table) Flat() []float64 {
	out := make([]float64, 0, len(t.Sites)*SiteStride)
	for _, s := range t.Sites {
		out = append(out, s.Dir[0], s.Dir[1], s.Dir[2], s.Beta, s.Gamma)
	}
	return out
}
