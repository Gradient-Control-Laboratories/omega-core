//go:build cgo

// Package shard5 runs shard 5 of the box sweep (package boxsweep) in a test
// binary of its own.
package shard5

import (
	"testing"

	"gradientcontrol.com/omega/core/eval/internal/boxsweep"
)

func TestBoundingBoxSoundness(t *testing.T) {
	boxsweep.Sweep(t, 5, shards)
}

// shards is how many shard packages there are.
const shards = 8
