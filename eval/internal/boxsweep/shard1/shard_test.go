//go:build cgo

// Package shard1 runs shard 1 of the box sweep (package boxsweep) in a test
// binary of its own.
package shard1

import (
	"testing"

	"gradientcontrol.com/omega/core/eval/internal/boxsweep"
)

func TestBoundingBoxSoundness(t *testing.T) {
	boxsweep.Sweep(t, 1, shards)
}

// shards is how many shard packages there are.
const shards = 8
