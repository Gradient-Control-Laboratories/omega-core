package meshasset

import (
	"math"
	"testing"
)

// boxMesh builds a triangulated axis-aligned box spanning min..max.
func boxMesh(min, max [3]float64) *TriMesh {
	v := func(x, y, z int) [3]float64 {
		pick := func(sel int, lo, hi float64) float64 {
			if sel == 0 {
				return lo
			}
			return hi
		}
		return [3]float64{pick(x, min[0], max[0]), pick(y, min[1], max[1]), pick(z, min[2], max[2])}
	}
	quads := [][4][3]float64{
		{v(0, 0, 0), v(1, 0, 0), v(1, 1, 0), v(0, 1, 0)}, // z = min
		{v(0, 0, 1), v(1, 0, 1), v(1, 1, 1), v(0, 1, 1)}, // z = max
		{v(0, 0, 0), v(1, 0, 0), v(1, 0, 1), v(0, 0, 1)}, // y = min
		{v(0, 1, 0), v(1, 1, 0), v(1, 1, 1), v(0, 1, 1)}, // y = max
		{v(0, 0, 0), v(0, 1, 0), v(0, 1, 1), v(0, 0, 1)}, // x = min
		{v(1, 0, 0), v(1, 1, 0), v(1, 1, 1), v(1, 0, 1)}, // x = max
	}
	mesh := &TriMesh{}
	for _, q := range quads {
		mesh.Triangles = append(mesh.Triangles,
			Triangle{V0: q[0], V1: q[1], V2: q[2]},
			Triangle{V0: q[0], V1: q[2], V2: q[3]})
	}
	return mesh
}

func TestComputeDigestBox(t *testing.T) {
	// An elongated box centered at (10, 20, 30): 40 long in x, 4 in y, 2 in z.
	min := [3]float64{-10, 18, 29}
	max := [3]float64{30, 22, 31}
	d := ComputeDigest(boxMesh(min, max))

	for k := 0; k < 3; k++ {
		if math.Abs(d.BboxMin[k]-min[k]) > 1e-9 || math.Abs(d.BboxMax[k]-max[k]) > 1e-9 {
			t.Fatalf("bbox = %v..%v, want %v..%v", d.BboxMin, d.BboxMax, min, max)
		}
	}

	want := [3]float64{10, 20, 30}
	for k := 0; k < 3; k++ {
		if math.Abs(d.Centroid[k]-want[k]) > 1e-6 {
			t.Errorf("centroid[%d] = %v, want %v", k, d.Centroid[k], want[k])
		}
	}

	// The dominant principal axis of a 40x4x2 box is x — signs are
	// normalized so the largest-magnitude component is positive.
	ax := d.PrincipalAxes[0]
	if math.Abs(ax[0]-1) > 1e-6 || math.Abs(ax[1]) > 1e-6 || math.Abs(ax[2]) > 1e-6 {
		t.Errorf("principal axis 0 = %v, want +x", ax)
	}
	if math.Abs(d.Extents[0]-20) > 1e-6 {
		t.Errorf("extent 0 = %v, want 20", d.Extents[0])
	}

	// The frame is right-handed: axis0 x axis1 = axis2.
	a, b, c := d.PrincipalAxes[0], d.PrincipalAxes[1], d.PrincipalAxes[2]
	cross := [3]float64{
		a[1]*b[2] - a[2]*b[1],
		a[2]*b[0] - a[0]*b[2],
		a[0]*b[1] - a[1]*b[0],
	}
	for k := 0; k < 3; k++ {
		if math.Abs(cross[k]-c[k]) > 1e-6 {
			t.Fatalf("axes not right-handed: axis0 x axis1 = %v, axis2 = %v", cross, c)
		}
	}
}

func TestComputeDigestRotatedBox(t *testing.T) {
	// Rotate the same elongated box 30 degrees about z; the dominant axis
	// should follow.
	base := boxMesh([3]float64{-10, -2, -1}, [3]float64{10, 2, 1})
	sin, cos := math.Sin(math.Pi/6), math.Cos(math.Pi/6)
	rot := func(v [3]float64) [3]float64 {
		return [3]float64{cos*v[0] - sin*v[1], sin*v[0] + cos*v[1], v[2]}
	}
	mesh := &TriMesh{}
	for _, tr := range base.Triangles {
		mesh.Triangles = append(mesh.Triangles, Triangle{V0: rot(tr.V0), V1: rot(tr.V1), V2: rot(tr.V2)})
	}

	d := ComputeDigest(mesh)
	ax := d.PrincipalAxes[0]
	// Expect +-(cos30, sin30, 0).
	dot := math.Abs(ax[0]*cos + ax[1]*sin)
	if math.Abs(dot-1) > 1e-6 {
		t.Errorf("principal axis 0 = %v, want +-(%v, %v, 0)", ax, cos, sin)
	}
	if math.Abs(d.Extents[0]-10) > 1e-6 {
		t.Errorf("extent 0 = %v, want 10", d.Extents[0])
	}
}

func TestComputeDigestEmpty(t *testing.T) {
	d := ComputeDigest(nil)
	if d.Area != 0 {
		t.Errorf("area = %v, want 0", d.Area)
	}
}
