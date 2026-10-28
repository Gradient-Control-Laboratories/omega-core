package spinodoid

import (
	"fmt"
	"math"
	"strings"
)

// Constants are a table's scalar constants, which depend only on the period,
// packet width and kernel density; a backend that cannot bake (WGSL) computes
// them at codegen from the call's constant arguments, so only the sites
// themselves have to be shipped.
type Constants struct {
	// Reach is the kernel's support radius R: hodge's cutoff radius, where
	// the untapered envelope falls to 0.01.
	Reach float64
	// Spacing is the lattice spacing, Reach / cbrt(impulses).
	Spacing float64
	// A is the envelope parameter 1 / (PacketWidth * Period), the same for
	// every kernel: the envelope is decoupled from the local wavelength.
	A float64
	// Sigma is the unit-variance normalization of the kernel sum.
	Sigma float64
	// BaseBeta is the wavenumber of Period, the fallback where no kernel reaches.
	BaseBeta float64
}

// TableConstants computes a table's constants. Sigma comes from the kernel
// density and the tapered kernel's energy: a sum of kernels k with random
// phases has variance (1 / (2 h^3)) * integral of k^2 over space.
func TableConstants(period, packetWidth float64, impulses int) Constants {
	a := 1.0 / (packetWidth * period)
	reach := math.Sqrt(-math.Log(envelopeCutoff)/pi) / a
	spacing := reach / math.Cbrt(float64(impulses))
	// Simpson's rule on integral_0^R k(r)^2 4 pi r^2 dr.
	const steps = 4096
	f := func(r float64) float64 {
		taper := 1 - r*r/(reach*reach)
		k := math.Exp(-pi*a*a*r*r) * taper * taper
		return k * k * 4 * pi * r * r
	}
	dr := reach / steps
	sum := f(0) + f(reach)
	for i := 1; i < steps; i++ {
		w := 2.0
		if i%2 == 1 {
			w = 4
		}
		sum += w * f(float64(i)*dr)
	}
	energy := sum * dr / 3
	return Constants{
		Reach:    reach,
		Spacing:  spacing,
		A:        a,
		Sigma:    math.Sqrt(0.5 * energy / (spacing * spacing * spacing)),
		BaseBeta: 2.0 * pi / period,
	}
}

// Call is one gaborSpinodoid call site, as every backend sees it: the
// constant arguments and the chain of inlined calls it sits in
// (build.Scope.InlineCallChain). Its Key names the call's table in every
// backend, so one backend's bake can be matched to another's call. The drivers
// are not part of it: the chain already tells driven calls apart, and leaving
// them out means no backend has to agree with another on whether a driver
// folded to a constant.
type Call struct {
	Period      float64
	Theta       []float64
	Poles       []float64
	PacketWidth float64
	Impulses    int
	Pool        int
	Seed        int64
	LevelBias   float64
	Local       bool
	// Fast is true for the lattice table (kernels.go), false for hodge's
	// (hodge.go). The two are different fields from the same arguments, so the
	// kind is part of the key: a preview asks for the lattice table of a
	// hodge-exact call by taking its key and setting this.
	Fast  bool
	Chain string
}

// Key is the call's table key.
func (c Call) Key() string {
	num := func(v float64) string { return fmt.Sprintf("%016x", math.Float64bits(v)) }
	list := func(vs []float64) string {
		parts := make([]string, len(vs))
		for i, v := range vs {
			parts[i] = num(v)
		}
		return strings.Join(parts, ",")
	}
	return fmt.Sprintf("p%s|t%s|o%s|w%s|i%d|n%d|s%d|l%s|L%t|F%t|c%s",
		num(c.Period), list(c.Theta), list(c.Poles), num(c.PacketWidth),
		c.Impulses, c.Pool, c.Seed, num(c.LevelBias), c.Local, c.Fast, c.Chain)
}

// GPU texel layout of a table, shared by the packer (GPUTexels) and the WGSL
// emitter: an RGBA32Float 2D texture GPUWidth x table.Rows(), one texel per
// site in y-z-x order, so neighbouring sites along x, which neighbouring
// pixels read, are neighbouring texels. The 64-cubed stationary table keeps
// its original packing. 2D
// because three.js binds every 3D texture as a plain texture_3d<f32> and the
// layout is the same either way. Its four floats are the wave vector, the
// direction times the wavenumber (the shader takes the wavenumber back as its
// length, and only when a scale driver makes it vary), and the phase.
const (
	GPUWidth = TableSites * TableSites
	GPURows  = TableSites
)

// GPUTexels packs the table as GPUWidth x table.Rows() RGBA float32 texels, row by
// row.
func (t *Table) GPUTexels() []float32 {
	out := make([]float32, GPUWidth*t.Rows()*4)
	for z := range t.Size[2] {
		for y := range t.Size[1] {
			for x := range t.Size[0] {
				s := t.Sites[(z*t.Size[1]+y)*t.Size[0]+x]
				i := ((y*t.Size[2]+z)*t.Size[0] + x) * 4
				out[i], out[i+1], out[i+2], out[i+3] =
					float32(s.Beta*s.Dir[0]), float32(s.Beta*s.Dir[1]), float32(s.Beta*s.Dir[2]), float32(s.Gamma)
			}
		}
	}
	return out
}
