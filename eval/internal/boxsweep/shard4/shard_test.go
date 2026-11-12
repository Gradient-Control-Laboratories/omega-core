//go:build cgo

// Package shard4 runs shard 4 of the box sweep (package boxsweep) in a test
// binary of its own.
package shard4

import (
	"testing"

	"gradientcontrol.com/omega/core/eval/internal/boxsweep"
)

func TestBoundingBoxSoundness(t *testing.T) {
	boxsweep.Sweep(t, 4, shards)
}

// shards is how many shard packages there are.
const shards = 8
