package scope

// Call represents a specific function call resolution.
type Call struct {
	Parent        Scope
	CalledFnScope *FnScope
}

var _ NamedElem = (*Call)(nil)

func (c *Call) ContainingScope() Scope {
	return c.Parent
}

func (c *Call) Name() string {
	if c.CalledFnScope != nil && c.CalledFnScope.parent != nil && c.CalledFnScope.parent.parent != nil {
		return c.CalledFnScope.parent.parent.Name()
	}
	return ""
}
