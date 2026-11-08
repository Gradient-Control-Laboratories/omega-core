package topofold

// MeshGeometry provides basic scalar measurements for a mesh.
type MeshGeometry interface {
	Area() float64
	Volume() float64
}

// AnnotatedMesh provides B-rep topological queries and medial axis measurements.
type AnnotatedMesh interface {
	FacePoint(face int, u, v float64) ([3]float64, error)
	EdgePoint(edge int, t float64) ([3]float64, error)
	EdgePointComp(edge, comp int, t float64) ([3]float64, error)
	VertexPoint(vertex int) ([3]float64, error)
	CircleCenter(edge int) ([3]float64, error)
	FacePointNormal(face int, u, v float64) ([3]float64, [3]float64, error)
	EdgePointNormal(edge int, t float64) ([3]float64, [3]float64, error)
	EdgePointNormalComp(edge, comp int, t float64) ([3]float64, [3]float64, error)
	VertexPointNormal(vertex int) ([3]float64, [3]float64, error)
	CircleCenterNormal(edge int) ([3]float64, [3]float64, error)
	MedialMax() (float64, [3]float64, error)
	MedialHotSpots(keep func([3]float64, [3]float64) bool, count int) ([][3]float64, error)
}

// GeometryResolver resolves mesh properties and annotated topology for body slots.
type GeometryResolver interface {
	GetMesh(slot int) (MeshGeometry, bool)
	GetAnnotatedMesh(slot int) (AnnotatedMesh, bool, error)
}

