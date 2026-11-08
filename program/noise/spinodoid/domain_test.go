package spinodoid

import (
	"math"
	"testing"
)

func domainParams(bounds [6]float64) Params {
	return Params{Period: 2, PacketWidth: 3, Impulses: 8, Theta: []float64{math.Pi / 2, math.Pi / 2, math.Pi / 2}, Domain: &bounds}
}

func TestGradedDomainContinuesPastFormerWrapBoundary(t *testing.T) {
	p := domainParams([6]float64{-60, -5, -15, 70, 280, 50})
	p.Scale = func(centers [][3]float64) ([]float64, error) {
		out := make([]float64, len(centers))
		for i, q := range centers {
			out[i] = 0.1 + 3.9*math.Max(0, math.Min(1, q[1]/264))
		}
		return out, nil
	}
	table, err := Bake(p)
	if err != nil {
		t.Fatal(err)
	}
	if table.Periodic {
		t.Fatal("graded geometry domain must not repeat")
	}
	last := 0.0
	for _, y := range []float64{10, 80, 112, 116, 120, 160, 220, 260} {
		_, _, beta, _ := table.Sample([3]float64{0, y, 10})
		got, want := table.BaseBeta/beta, 0.1+3.9*y/264
		if math.Abs(got-want) > 0.03 || got <= last {
			t.Fatalf("at y=%g scale=%g, want %g and greater than %g", y, got, want, last)
		}
		last = got
	}
}

func TestTranslatedRectangularDomainPreservesStationaryRealization(t *testing.T) {
	p := domainParams([6]float64{980, -30, -10, 1260, 40, 30})
	p.Scale = func(centers [][3]float64) ([]float64, error) {
		out := make([]float64, len(centers))
		for i := range out {
			out[i] = 1
		}
		return out, nil
	}
	bounded, err := Bake(p)
	if err != nil {
		t.Fatal(err)
	}
	p.Scale, p.Domain = nil, nil
	stationary, err := Bake(p)
	if err != nil {
		t.Fatal(err)
	}
	if bounded.Origin[0] <= 0 || bounded.Size[0] <= TableSites || bounded.Size[1] >= TableSites || bounded.Size[2] >= TableSites {
		t.Fatalf("expected shifted rectangular table, got %+v", bounded.Layout)
	}
	for _, q := range [][3]float64{{981.5, 0, 0}, {1100.25, 20.1, 10.2}, {1259.5, 39, 29}} {
		got, ggot := bounded.Distance(q, 0, true)
		want, gwant := stationary.Distance(q, 0, true)
		if math.Abs(got-want) > 1e-12 {
			t.Fatalf("random realization moved at %v: %g != %g", q, got, want)
		}
		for a := range 3 {
			if math.Abs(ggot[a]-gwant[a]) > 1e-12 {
				t.Fatal("gradient realization changed")
			}
		}
	}
	for _, q := range [][3]float64{{979, 0, 0}, {1100, 41, 0}, {1100, 0, 31}} {
		d, g := bounded.Distance(q, 0, true)
		if d != 1 || math.Abs(dot3(g, g)-1) > 1e-12 {
			t.Fatalf("outside domain must guide ray marching to its boundary: %g %v", d, g)
		}
	}
	// The scene can be much larger than the imported part. The conservative
	// extension must never step past the domain, even with a nonzero level.
	q := [3]float64{800, -270, -10}
	d, g := bounded.Distance(q, 8, false)
	if math.Abs(d-300) > 1e-12 || math.Abs(g[0]+.6) > 1e-12 || math.Abs(g[1]+.8) > 1e-12 {
		t.Fatalf("outside corner distance/gradient: %g %v", d, g)
	}
}

func TestRectangularTexturePackingAndLimits(t *testing.T) {
	p := domainParams([6]float64{-10, -5, 40, 20, 80, 55})
	p.Scale = func(cs [][3]float64) ([]float64, error) {
		out := make([]float64, len(cs))
		for i, c := range cs {
			out[i] = 1 + c[1]/1000
		}
		return out, nil
	}
	table, err := Bake(p)
	if err != nil {
		t.Fatal(err)
	}
	texels := table.GPUTexels()
	if len(texels) != GPUWidth*table.Rows()*4 {
		t.Fatal("texture rows do not cover rectangular data")
	}
	for z := range table.Size[2] {
		for y := range table.Size[1] {
			for x := range table.Size[0] {
				s := table.Sites[(z*table.Size[1]+y)*table.Size[0]+x]
				i := ((y*table.Size[2]+z)*table.Size[0] + x) * 4
				if texels[i] != float32(s.Beta*s.Dir[0]) || texels[i+3] != float32(s.Gamma) {
					t.Fatal("rectangular GPU address differs from CPU")
				}
			}
		}
	}
	huge := [6]float64{-10000, -10000, -10000, 10000, 10000, 10000}
	if _, err := NewLayout(table.Constants, &huge, false); err == nil {
		t.Fatal("unbounded allocation accepted")
	}
	bad := [6]float64{0, 0, 0, math.Inf(1), 1, 1}
	if _, err := NewLayout(table.Constants, &bad, false); err == nil {
		t.Fatal("nonfinite bounds accepted")
	}
}

func TestConstantAndHeaderlessTablesKeepLegacyPeriodicity(t *testing.T) {
	p := domainParams([6]float64{1000, 1000, 1000, 1001, 1001, 1001})
	table, err := Bake(p)
	if err != nil {
		t.Fatal(err)
	}
	if !table.Periodic || table.Size != [3]int{64, 64, 64} || table.Rows() != 64 {
		t.Fatal("stationary table layout changed")
	}
	p.Domain = nil
	p.Scale = func(cs [][3]float64) ([]float64, error) {
		out := make([]float64, len(cs))
		for i, c := range cs {
			out[i] = 1 + .001*c[1]
		}
		return out, nil
	}
	table, err = Bake(p)
	if err != nil {
		t.Fatal(err)
	}
	if !table.Periodic {
		t.Fatal("headerless compatibility domain must remain periodic")
	}
}
