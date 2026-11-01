package fix

// Fixer defines a single named source-to-source refactoring transformation on a .gsl file.
type Fixer struct {
	Name  string
	Doc   string
	Help  string
	Apply func(src string) (out string, changed bool, err error)
}

var registry = []Fixer{
	sceneFixer,
}

// All returns a slice of all registered fixers in deterministic execution order.
func All() []Fixer {
	out := make([]Fixer, len(registry))
	copy(out, registry)
	return out
}

// ByName looks up a registered fixer by name.
func ByName(name string) (Fixer, bool) {
	for _, f := range registry {
		if f.Name == name {
			return f, true
		}
	}
	return Fixer{}, false
}

// Apply executes the selected fixers (or all registered fixers when active is empty/nil)
// sequentially against src, returning the transformed source and whether any fixer modified it.
func Apply(src string, active []Fixer) (string, bool, error) {
	fixers := active
	if len(fixers) == 0 {
		fixers = registry
	}
	curr := src
	anyChanged := false
	for _, f := range fixers {
		next, changed, err := f.Apply(curr)
		if err != nil {
			return src, false, err
		}
		if changed {
			curr = next
			anyChanged = true
		}
	}
	return curr, anyChanged, nil
}
