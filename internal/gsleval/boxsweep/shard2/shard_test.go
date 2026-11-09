// Package shard2 runs shard 2 of the box sweep (package boxsweep) in a test
// binary of its own.
package shard2

import (
	"testing"

	"gradientcontrol.com/omega/core/internal/gsleval/boxsweep"
)

func TestBoundingBoxSoundness(t *testing.T) {
	boxsweep.Sweep(t, 2, shards)
}

// shards is how many shard packages there are.
const shards = 8
