package eval

import "fmt"

// value is a shaped constant: a leaf scalar or vector (elems nil, scalars
// holds the components; a scalar is length 1) or a list of shaped elements
// (a matrix is the list of its column vectors). The shape exists only inside
// the evaluator — the exported entry points flatten leaves and lists alike
// to the flat []float64 callers consume, in the same depth-first
// (column-major) order the flat evaluator produced.
type value struct {
	scalars []float64
	elems   []value
}

func scalarVal(x float64) value { return value{scalars: []float64{x}} }

func vectorVal(xs []float64) value { return value{scalars: xs} }

func listVal(elems []value) value { return value{elems: elems} }

func (v value) isList() bool { return v.elems != nil }

// isScalar reports whether v is a single scalar value.
func (v value) isScalar() bool { return v.elems == nil && len(v.scalars) == 1 }

// count is the number of top-level elements: list elements for a list,
// components for a leaf.
func (v value) count() int {
	if v.isList() {
		return len(v.elems)
	}
	return len(v.scalars)
}

// elementAt returns the k-th top-level element: a whole element value for a
// list (a matrix yields its k-th column vector), the k-th component as a
// scalar for a leaf.
func (v value) elementAt(k int) (value, error) {
	if k < 0 || k >= v.count() {
		return value{}, fmt.Errorf("index %d out of range for length %d", k, v.count())
	}
	if v.isList() {
		return v.elems[k], nil
	}
	return scalarVal(v.scalars[k]), nil
}

// elements splits v into its top-level elements (see elementAt).
func (v value) elements() []value {
	if v.isList() {
		return v.elems
	}
	out := make([]value, len(v.scalars))
	for i, x := range v.scalars {
		out[i] = scalarVal(x)
	}
	return out
}

// flatten appends v's scalars to dst depth-first — for a list of column
// vectors this is column-major order, matching the flat representation
// callers consume.
func (v value) flatten(dst []float64) []float64 {
	if v.isList() {
		for _, e := range v.elems {
			dst = e.flatten(dst)
		}
		return dst
	}
	return append(dst, v.scalars...)
}

// floats is v flattened to a fresh slice.
func (v value) floats() []float64 {
	return v.flatten(make([]float64, 0, len(v.scalars)))
}
