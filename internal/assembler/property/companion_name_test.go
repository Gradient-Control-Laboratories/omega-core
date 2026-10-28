package property

import (
	"testing"

	"gradientcontrol.com/omega/core/program"
)

// A companion keeps the plain __d name while its function has one overload,
// and takes the arity suffix only when the name is declared at several
// arities. A namespace lists a dependency's functions more than once (as the
// dependency and again as the dot import naming it), which must not read as
// overloading: Slider Mechanism's normalizeField is one function, listed
// twice, and its companion is normalizeField__d as the golden has it.
func TestCompanionNameCountsOverloadsByArity(t *testing.T) {
	one := &program.Function{Parameters: []*program.Parameter{{Name: "v"}}}
	two := &program.Function{Parameters: []*program.Parameter{{Name: "a"}, {Name: "b"}}}
	c := &companions{own: map[string]int{}, funcs: map[string][]*program.Function{
		"listedTwice": {one, one},
		"overloaded":  {one, two, one},
	}}
	if got := c.nameFor("listedTwice", 1); got != "listedTwice__d" {
		t.Errorf("a function listed twice named %q, want listedTwice__d", got)
	}
	if got := c.nameFor("overloaded", 2); got != "overloaded__d2" {
		t.Errorf("an overloaded function named %q, want overloaded__d2", got)
	}
	if got := c.nameFor("unknown", 1); got != "unknown__d" {
		t.Errorf("an unlisted function named %q, want unknown__d", got)
	}
}
