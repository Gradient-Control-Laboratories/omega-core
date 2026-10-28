package closure_test

import (
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// TestClosure_ParameterShadowsForeignFunction: a parameter named like a
// function some OTHER package declares is a value. geom/sizes copies its
// `nominal` argument into a closure, as catalog's lookups do; geom/joints
// imports it and declares a function `nominal`. The copy once resolved to that
// function, found by a scan over every dependency in a package geom/sizes does
// not import, and the linker then failed on `nominal` in geom/sizes.
func TestClosure_ParameterShadowsForeignFunction(t *testing.T) {
	mgr := &source.MockManager{PkgSources: map[string][]string{
		"geom/sizes": {`
rows = [[3.0, 0.5], [4.0, 0.7]];
onRow(key, rowKey) {
  return 1.0 - clamp(abs(key - rowKey), 0.0, 1.0);
}
pitch(nominal) {
  k = nominal;
  return reduce((acc, v) { return acc + v; },
                map((r) { return r[1] * onRow(k, r[0]); }, rows));
}
`},
		"geom/joints": {`
import sizes "geom/sizes";
nominal(iface) {
  return iface * 2.0;
}
lead(d) {
  return sizes.pitch(d) + nominal(d);
}
`},
	}}
	src := `
import joints "geom/joints";
return length(p) - joints.lead(3.0);
`
	if _, err := runPipelineWithManager(src, mgr); err != nil {
		t.Fatalf("a parameter named like another package's function broke the build: %v", err)
	}
}

// TestClosure_LocalShadowsPackageFunction: a multi-return function's locals
// are values even when its own package declares functions by those names.
// `return channel, envelope;` once read both as functions and was rejected as a
// function returned alongside other values.
func TestClosure_LocalShadowsPackageFunction(t *testing.T) {
	mgr := &source.MockManager{PkgSources: map[string][]string{
		"geom/lib": {`
slide(w) {
  channel = abs(p.x) - w;
  envelope = abs(p.x) - w * 0.5;
  return channel, envelope;
}
channel(g) {
  c, e = slide(g);
  return c;
}
`},
	}}
	src := `
import lib "geom/lib";
return lib.channel(2.0);
`
	if _, err := runPipelineWithManager(src, mgr); err != nil {
		t.Fatalf("a local named like a package function broke the build: %v", err)
	}
}
