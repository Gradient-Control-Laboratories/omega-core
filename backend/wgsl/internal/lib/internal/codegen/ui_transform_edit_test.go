package codegen_test

import (
	"fmt"
	"strings"
	"testing"

	assembler "gradientcontrol.com/omega/core/internal/assembler"
	"gradientcontrol.com/omega/core/internal/edit"
	"gradientcontrol.com/omega/core/internal/ui"
)

func TestSplineTransformEdits(t *testing.T) {
	source := `frame(point, tangent) {
  x = normalize(tangent);
  up = abs(x.z) < 0.9 ? vector.dirZ : vector.dirY;
  y = normalize(cross(up, x));
  z = cross(x, y);
  return [[x.x,x.y,x.z,0],[y.x,y.y,y.z,0],[z.x,z.y,z.z,0],[point.x,point.y,point.z,1]];
}
branch() {
  point = [40,0,-20];
  normal = [-1,0,0];
  bendPoint = point + normal * 15;
  endPoint = [0,0,point.z];
  start = uiTransform(frame(point, normal));
  bend = uiTransform(frame(bendPoint, normal));
  end = uiTransform(frame(endPoint, endPoint - bendPoint));
  path = uiSpline([start,bend,end]);
  return length(vector.multiplyPoint(path[0],p)) + length(vector.multiplyPoint(path[1],p)) + length(vector.multiplyPoint(path[2],p));
}
return branch();`
	for first := range 3 {
		t.Run(fmt.Sprint(first), func(t *testing.T) {
			src := source
			for step := 0; step <= 3; step++ {
				res, err := assembler.Assemble([]string{src}, mockMgr)
				if err != nil {
					t.Fatal(err)
				}
				var spline ui.Control
				var spans []edit.Span
				for _, control := range ui.Extract(res.Program.Primary, res.Program.GetImport) {
					if control.Kind == ui.KindSpline {
						spline = control
					}
					for _, param := range control.Params {
						spans = append(spans, param.Span)
					}
				}
				if len(spline.Controls) != 3 || len(spans) != step*16 {
					t.Fatalf("step %d: missing editable spline controls: %+v", step, spline)
				}
				render, err := renderHelper(src, res.Program)
				if err != nil {
					t.Fatalf("step %d: %v", step, err)
				}
				for _, span := range spans {
					if strings.Count(render.WGSL, span.VarName()) < 2 {
						t.Fatalf("step %d: transform parameter was folded away: %s", step, span.VarName())
					}
				}
				if step == 3 {
					break
				}
				handle := spline.Controls[(first+step)%3]
				span := handle.Span
				lines := strings.Split(src, "\n")
				line := lines[span.StartLine-1]
				if !strings.HasPrefix(line[span.StartColumn-1:span.EndColumn-1], "frame(") {
					t.Fatalf("gumball does not target the complete frame: %+v", span)
				}
				e := handle.Value
				matrix := fmt.Sprintf("[[0,1,0,0],[-1,0,0,0],[0,0,1,0],[%g,%g,%g,1]]", e[12]+1, e[13]+2, e[14]+3)
				lines[span.StartLine-1] = line[:span.StartColumn-1] + matrix + line[span.EndColumn-1:]
				src = strings.Join(lines, "\n")
			}
		})
	}
}
