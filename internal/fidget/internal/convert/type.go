package convert

import (
	"fmt"

	"gradientcontrol.com/omega/core/internal/fidget/internal/fidget"
	"gradientcontrol.com/omega/core/program"
)

var NumType = &numType{}

var Vec3Type = &vecType{len: 3}

var vecTypes = map[int]FidgetNumType{
	2: &vecType{len: 2},
	3: Vec3Type,
	4: &vecType{len: 4},
}

var listVecTypes = make(map[int]map[int]*listVecType)

func VecType(len int, elemType FidgetNumType) (FidgetNumType, error) {
	switch elemType {
	case NumType:
		return VecNumType(len), nil
	default:
		return nil, fmt.Errorf("collection can only contain numbers, got %s", elemType.String())
	}
}

func VecNumType(len int) FidgetNumType {
	vt, ok := vecTypes[len]
	if !ok {
		vt = &vecType{len: len}
		vecTypes[len] = vt
	}
	return vt
}

func ListVecNumType(listLen, elemLen int) FidgetNumType {
	l, ok := listVecTypes[listLen]
	if !ok {
		l = make(map[int]*listVecType)
		listVecTypes[listLen] = l
	}
	lvt, ok := l[elemLen]
	if !ok {
		lvt = &listVecType{listLen: listLen, elemLen: elemLen}
		l[elemLen] = lvt
	}
	return lvt
}

type FidgetNumType interface {
	NumTrees() int
	ExpandTrees(trees []fidget.TreeIndex) ([]fidget.TreeIndex, error)
	Equal(other FidgetNumType) bool
	Common(other FidgetNumType) (FidgetNumType, error)
	String() string
}

func toFidgetNumType(t any) FidgetNumType {
	if t == nil {
		return nil
	}
	switch v := t.(type) {
	case FidgetNumType:
		return v
	case *program.Type:
		if v == nil || (v.Size == 0 && v.FnPath != "") {
			return nil
		}
		if v.ElemType == nil || (v.ElemType.Size == 1 && v.ElemType.ElemType == nil) {
			if v.Size <= 1 {
				return NumType
			}
			return VecNumType(int(v.Size))
		}
		listLen := int(v.Size)
		elemLen := 1
		if v.ElemType != nil {
			elemLen = int(v.ElemType.Size)
		}
		return ListVecNumType(listLen, elemLen)
	default:
		return nil
	}
}

// ----------------------------------------------------------------------------

type numType struct {
	FidgetNumType
}

func (t *numType) Common(other FidgetNumType) (FidgetNumType, error) {
	if other == nil || other == NumType {
		return NumType, nil
	}
	return other.Common(t)
}

func (t *numType) Equal(other FidgetNumType) bool {
	return other == NumType
}

func (t *numType) String() string {
	return "num"
}

func (t *numType) NumTrees() int {
	return 1
}

func (t *numType) ExpandTrees(trees []fidget.TreeIndex) ([]fidget.TreeIndex, error) {
	if len(trees) == 1 {
		return trees, nil
	}
	return nil, fmt.Errorf("expected 1 tree, got %d", len(trees))
}

// ----------------------------------------------------------------------------

type vecType struct {
	FidgetNumType

	len int
}

func (t *vecType) Common(other FidgetNumType) (FidgetNumType, error) {
	if other == nil || other == NumType || other == t {
		return t, nil
	}
	return nil, fmt.Errorf("incompatible types: %s and %s", t, other)
}

func (t *vecType) Equal(other FidgetNumType) bool {
	return other == t
}

func (t *vecType) Len() int {
	return t.len
}

func (t *vecType) String() string {
	return fmt.Sprintf("vec%d<num>", t.len)
}

func (t *vecType) NumTrees() int {
	return t.len
}

func (t *vecType) ExpandTrees(trees []fidget.TreeIndex) ([]fidget.TreeIndex, error) {
	if len(trees) == 1 {
		switch t.len {
		case 2:
			return []fidget.TreeIndex{trees[0], trees[0]}, nil
		case 3:
			return []fidget.TreeIndex{trees[0], trees[0], trees[0]}, nil
		case 4:
			return []fidget.TreeIndex{trees[0], trees[0], trees[0], trees[0]}, nil
		}
	} else if len(trees) == t.len {
		return trees, nil
	}
	return nil, fmt.Errorf("expected %d trees, got %d", t.len, len(trees))
}

// ----------------------------------------------------------------------------

// listVecType implements the [build.ListVecType] interface for a list whose
// elements are vectors. The trees of a list-typed [TreeNode] are stored
// flattened in element-major order: element i occupies
// trees[i*elemLen : (i+1)*elemLen].
type listVecType struct {
	FidgetNumType

	listLen int
	elemLen int
}

func NewListVecType(len int, elemLen int) FidgetNumType {
	return &listVecType{listLen: len, elemLen: elemLen}
}

func (t *listVecType) Common(other FidgetNumType) (FidgetNumType, error) {
	if t.Equal(other) {
		return t, nil
	}
	return nil, fmt.Errorf("incompatible types: %s and %s", t, other)
}

func (t *listVecType) Equal(other FidgetNumType) bool {
	otherList, isListType := other.(*listVecType)
	return isListType && t.listLen == otherList.listLen && t.elemLen == otherList.elemLen
}

func (t *listVecType) Len() int {
	return t.listLen
}

func (t *listVecType) ElemType() FidgetNumType {
	return VecNumType(t.elemLen)
}

func (t *listVecType) String() string {
	return fmt.Sprintf("array<vec%d<num>, %d>", t.elemLen, t.listLen)
}

func (t *listVecType) NumTrees() int {
	return t.listLen * t.elemLen
}

func (t *listVecType) ExpandTrees(trees []fidget.TreeIndex) ([]fidget.TreeIndex, error) {
	if len(trees) == t.NumTrees() {
		return trees, nil
	}
	return nil, fmt.Errorf("expected %d trees, got %d", t.NumTrees(), len(trees))
}
