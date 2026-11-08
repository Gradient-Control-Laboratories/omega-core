package spinodoid

import (
	"fmt"
	"math"
)

// MaxTableSites bounds driver evaluation, CPU tables, and GPU uploads. A
// rectangular domain around an imported part is usually much smaller than
// the original 64-cubed table; a huge scene must fail instead of exhausting RAM.
const MaxTableSites = 1 << 20

// Layout describes the actual world-space kernel centres stored in a table.
// Stationary and legacy headerless tables are periodic. Bounded driven tables
// cover Bounds, with enough extra sites for every kernel reaching that domain,
// and never wrap drivers.
type Layout struct {
	Origin   [3]int
	Size     [3]int
	Bounds   [6]float64
	Periodic bool
}

func (l Layout) Count() int { return l.Size[0] * l.Size[1] * l.Size[2] }
func (l Layout) Rows() int  { return (l.Count() + GPUWidth - 1) / GPUWidth }

func (l Layout) Contains(q [3]float64) bool {
	if l.Periodic {
		return true
	}
	for a := range 3 {
		if q[a] < l.Bounds[a] || q[a] > l.Bounds[a+3] {
			return false
		}
	}
	return true
}

// Outside returns distance to the finite domain and its outward unit gradient.
// This conservative extension lets a ray marcher reach the domain from the
// surrounding scene instead of jumping over the part with an infinite value.
func (l Layout) Outside(q [3]float64) (float64, [3]float64) {
	var g [3]float64
	var squared float64
	for a := range 3 {
		g[a] = q[a] - max(l.Bounds[a], min(q[a], l.Bounds[a+3]))
		squared += g[a] * g[a]
	}
	d := math.Sqrt(squared)
	if d > 0 {
		for a := range 3 {
			g[a] /= d
		}
	}
	return d, g
}

func (l Layout) index(x, y, z int64) (int, bool) {
	coords := [3]int64{x, y, z}
	var local [3]int
	for a, c := range coords {
		if l.Periodic {
			local[a] = wrapSite(c)
		} else {
			d := c - int64(l.Origin[a])
			if d < 0 || d >= int64(l.Size[a]) {
				return 0, false
			}
			local[a] = int(d)
		}
	}
	return (local[2]*l.Size[1]+local[1])*l.Size[0] + local[0], true
}

func NewLayout(c Constants, domain *[6]float64, periodic bool) (Layout, error) {
	if periodic || domain == nil {
		return Layout{Origin: [3]int{tableOrigin, tableOrigin, tableOrigin}, Size: [3]int{TableSites, TableSites, TableSites}, Periodic: true}, nil
	}
	l := Layout{Bounds: *domain}
	count := 1
	for a := range 3 {
		lo, hi := domain[a], domain[a+3]
		if math.IsNaN(lo) || math.IsNaN(hi) || math.IsInf(lo, 0) || math.IsInf(hi, 0) || hi <= lo {
			return Layout{}, fmt.Errorf("spinodoid domain must have finite increasing bounds on every axis")
		}
		first := math.Floor((lo - c.Reach) / c.Spacing)
		last := math.Ceil((hi + c.Reach) / c.Spacing)
		if math.Abs(first) > 1<<29 || math.Abs(last) > 1<<29 || last-first+1 > MaxTableSites {
			return Layout{}, fmt.Errorf("graded spinodoid domain is too large; reduce the scene/domain extent or increase the base period")
		}
		l.Origin[a] = int(first)
		l.Size[a] = int(last-first) + 1
		if l.Size[a] > MaxTableSites/count {
			return Layout{}, fmt.Errorf("graded spinodoid domain exceeds %d kernel sites; reduce the scene/domain extent or increase the base period", MaxTableSites)
		}
		count *= l.Size[a]
	}
	return l, nil
}

// SceneDomain reads the canvas's geometry domain or falls back to the authored
// scene cube. A caller without either retains the legacy periodic table: an
// arbitrary headerless analytical shape has no known finite sampling domain.
func SceneDomain(bounds []float64, sceneSize float64) *[6]float64 {
	if len(bounds) == 6 {
		domain := [6]float64(bounds)
		return &domain
	}
	if sceneSize > 0 && !math.IsInf(sceneSize, 0) {
		half := sceneSize / 2
		return &[6]float64{-half, -half, -half, half, half, half}
	}
	return nil
}
