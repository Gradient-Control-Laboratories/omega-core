package suggest

import (
	"reflect"
	"testing"
)

func TestDistance(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want int
	}{
		{"cylindrZ", "cylinderZ", 1},
		{"radus", "radius", 1},
		{"lenght", "length", 1}, // a swap of neighbours is one edit
		{"translat", "translate", 1},
		{"abc", "abc", 0},
		{"", "abc", 3},
	} {
		if got := distance(tc.a, tc.b); got != tc.want {
			t.Errorf("distance(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestClosest(t *testing.T) {
	cands := []string{"cylinderZ", "cylinderX", "cylinder", "sphere", "Radius", "radius"}
	if got, want := closest("cylindrZ", cands), []string{"cylinderZ", "cylinder", "cylinderX"}; !reflect.DeepEqual(got, want) {
		t.Errorf("closest(cylindrZ) = %v, want %v", got, want)
	}
	if got := closest("RADIUS", cands); len(got) == 0 || (got[0] != "Radius" && got[0] != "radius") {
		t.Errorf("closest(RADIUS) = %v, want a case-insensitive match first", got)
	}
	if got := closest("zz", cands); len(got) != 0 {
		t.Errorf("closest(zz) = %v, want none", got)
	}
}
