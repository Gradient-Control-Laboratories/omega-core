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
	exprs := []string{expr}
	if strings.TrimSpace(expr) == "" {
		exprs = TopLevelReturnExprs(code)
		if len(exprs) == 0 {
			return nil, fmt.Errorf("the document returns nothing to render")
		}
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
		png, err := lib.RenderLitSizedView(prog, width, height, 0, 0, v)
		if err != nil {
			return nil, fmt.Errorf("view %s: %w", views[i], err)
		}
		if len(png) == 0 {
			return nil, fmt.Errorf("view %s: renderer returned no image", views[i])
		}
		panels = append(panels, png)
	}
	if len(panels) == 1 {
		return panels[0], nil
	}
	return lib.ComposeSheet(panels, 2)
}
