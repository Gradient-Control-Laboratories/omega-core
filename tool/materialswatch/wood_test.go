package main

import (
	"regexp"
	"testing"

	gsllibs "gradientcontrol.com/omega/core/stdlib"
)

// Every wood species, declared as species(axisOrigin, axisDir) at its default
// look, has a zero-argument @material overload for the browser, so a species
// added to the library without one is caught here rather than missing from
// the panel.
func TestEveryWoodSpeciesIsBrowsable(t *testing.T) {
	speciesRE := regexp.MustCompile(`(?m)^([a-zA-Z]+)\(@point axisOrigin, @direction\(axisOrigin\) axisDir\) \{`)
	browsable := map[string]bool{}
	for _, e := range ReadCatalog() {
		if e.Package == "wood" && e.Kind == "material" {
			browsable[e.Symbol] = true
		}
	}
	species := 0
	for _, lib := range gsllibs.All() {
		if lib.ImportPath != "material/bulk/wood" {
			continue
		}
		for _, src := range lib.Sources {
			for _, m := range speciesRE.FindAllStringSubmatch(src, -1) {
				species++
				if !browsable[m[1]] {
					t.Errorf("wood.%s has no zero-argument @material overload", m[1])
				}
			}
		}
	}
	if species == 0 {
		t.Fatal("no wood species found; the declaration form has changed")
	}
	if len(browsable) != species {
		t.Errorf("%d wood entries in the catalog, %d species declared", len(browsable), species)
	}
}
