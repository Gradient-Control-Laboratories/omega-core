package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gradientcontrol.com/omega/core/assembler"
	fidgetlib "gradientcontrol.com/omega/core/internal/fidget/lib"
)

// The operators' thumbnails: each operator drawn over one stand-in surface, so
// the panel can show what it does. A catalog swatch is 100 mm across, which is
// too far back to see a tool mark, so these are drawn on a ball the size of a
// small part. Two PNGs per operator, at operators/<package>.<symbol>.png and,
// for the grid's hover preview, operators/<package>.<symbol>.large.png.

const (
	thumbRadius    = 6.0
	thumbSize      = 96
	thumbLargeSize = 240
)

// thumbDoc is the document drawing op over the stand-in: a metal in a mid gray,
// the wrap as the panel writes it, on the ball's own frame.
func thumbDoc(op *Operator) string {
	args := make([]string, 1+max(op.Slots.Color, op.Slots.Frame, 0))
	args[op.Slots.Material] = "thumbBase"
	if op.Slots.Color >= 0 {
		args[op.Slots.Color] = "thumbColor"
	}
	if op.Slots.Frame >= 0 {
		args[op.Slots.Frame] = "[ball.u, ball.v]"
	}
	names := []string{"layer", "layerColor", "layerHeight", "layerSlope"}[:op.Arity]
	var b strings.Builder
	fmt.Fprintf(&b, "ball = sphere(%g);\n", thumbRadius)
	b.WriteString("thumbBase = material.brushedMetal;\nthumbColor = [0.77, 0.78, 0.78];\n")
	fmt.Fprintf(&b, "%s = %s(%s);\n", strings.Join(names, ", "), op.Call, strings.Join(args, ", "))
	if op.Relief {
		b.WriteString("ball.material = material.withDisplacement(layer, layerHeight);\n")
	} else {
		b.WriteString("ball.material = layer;\n")
	}
	if op.Arity >= 2 {
		b.WriteString("ball.color = layerColor;\n")
	} else {
		b.WriteString("ball.color = thumbColor;\n")
	}
	b.WriteString("return ball;\n")
	return b.String()
}

// writeThumbs renders every operator's thumbnail into outDir/operators.
func writeThumbs(outDir string, operators map[string]*Operator, extent float64, env string, light bool) error {
	dir := filepath.Join(outDir, "operators")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	keys := make([]string, 0, len(operators))
	for key := range operators {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	started := time.Now()
	for i, key := range keys {
		res, err := assembler.Assemble([]string{thumbDoc(operators[key])})
		if err != nil {
			return fmt.Errorf("%s: assemble: %w", key, err)
		}
		for suffix, size := range map[string]int{".png": thumbSize, ".large.png": thumbLargeSize} {
			png, err := fidgetlib.RenderLitSizedView(res.Program, size, size, extent*thumbRadius, 0, fidgetlib.LitView{
				Light:        light,
				NoAxes:       true,
				Environment:  env,
				GlassSamples: swatchGlassSamples,
			})
			if err != nil {
				return fmt.Errorf("%s: %w", key, err)
			}
			if err := os.WriteFile(filepath.Join(dir, key+suffix), png, 0o644); err != nil {
				return err
			}
		}
		fmt.Fprintf(os.Stderr, "\roperator %d/%d %-32s", i+1, len(keys), key)
	}
	fmt.Fprintf(os.Stderr, "\rrendered %d operator thumbnails in %s\n", len(keys), time.Since(started).Round(time.Second))
	return nil
}
