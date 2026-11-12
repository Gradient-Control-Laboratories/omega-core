package eval

import (
	"fmt"
	"strconv"
	"strings"
)

// viewPresets names the cameras a render can ask for by word. Directions
// point from the look-at toward the camera; "iso-below" looks up at a part
// from beneath with the image turned so the underside reads as the top — the
// view a gripper hanging from its flange needs, without rewriting the
// document to hang upward.
var viewPresets = map[string]LitView{
	"iso":       {Dir: [3]float64{1, 1, 1}},
	"iso-below": {Dir: [3]float64{1, 1, -1}, Up: [3]float64{0, 0, -1}},
	"front":     {Dir: [3]float64{0, -1, 0}},
	"back":      {Dir: [3]float64{0, 1, 0}},
	"left":      {Dir: [3]float64{-1, 0, 0}},
	"right":     {Dir: [3]float64{1, 0, 0}},
	"top":       {Dir: [3]float64{0, 0, 1}, Up: [3]float64{0, 1, 0}},
	"bottom":    {Dir: [3]float64{0, 0, -1}, Up: [3]float64{0, 1, 0}},
}

// ViewPresetNames lists the preset names, for help text.
func ViewPresetNames() []string {
	return []string{"iso", "iso-below", "front", "back", "left", "right", "top", "bottom"}
}

// ParseView reads a camera as a preset name ("iso-below") or a direction
// ("1,1,-1"). The light and axis choices are the caller's and are left zero.
func ParseView(s string) (LitView, error) {
	s = strings.TrimSpace(s)
	if v, ok := viewPresets[s]; ok {
		return v, nil
	}
	dir, err := ParseVec3(s)
	if err != nil {
		return LitView{}, fmt.Errorf("view %q: not a preset (%s) or an x,y,z direction", s, strings.Join(ViewPresetNames(), ", "))
	}
	if dir == [3]float64{} {
		return LitView{}, fmt.Errorf("view %q: the direction is zero", s)
	}
	return LitView{Dir: dir}, nil
}

// ParseViews reads a comma-separated list of cameras. A direction carries its
// own commas, so the list is not split blindly: a preset name is one entry,
// and anything else takes three fields as x,y,z — "iso,1,1,-1,top" is three
// cameras.
func ParseViews(s string) ([]LitView, error) {
	fields := strings.Split(s, ",")
	var views []LitView
	for i := 0; i < len(fields); {
		if v, ok := viewPresets[strings.TrimSpace(fields[i])]; ok {
			views = append(views, v)
			i++
			continue
		}
		if i+3 > len(fields) {
			return nil, fmt.Errorf("views %q: %q is not a preset (%s) and too few fields follow it for an x,y,z direction",
				s, strings.TrimSpace(fields[i]), strings.Join(ViewPresetNames(), ", "))
		}
		v, err := ParseView(strings.Join(fields[i:i+3], ","))
		if err != nil {
			return nil, err
		}
		views = append(views, v)
		i += 3
	}
	return views, nil
}

// ParseVec3 reads "x,y,z".
func ParseVec3(s string) ([3]float64, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 3 {
		return [3]float64{}, fmt.Errorf("%q: want x,y,z", s)
	}
	var v [3]float64
	for i, p := range parts {
		f, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return [3]float64{}, fmt.Errorf("%q: %w", s, err)
		}
		v[i] = f
	}
	return v, nil
}
