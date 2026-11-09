package types

import (
	"fmt"

	"gradientcontrol.com/omega/core/program"
)

func (pa *PackageAnalyzer) inferBuiltinCall(name string, args []*program.Operand, env *TypeEnv, file *program.PackageFile) ([]*program.Type, error) {
	builtin := program.BuiltinNamed(name)
	if builtin == nil {
		return nil, fmt.Errorf("unknown builtin: %s", name)
	}

	if len(args) < builtin.MinArgs || len(args) > builtin.MaxArgs {
		return nil, fmt.Errorf("%s expects %d..%d arguments, got %d", name, builtin.MinArgs, builtin.MaxArgs, len(args))
	}

	argTypes := make([]*program.Type, len(args))
	for i, arg := range args {
		t, err := inferSingleType(arg, env, file, pa)
		if err != nil {
			return nil, err
		}
		if !t.IsNumeric() {
			return nil, fmt.Errorf("%s expects a numeric, got %s", name, t.TypeString())
		}
		argTypes[i] = t
	}

	switch name {
	case "acosh", "asinh", "atanh", "cosh", "sinh", "tanh", "trunc", "degrees", "radians", "fract", "sign", "normalize", "exp2", "erfinv":
		return []*program.Type{argTypes[0]}, nil

	case "clamp":
		x := argTypes[0]
		lo := argTypes[1]
		hi := argTypes[2]
		_, err := commonType(x, lo)
		if err != nil {
			return nil, fmt.Errorf("clamp bounds incompatible: %v", err)
		}
		_, err = commonType(x, hi)
		if err != nil {
			return nil, fmt.Errorf("clamp bounds incompatible: %v", err)
		}
		return []*program.Type{x}, nil

	case "dot":
		a := argTypes[0]
		b := argTypes[1]
		if a.Size != b.Size || a.IsMatrix() || b.IsMatrix() {
			return nil, fmt.Errorf("dot expects vectors of the same size, got %s and %s", a.TypeString(), b.TypeString())
		}
		return []*program.Type{program.NewScalarType()}, nil

	case "cross":
		a := argTypes[0]
		b := argTypes[1]
		if !isVec3OrScalar(a) || !isVec3OrScalar(b) {
			return nil, fmt.Errorf("cross expects 3-component vectors, got %s and %s", a.TypeString(), b.TypeString())
		}
		return []*program.Type{program.NewListType(3, program.NewScalarType())}, nil

	case "distance":
		a := argTypes[0]
		b := argTypes[1]
		if a.Size != b.Size || a.IsMatrix() || b.IsMatrix() {
			return nil, fmt.Errorf("distance expects vectors of same size, got %s and %s", a.TypeString(), b.TypeString())
		}
		return []*program.Type{program.NewScalarType()}, nil

	case "length":
		if argTypes[0].IsMatrix() {
			return nil, fmt.Errorf("length is not supported on matrices")
		}
		return []*program.Type{program.NewScalarType()}, nil

	case "uiPoint":
		if !isVec3OrScalar(argTypes[0]) {
			return nil, fmt.Errorf("uiPoint expects a 3-component vector, got %s", argTypes[0].TypeString())
		}
		return []*program.Type{program.NewListType(3, program.NewScalarType())}, nil

	case "uiColor":
		if !isVec3OrScalar(argTypes[0]) {
			return nil, fmt.Errorf("uiColor expects a 3-component vector, got %s", argTypes[0].TypeString())
		}
		return []*program.Type{program.NewListType(3, program.NewScalarType())}, nil

	case "uiDirection":
		if !isVec3OrScalar(argTypes[0]) {
			return nil, fmt.Errorf("uiDirection expects a 3-component vector, got %s", argTypes[0].TypeString())
		}
		if len(argTypes) == 2 && !isVec3OrScalar(argTypes[1]) {
			return nil, fmt.Errorf("uiDirection's anchor is a 3-component point, got %s", argTypes[1].TypeString())
		}
		return []*program.Type{program.NewListType(3, program.NewScalarType())}, nil

	case "uiTransform":
		m := argTypes[0]
		if m.Size != 4 || m.ElemType == nil || m.ElemType.Size != 4 {
			return nil, fmt.Errorf("uiTransform expects a 4x4 matrix, got %s", m.TypeString())
		}
		return []*program.Type{m}, nil

	case "uiSpline":
		list := argTypes[0]
		m := list.ElemType
		if list.Size < 3 || list.Size%2 != 1 || m == nil || m.Size != 4 || m.ElemType == nil || m.ElemType.Size != 4 || m.ElemType.ElemType == nil || m.ElemType.ElemType.Size != 1 {
			return nil, fmt.Errorf("uiSpline expects an odd list of at least three 4x4 matrices, got %s", list.TypeString())
		}
		return []*program.Type{list}, nil

	case "facePoint", "edgePoint", "vertexPoint", "circleCenter", "warpField":
		return []*program.Type{program.NewListType(3, program.NewScalarType())}, nil

	case "facePointNormal", "edgePointNormal", "vertexPointNormal", "circleCenterNormal":
		t := program.NewListType(3, program.NewScalarType())
		return []*program.Type{t, t}, nil

	case "meshUnion", "meshIntersect", "meshDifference":
		for _, t := range argTypes {
			if t.Size != 1 && (t.ElemType == nil || t.ElemType.Size != 1) {
				return nil, fmt.Errorf("%s expects meshes, got %s", name, t.TypeString())
			}
		}
		return []*program.Type{program.NewScalarType()}, nil

	case "mesh", "importMesh", "importStep", "importWarp", "bvhSdf", "vdbSdf", "meshWarped", "vdbSdfWarped":
		return []*program.Type{program.NewScalarType()}, nil

	case "meshVolume", "meshArea", "meshMaxThickness":
		return []*program.Type{program.NewScalarType()}, nil

	case "meshThickestPoint":
		return []*program.Type{program.NewListType(3, program.NewScalarType())}, nil

	case "meshThickestPoints":
		return []*program.Type{program.NewMatrixType(10, program.NewListType(3, program.NewScalarType()))}, nil

	case program.NoiseHash3:
		// The case label was the Go constant's NAME ("NoiseHash3") rather than
		// its value ("noiseHash3"), so this arm was dead and the builtin fell
		// through to the scalar default. It returns a vec3: a pseudo-random
		// direction per lattice cell. Cellular noise never noticed, because it
		// only adds the hash to a vector and a scalar broadcasts; Perlin's
		// dot(cornerGradient, offset) does, and so does differentiating either.
		if !isVec3OrScalar(argTypes[0]) {
			return nil, fmt.Errorf("noiseHash3 expects a 3-component vector, got %s", argTypes[0].TypeString())
		}
		// The optional second argument is the octave count: how many tile
		// lookups the hash folds together, hence its repeat period. A scalar,
		// and the backends additionally require it to be a literal 1 or 2.
		if len(argTypes) == 2 && !argTypes[1].IsScalar() {
			return nil, fmt.Errorf("noiseHash3 octaves must be a scalar, got %s", argTypes[1].TypeString())
		}
		return []*program.Type{program.NewListType(3, program.NewScalarType())}, nil

	case program.WorleyF1, program.WorleyF2, program.WorleyFeature,
		program.WorleyFarFeature, program.WorleyNearCell, program.WorleyFarCell:
		// q is a 3-component lattice coordinate. cellularF1 returns the
		// scalar distance; cellularFeature returns the feature point, a vec3.
		if !argTypes[0].IsList() || argTypes[0].Size != 3 {
			return nil, fmt.Errorf("%s expects a 3-component coordinate, got %s",
				name, argTypes[0].TypeString())
		}
		if len(argTypes) == 2 && !argTypes[1].IsScalar() {
			return nil, fmt.Errorf("%s octaves must be a scalar, got %s",
				name, argTypes[1].TypeString())
		}
		switch name {
		case program.WorleyFeature, program.WorleyFarFeature,
			program.WorleyNearCell, program.WorleyFarCell:
			return []*program.Type{program.NewListType(3, program.NewScalarType())}, nil
		}
		return []*program.Type{program.NewScalarType()}, nil

	case program.GaborSpinodoid, program.GaborSpinodoidGradient,
		program.FastGaborSpinodoid, program.FastGaborSpinodoidGradient:
		// q is the query point. theta holds one cone half-angle per pole and
		// poles one axis per angle, as a list of vec3 (a 3xK matrix, which a
		// single pole leaves a plain vec3). orientation is a 3x3 rotation, R . n.
		if !argTypes[0].IsList() || argTypes[0].Size != 3 {
			return nil, fmt.Errorf("%s expects a 3-component point, got %s",
				name, argTypes[0].TypeString())
		}
		poleCount := int32(1)
		if argTypes[2].IsList() {
			poleCount = argTypes[2].Size
		}
		poles := argTypes[3]
		switch {
		case poles.IsMatrix() && poles.Size == poleCount && poles.ElemType != nil && poles.ElemType.Size == 3:
		case poleCount == 1 && poles.IsList() && poles.Size == 3:
		default:
			return nil, fmt.Errorf("%s expects one 3-component pole per cone angle (%d), got %s",
				name, poleCount, poles.TypeString())
		}
		for _, i := range []int{1, 4, 5, 6, 7, 8, 9, 10} {
			if !argTypes[i].IsScalar() {
				return nil, fmt.Errorf("%s argument %d must be a scalar, got %s",
					name, i+1, argTypes[i].TypeString())
			}
		}
		if o := argTypes[11]; !o.IsMatrix() || o.Size != 3 || o.ElemType == nil || o.ElemType.Size != 3 {
			return nil, fmt.Errorf("%s orientation must be a 3x3 matrix, got %s",
				name, o.TypeString())
		}
		if program.IsGaborSpinodoidGradient(name) {
			return []*program.Type{program.NewListType(3, program.NewScalarType())}, nil
		}
		return []*program.Type{program.NewScalarType()}, nil
	}

	return []*program.Type{program.NewScalarType()}, nil
}

func isVec3OrScalar(t *program.Type) bool {
	return t.IsScalar() || (t.IsList() && t.Size == 3)
}
