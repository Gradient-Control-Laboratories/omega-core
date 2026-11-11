package eval

import "testing"

// TestParseViews: a -views list mixes presets and directions, and a direction's
// own commas do not split it into three cameras.
func TestParseViews(t *testing.T) {
	views, err := ParseViews("iso, 1,1,-1 ,top,0,-1,0")
	if err != nil {
		t.Fatal(err)
	}
	want := [][3]float64{{1, 1, 1}, {1, 1, -1}, {0, 0, 1}, {0, -1, 0}}
	if len(views) != len(want) {
		t.Fatalf("got %d views, want %d: %+v", len(views), len(want), views)
	}
	for i, w := range want {
		if views[i].Dir != w {
			t.Errorf("view %d = %v, want %v", i, views[i].Dir, w)
		}
	}
	if views[2].Up != [3]float64{0, 1, 0} {
		t.Errorf("top keeps its preset up, got %v", views[2].Up)
	}

	for _, bad := range []string{"iso,1,1", "sideways", "1,1,x"} {
		if _, err := ParseViews(bad); err == nil {
			t.Errorf("ParseViews(%q) should fail", bad)
		}
	}
}
