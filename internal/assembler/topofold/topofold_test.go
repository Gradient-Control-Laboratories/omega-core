package topofold

import (
	"testing"

	"gradientcontrol.com/omega/core/assembler/topofold"
	"gradientcontrol.com/omega/core/program"
)

type mockMesh struct {
	area   float64
	volume float64
}

func (m *mockMesh) Area() float64   { return m.area }
func (m *mockMesh) Volume() float64 { return m.volume }

type mockAnnotatedMesh struct {
	facePt [3]float64
	edgePt [3]float64
	vertPt [3]float64
	normal [3]float64
}

func (m *mockAnnotatedMesh) FacePoint(face int, u, v float64) ([3]float64, error) {
	return [3]float64{m.facePt[0] + u, m.facePt[1] + v, m.facePt[2]}, nil
}

func (m *mockAnnotatedMesh) EdgePoint(edge int, t float64) ([3]float64, error) {
	return [3]float64{m.edgePt[0] + t, m.edgePt[1], m.edgePt[2]}, nil
}

func (m *mockAnnotatedMesh) EdgePointComp(edge, comp int, t float64) ([3]float64, error) {
	return [3]float64{m.edgePt[0] + t, m.edgePt[1] + float64(comp), m.edgePt[2]}, nil
}

func (m *mockAnnotatedMesh) VertexPoint(vertex int) ([3]float64, error) {
	return m.vertPt, nil
}

func (m *mockAnnotatedMesh) CircleCenter(edge int) ([3]float64, error) {
	return [3]float64{10, 20, 30}, nil
}

func (m *mockAnnotatedMesh) FacePointNormal(face int, u, v float64) ([3]float64, [3]float64, error) {
	return [3]float64{m.facePt[0] + u, m.facePt[1] + v, m.facePt[2]}, m.normal, nil
}

func (m *mockAnnotatedMesh) EdgePointNormal(edge int, t float64) ([3]float64, [3]float64, error) {
	return [3]float64{m.edgePt[0] + t, m.edgePt[1], m.edgePt[2]}, m.normal, nil
}

func (m *mockAnnotatedMesh) EdgePointNormalComp(edge, comp int, t float64) ([3]float64, [3]float64, error) {
	return [3]float64{m.edgePt[0] + t, m.edgePt[1] + float64(comp), m.edgePt[2]}, m.normal, nil
}

func (m *mockAnnotatedMesh) VertexPointNormal(vertex int) ([3]float64, [3]float64, error) {
	return m.vertPt, m.normal, nil
}

func (m *mockAnnotatedMesh) CircleCenterNormal(edge int) ([3]float64, [3]float64, error) {
	return [3]float64{10, 20, 30}, m.normal, nil
}

func (m *mockAnnotatedMesh) MedialMax() (float64, [3]float64, error) {
	return 5.5, [3]float64{1, 2, 3}, nil
}

func (m *mockAnnotatedMesh) MedialHotSpots(keep func([3]float64, [3]float64) bool, count int) ([][3]float64, error) {
	return [][3]float64{{1, 2, 3}, {4, 5, 6}}, nil
}

type mockResolver struct {
	meshes    map[int]*mockMesh
	annotated map[int]*mockAnnotatedMesh
}

func (r *mockResolver) GetMesh(slot int) (topofold.MeshGeometry, bool) {
	m, ok := r.meshes[slot]
	return m, ok
}

func (r *mockResolver) GetAnnotatedMesh(slot int) (topofold.AnnotatedMesh, bool, error) {
	am, ok := r.annotated[slot]
	if !ok {
		return nil, false, nil
	}
	return am, true, nil
}

func TestTopoFoldPointAndMeshQueries(t *testing.T) {
	resolver := &mockResolver{
		meshes: map[int]*mockMesh{
			0: {area: 120.0, volume: 450.0},
		},
		annotated: map[int]*mockAnnotatedMesh{
			0: {
				facePt: [3]float64{10, 0, 0},
				edgePt: [3]float64{0, 20, 0},
				vertPt: [3]float64{0, 0, 30},
				normal: [3]float64{0, 0, 1},
			},
		},
	}

	// Build a program with facePoint and meshVolume calls
	prog := &program.Program{
		Primary: &program.Package{
			Files: []*program.PackageFile{
				{
					StepImports: []*program.StepImport{
						{Index: 0, Path: "part.step"},
					},
					Assignments: []*program.Assignment{
						{
							Variables: []*program.Variable{{Name: "pt"}},
							Expression: &program.Operand{
								T: &program.Operand_FnCall{
									FnCall: &program.FnCall{
										Reference: &program.Reference{Name: "facePoint"},
										Arguments: []*program.Operand{
											{T: &program.Operand_Value{Value: 0}}, // body 0
											{T: &program.Operand_Value{Value: 1}}, // face 1
											{
												T: &program.Operand_Expression{
													Expression: &program.Expression{
														Operation: program.Operation_COLLECT,
														Operands: []*program.Operand{
															{T: &program.Operand_Value{Value: 0.5}},
															{T: &program.Operand_Value{Value: 0.25}},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						{
							Variables: []*program.Variable{{Name: "vol"}},
							Expression: &program.Operand{
								T: &program.Operand_FnCall{
									FnCall: &program.FnCall{
										Reference: &program.Reference{Name: "meshVolume"},
										Arguments: []*program.Operand{
											{T: &program.Operand_Value{Value: 0}},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		Dependencies: []*program.Package{
			{Path: "builtin"},
		},
	}

	if !HasTopoCall(prog) {
		t.Fatalf("expected HasTopoCall to be true")
	}

	if err := FoldProgram(prog, nil, resolver); err != nil {
		t.Fatalf("FoldProgram failed: %v", err)
	}

	// Verify assignments were folded to constants in place
	if len(prog.Primary.Files[0].Assignments) != 2 {
		t.Fatalf("expected 2 assignments, got %d", len(prog.Primary.Files[0].Assignments))
	}

	ptAssign := prog.Primary.Files[0].Assignments[0]
	if ptAssign.Expression.GetExpression() == nil || ptAssign.Expression.GetExpression().Operation != program.Operation_COLLECT {
		t.Fatalf("expected pt assignment to be folded to COLLECT vector, got %v", ptAssign.Expression)
	}

	volAssign := prog.Primary.Files[0].Assignments[1]
	if volAssign.Expression.GetValue() != 450.0 {
		t.Fatalf("expected vol assignment to fold to 450.0, got %v", volAssign.Expression.GetValue())
	}
}
