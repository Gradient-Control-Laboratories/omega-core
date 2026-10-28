package environment

import "testing"

func TestPresetsLoad(t *testing.T) {
	all, err := Presets()
	if err != nil {
		t.Fatal(err)
	}
	viewer := all[Default]
	if viewer.Mode != 0 || viewer.Rig != 1 || len(viewer.Panels) != 0 {
		t.Errorf("%s must be mode 0, rig 1, no panels (the viewer's own lights): %+v", Default, viewer)
	}
	for name, p := range all {
		if p.Mode == 1 && len(p.Panels) == 0 && p.Sky == [3]float64{} && p.Ground == [3]float64{} {
			t.Errorf("%s is analytic but emits no light", name)
		}
		for i, panel := range p.Panels {
			if panel.Axis == [3]float64{} || panel.Sharpness <= 0 {
				t.Errorf("%s panel %d: axis %v, sharpness %v", name, i, panel.Axis, panel.Sharpness)
			}
		}
	}
	if _, err := Lookup("no such preset"); err == nil {
		t.Error("Lookup of an unknown name succeeded")
	}
}
