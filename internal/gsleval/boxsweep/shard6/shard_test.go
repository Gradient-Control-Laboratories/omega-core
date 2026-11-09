// Package shard6 runs shard 6 of the box sweep (package boxsweep) in a test
// binary of its own.
package shard6

import (
	"testing"

	"gradientcontrol.com/omega/core/internal/gsleval/boxsweep"
)

func TestBoundingBoxSoundness(t *testing.T) {
	boxsweep.Sweep(t, 6, shards)
}

// shards is how many shard packages there are.
const shards = 8
