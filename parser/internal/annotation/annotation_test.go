package annotation_test

import (
	"testing"

	"gradientcontrol.com/omega/core/parser/internal/annotation"
)

func TestSupportedAnnotations(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{annotation.Range, true},
		{annotation.Values, true},
		{annotation.Point, true},
		{annotation.Transform, true},
		{annotation.Spline, true},
		{annotation.Color, true},
		{annotation.Direction, true},
		{annotation.Frame, true},
		{annotation.Surface, true},
		{"unknown", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := annotation.IsSupported(tt.name); got != tt.want {
			t.Errorf("IsSupported(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestTargetValidation(t *testing.T) {
	tests := []struct {
		name    string
		target  annotation.Target
		wantErr bool
	}{
		// Spline
		{annotation.Spline, annotation.TargetAssignment, false},
		{annotation.Spline, annotation.TargetExpression, false},
		{annotation.Spline, annotation.TargetParameter, true},

		// Point: on a parameter it declares what the argument is
		{annotation.Point, annotation.TargetAssignment, false},
		{annotation.Point, annotation.TargetExpression, false},
		{annotation.Point, annotation.TargetParameter, false},

		// Transform
		{annotation.Transform, annotation.TargetAssignment, false},
		{annotation.Transform, annotation.TargetExpression, false},
		{annotation.Transform, annotation.TargetParameter, false},

		// Color
		{annotation.Color, annotation.TargetAssignment, false},
		{annotation.Color, annotation.TargetExpression, false},
		{annotation.Color, annotation.TargetParameter, false},
		{annotation.Color, annotation.TargetFunction, true},

		// Direction
		{annotation.Direction, annotation.TargetAssignment, false},
		{annotation.Direction, annotation.TargetExpression, false},
		{annotation.Direction, annotation.TargetParameter, false},
		{annotation.Direction, annotation.TargetFunction, true},

		// Frame: a parameter kind alone; a variable has no control for one
		{annotation.Frame, annotation.TargetParameter, false},
		{annotation.Frame, annotation.TargetAssignment, true},
		{annotation.Frame, annotation.TargetExpression, true},
		{annotation.Frame, annotation.TargetFunction, true},

		// Surface: recorded on a function declaration alone
		{annotation.Surface, annotation.TargetFunction, false},
		{annotation.Surface, annotation.TargetAssignment, true},
		{annotation.Surface, annotation.TargetParameter, true},
		{annotation.Surface, annotation.TargetExpression, true},

		// Range
		{annotation.Range, annotation.TargetAssignment, false},
		{annotation.Range, annotation.TargetParameter, false},
		{annotation.Range, annotation.TargetExpression, false},

		// Values
		{annotation.Values, annotation.TargetAssignment, false},
		{annotation.Values, annotation.TargetParameter, true},
		{annotation.Values, annotation.TargetExpression, true},

		// Unknown
		{"unknown", annotation.TargetAssignment, true},
	}
	for _, tt := range tests {
		err := annotation.ValidateTarget(tt.name, tt.target)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateTarget(%q, %v) error = %v, wantErr = %v", tt.name, tt.target, err, tt.wantErr)
		}
	}
}

func TestArgCountValidation(t *testing.T) {
	tests := []struct {
		name    string
		count   int
		wantErr bool
	}{
		// Spline, Point, and Transform take 0 arguments
		{annotation.Spline, 0, false},
		{annotation.Spline, 1, true},
		{annotation.Spline, 2, true},
		{annotation.Point, 0, false},
		{annotation.Point, 1, true},
		{annotation.Point, 2, true},
		{annotation.Transform, 0, false},
		{annotation.Transform, 1, true},
		{annotation.Transform, 2, true},
		{annotation.Color, 0, false},
		{annotation.Color, 1, true},
		{annotation.Direction, 0, false},
		{annotation.Direction, 1, false}, // @direction(anchor)
		{annotation.Direction, 2, true},
		{annotation.Frame, 0, false},
		{annotation.Frame, 1, true},
		{annotation.Surface, 0, false},
		{annotation.Surface, 1, true},

		// Range takes 2 or 3 arguments on declarations
		{annotation.Range, 0, true},
		{annotation.Range, 1, true},
		{annotation.Range, 2, false},
		{annotation.Range, 3, false},
		{annotation.Range, 4, true},

		// Values takes at least 2 arguments
		{annotation.Values, 0, true},
		{annotation.Values, 1, true},
		{annotation.Values, 2, false},
		{annotation.Values, 5, false},
	}
	for _, tt := range tests {
		err := annotation.ValidateArgCount(tt.name, tt.count)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateArgCount(%q, %d) error = %v, wantErr = %v", tt.name, tt.count, err, tt.wantErr)
		}
	}
}

func TestTargetBuiltin(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{annotation.Point, "uiPoint"},
		{annotation.Transform, "uiTransform"},
		{annotation.Spline, "uiSpline"},
		{annotation.Color, "uiColor"},
		{annotation.Direction, "uiDirection"},
		{annotation.Range, ""},
		{annotation.Values, ""},
		{annotation.Frame, ""},
		{annotation.Surface, ""},
	}
	for _, tt := range tests {
		if got := annotation.TargetBuiltin(tt.name); got != tt.want {
			t.Errorf("TargetBuiltin(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestRequiresElementPromotion(t *testing.T) {
	if !annotation.RequiresElementPromotion(annotation.Spline) {
		t.Errorf("RequiresElementPromotion(Spline) = false, want true")
	}
	if annotation.RequiresElementPromotion(annotation.Point) {
		t.Errorf("RequiresElementPromotion(Point) = true, want false")
	}
	if annotation.RequiresElementPromotion(annotation.Transform) {
		t.Errorf("RequiresElementPromotion(Transform) = true, want false")
	}
}
