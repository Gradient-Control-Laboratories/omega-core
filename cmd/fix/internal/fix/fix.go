package fix

// Note reports a manual-migration diagnostic produced by a fixer.
type Note struct {
	Line    int
	Message string
}

// Fixer defines a single named source-to-source refactoring transformation on a .gsl file.
type Fixer struct {
	Name           string
	Doc            string
	Help           string
	Apply          func(src string) (out string, changed bool, err error)
	ApplyWithNotes func(src string) (out string, changed bool, notes []Note, err error)
}

var registry = []Fixer{
	sceneFixer,
	rangeFixer,
	stdlibFixer,
	omegaImportFixer,
	domainFixer,
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
	out, changed, _, err := ApplyWithNotes(src, active)
	return out, changed, err
}

// ApplyWithNotes executes the selected fixers (or all registered fixers when active is empty/nil)
// sequentially against src, returning the transformed source, whether any fixer modified it,
// and any manual-migration notes emitted.
func ApplyWithNotes(src string, active []Fixer) (string, bool, []Note, error) {
	fixers := active
	if len(fixers) == 0 {
		fixers = registry
	}
	curr := src
	anyChanged := false
	var notes []Note
	for _, f := range fixers {
		var (
			next    string
			changed bool
			fNotes  []Note
			err     error
		)
		if f.ApplyWithNotes != nil {
			next, changed, fNotes, err = f.ApplyWithNotes(curr)
		} else {
			next, changed, err = f.Apply(curr)
		}
		if err != nil {
			return src, false, nil, err
		}
		if len(fNotes) > 0 {
			notes = append(notes, fNotes...)
		}
		if changed {
			curr = next
			anyChanged = true
		}
	}
	return curr, anyChanged, notes, nil
}
