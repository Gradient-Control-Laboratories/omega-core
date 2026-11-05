package scene

import (
	"fmt"

	"gradientcontrol.com/omega/core/program"
)

// New constructs and validates a new Scene protobuf message.
// If grid is nil, grid spacing defaults to size / 10.0.
func New(size float64, grid *float64, start, end *program.Location) (*program.Scene, error) {
	if size <= 0 {
		return nil, fmt.Errorf("scene size must be greater than zero, got %v", size)
	}

	var g float64
	if grid != nil {
		if *grid <= 0 {
			return nil, fmt.Errorf("scene grid spacing must be greater than zero, got %v", *grid)
		}
		g = *grid
	} else {
		g = size / 10.0
	}

	return &program.Scene{
		Start: start,
		End:   end,
		Size:  size,
		Grid:  g,
	}, nil
}

// ResolveWorldExtent returns the worldExtent (half-size of the scene bounding box)
// based on the scene definition, or falls back to the requested fallback or default 10.0.
func ResolveWorldExtent(s *program.Scene, fallback float64) float64 {
	if s != nil && s.Size > 0 {
		return s.Size / 2.0
	}
	if fallback > 0 {
		return fallback
	}
	return 10.0
}
