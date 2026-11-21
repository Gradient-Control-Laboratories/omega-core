package codegen

import (
	"fmt"
	"strings"
)

// fnFrame holds the state for one level of function compilation.
// We use a stack so that nested compilations (A calls B which hasn't been compiled
// yet) don't clobber each other's body buffers.
type fnFrame struct {
	name   string
	params []string
	body   strings.Builder
	names  map[string]bool // JS identifiers declared in this function: params and locals
}

func (d *Delegate) inFunction() bool {
	return len(d.fnStack) > 0
}

func (d *Delegate) currentFrame() *fnFrame {
	return d.fnStack[len(d.fnStack)-1]
}

func (d *Delegate) pushFrame(name string, params []string) {
	names := make(map[string]bool, len(params))
	for _, p := range params {
		names[toJSIdentifier(p)] = true
	}
	d.fnStack = append(d.fnStack, &fnFrame{name: name, params: params, names: names})
}

// topNames returns the JS identifiers declared at the top of computeAudio:
// its parameter, the built-in constants it opens with, and the top-level
// declarations so far.
func (d *Delegate) topNames() map[string]bool {
	if d.topDeclNames == nil {
		d.topDeclNames = map[string]bool{"gsl_time": true, "gsl_p": true, "gsl_pi": true, "gsl_e": true}
	}
	return d.topDeclNames
}

// uniqueJSName returns name, or name with the first free numeric suffix if
// names already holds it, and records the result in names.
func uniqueJSName(names map[string]bool, name string) string {
	unique := name
	for k := 2; names[unique]; k++ {
		unique = fmt.Sprintf("%s_%d", name, k)
	}
	names[unique] = true
	return unique
}

func (d *Delegate) popFrame() *fnFrame {
	n := len(d.fnStack)
	frame := d.fnStack[n-1]
	d.fnStack = d.fnStack[:n-1]
	return frame
}
