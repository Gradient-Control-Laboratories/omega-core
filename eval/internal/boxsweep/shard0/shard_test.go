//go:build cgo

// Package shard0 runs shard 0 of the box sweep (package boxsweep) in a test
// binary of its own.
package shard0

import (
	"testing"

	"gradientcontrol.com/omega/core/eval/internal/boxsweep"
)

func TestBoundingBoxSoundness(t *testing.T) {
	boxsweep.Sweep(t, 0, shards)
}

// shards is how many shard packages there are.
const shards = 8
