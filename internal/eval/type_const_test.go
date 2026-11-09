package eval_test

import (
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
	"gradientcontrol.com/omega/core/internal/eval"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/program"
)

// A fold's verdict is stamped on the operand and nowhere else: no type in
// the folded program carries a copy of a constant.
func TestFoldStampsOperandsNotTypes(t *testing.T) {
	prog := analyzed(t, `
k = 2.0 * 3.0;
v = [1.0, 2.0] * k;
f(x) { return x * k + 1.0; }
return length(p.xy - v) - f(0.5);
`)
	_, err := eval.Fold(prog, sliderChecker{})
	assert.NoError(t, err)

	vRoot := prog.Primary.Files[0].Assignments[1].Expression
	assert.True(t, approxEqual(vRoot.ConstValues, []float64{6, 12}), "the vector keeps its stamp on the operand")

	types := 0
	var walk func(m protoreflect.Message)
	walk = func(m protoreflect.Message) {
		if typ, ok := m.Interface().(*program.Type); ok {
			types++
			assert.True(t, typ.ConstValue == nil, "a type carries a constant")
		}
		m.Range(func(fd protoreflect.FieldDescriptor, val protoreflect.Value) bool {
			switch {
			case fd.Message() == nil || fd.IsMap():
			case fd.IsList():
				for i := 0; i < val.List().Len(); i++ {
					walk(val.List().Get(i).Message())
				}
			default:
				walk(val.Message())
			}
			return true
		})
	}
	walk(prog.ProtoReflect())
	assert.True(t, types > 0, "the program should be typed")
}
