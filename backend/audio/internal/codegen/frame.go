package codegen

import "strings"

// fnFrame holds the state for one level of function compilation.
// We use a stack so that nested compilations (A calls B which hasn't been compiled
// yet) don't clobber each other's body buffers.
type fnFrame struct {
	name   string
	params []string
	body   strings.Builder
}

func (d *Delegate) inFunction() bool {
	return len(d.fnStack) > 0
}

func (d *Delegate) currentFrame() *fnFrame {
	return d.fnStack[len(d.fnStack)-1]
}

func (d *Delegate) pushFrame(name string, params []string) {
	d.fnStack = append(d.fnStack, &fnFrame{name: name, params: params})
}

func (d *Delegate) popFrame() *fnFrame {
	n := len(d.fnStack)
	frame := d.fnStack[n-1]
	d.fnStack = d.fnStack[:n-1]
	return frame
}
