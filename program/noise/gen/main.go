// Command gen writes the baked noise tile to the path given as its first
// argument. The output is consumed by the WGSL backend's browser runtime
// (a texture_3d the shader samples for noiseHash3); it is exactly the same
// bytes the Fidget oracle bakes in-process, so the two backends agree.
//
// Regenerate with:
//
//	go run ./program/noise/gen examples/viewer/render/noise_tile.bin
package main

import (
	"fmt"
	"os"

	"gradientcontrol.com/omega/core/program/noise"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: gen <output-path>")
		os.Exit(2)
	}
	if err := os.WriteFile(os.Args[1], noise.Bake(), 0644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
