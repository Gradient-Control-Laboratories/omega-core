package gsleval

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"gradientcontrol.com/omega/core/internal/fidget/lib"
)

// Render draws the document's returns — or, when expr is set, that one
// expression in the document's scope — from each named view preset, shaded
// the way the viewer shades it, and tiles the panels two across into one PNG.
// It backs the chat's render tool. `defs` is optional GSL statements in the
// document's scope, as for EvalGsl. An empty view list draws the iso view.
func (e *Evaluator) Render(ctx context.Context, manifest map[string]string, code, defs, expr string, views []string, width, height int) ([]byte, error) {
	if strings.TrimSpace(expr) == "" && len(TopLevelReturnExprs(code)) == 0 {
		return nil, fmt.Errorf("the document returns nothing to render")
	}
	if len(views) == 0 {
		views = []string{"iso"}
	}
	litViews := make([]lib.LitView, 0, len(views))
	for _, name := range views {
		if !slices.Contains(lib.ViewPresetNames(), name) {
			return nil, fmt.Errorf("view %q: not a preset (%s)", name, strings.Join(lib.ViewPresetNames(), ", "))
		}
		v, err := lib.ParseView(name)
		if err != nil {
			return nil, err
		}
		litViews = append(litViews, v)
	}
	return e.render(ctx, manifest, code, defs, expr, litViews, views, width, height, 0, 0)
}

// RenderViews is Render from any cameras -- presets or x,y,z directions, each with its own ground (lib.ParseViews,
// LitView.Light) -- over a scene of half-size extent (0 frames the geometry) at animation time timeAt: what
// `compile -t png` draws, with defs and expr as for Render. An empty view list draws the iso view.
func (e *Evaluator) RenderViews(ctx context.Context, manifest map[string]string, code, defs, expr string, views []lib.LitView, width, height int, extent, timeAt float64) ([]byte, error) {
	if len(views) == 0 {
		views = []lib.LitView{{Dir: [3]float64{1, 1, 1}}}
	}
	if width < 1 || height < 1 {
		return nil, fmt.Errorf("a panel of %dx%d pixels: each side is at least 1", width, height)
	}
	if extent < 0 {
		return nil, fmt.Errorf("extent %g: 0 frames the geometry, or a half-size in mm", extent)
	}
	names := make([]string, len(views))
	for i := range views {
		names[i] = fmt.Sprint(i + 1)
	}
	return e.render(ctx, manifest, code, defs, expr, views, names, width, height, extent, timeAt)
}

// render draws expr (or the document's returns) from each view, named in errors by names, tiled two across.
func (e *Evaluator) render(ctx context.Context, manifest map[string]string, code, defs, expr string, litViews []lib.LitView, names []string, width, height int, extent, timeAt float64) ([]byte, error) {
	exprs := []string{expr}
	if strings.TrimSpace(expr) == "" {
		exprs = TopLevelReturnExprs(code)
		if len(exprs) == 0 {
			return nil, fmt.Errorf("the document returns nothing to render")
		}
	}

	if err := startQuery(ctx); err != nil {
		return nil, err
	}
	defer endQuery()
	prog, err := e.prepareQuery(ctx, manifest, code, defs, exprs)
	if err != nil {
		return nil, err
	}

	panels := make([][]byte, 0, len(litViews))
	for i, v := range litViews {
		png, err := lib.RenderLitSizedView(prog, width, height, extent, timeAt, v)
		if err != nil {
			return nil, fmt.Errorf("view %s: %w", names[i], err)
		}
		if len(png) == 0 {
			return nil, fmt.Errorf("view %s: renderer returned no image", names[i])
		}
		panels = append(panels, png)
	}
	if len(panels) == 1 {
		return panels[0], nil
	}
	return lib.ComposeSheet(panels, 2)
}
