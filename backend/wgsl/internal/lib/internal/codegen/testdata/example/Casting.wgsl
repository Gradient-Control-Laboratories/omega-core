fn builtin_plane(p: vec3<f32>, point: vec3<f32>, normal: vec3<f32>) -> f32 {
  return dot(p - point, normalize(normal));
}

fn builtin_intersectChamfer(d1: f32, d2: f32, setback: f32) -> f32 {
  let face = max(d1, d2);
  let edge = (d1 + d2 + setback) / 1.4142135;
  return max(face, edge);
}

fn builtin_intersectRound(d1: f32, d2: f32, radius: f32) -> f32 {
  let inset1 = max(d1 + radius, 0.0);
  let inset2 = max(d2 + radius, 0.0);
  return min(-radius, max(d1, d2)) + length(vec2<f32>(inset1, inset2));
}

fn builtin_unionRound(d1: f32, d2: f32, radius: f32) -> f32 {
  return -builtin_intersectRound(-d1, -d2, radius);
}

fn gsl_remap(p: vec3<f32>, vector_zero: vec3<f32>, vector_dirZ: vec3<f32>, vector_dirX: vec3<f32>) -> f32 {
  let gsl_tmp_5 = dot(p - vec3<f32>(1.75, 0.0, 0.0), normalize(vec3<f32>(1.0, 0.0, 0.0)));
  return builtin_unionRound(builtin_intersectChamfer(builtin_axis(p, vec3<f32>(0.0, 0.0, 3.5), vector_dirX) - 0.75, gsl_tmp_5 - 0.125, 0.03), gsl_tmp_5, 0.03);
}

fn gsl_remap_2(p: vec3<f32>, vector_zero: vec3<f32>, vector_dirZ: vec3<f32>, vector_dirX: vec3<f32>) -> f32 {
  return builtin_intersectRound(builtin_plane(p, vec3<f32>(0.0, 1.25, 0.5), vec3<f32>(0.0, 1.0, 0.1)), builtin_plane(p, vec3<f32>(0.0, 0.0, 4.5), vec3<f32>(0.0, 0.0, 1.0)), 0.5);
}

fn gsl_remap_3(p: vec3<f32>, vector_zero: vec3<f32>, vector_dirZ: vec3<f32>, vector_dirX: vec3<f32>) -> f32 {
  return builtin_unionRound(builtin_plane(p, vec3<f32>(1.25, 0.0, 0.0), vec3<f32>(-1.0, 0.0, 0.0)), builtin_plane(p, vec3<f32>(0.0, 0.0, 1.0), vec3<f32>(0.0, 0.0, 1.0)), 0.5);
}

fn gsl_remap_4(p: vec3<f32>, vector_zero: vec3<f32>, vector_dirZ: vec3<f32>, vector_dirX: vec3<f32>) -> f32 {
  let gsl_tmp_4 = dot(p - vec3<f32>(0.0, 0.0, 0.5), normalize(vec3<f32>(0.0, 0.0, 1.0)));
  return builtin_intersectChamfer(max(max(max(builtin_axis(p, vector_zero, vector_dirZ) - 3.0, builtin_plane(p, vector_zero, vec3<f32>(0.0, 0.0, -1.0))), -min(min(builtin_axis(p, vec3<f32>(2.5, 0.0, 0.0), vector_dirZ) - 0.25, builtin_axis(p, vec3<f32>(0.00000000000000015308085, 2.5, 0.0), vector_dirZ) - 0.25), min(builtin_axis(p, vec3<f32>(-2.5, 0.0000000000000003061617, 0.0), vector_dirZ) - 0.25, builtin_axis(p, vec3<f32>(-0.0000000000000004592425, -2.5, 0.0), vector_dirZ) - 0.25))), builtin_intersectChamfer(builtin_intersectChamfer(builtin_unionRound(gsl_remap(vec3<f32>(abs(p.x), abs(p.y), p.z), vector_zero, vector_dirZ, vector_dirX), gsl_tmp_4, 0.03), builtin_unionRound(gsl_remap_2(vec3<f32>(p.x, abs(p.y), p.z), vector_zero, vector_dirZ, vector_dirX), gsl_tmp_4, 0.5), 0.03), gsl_remap_3(vec3<f32>(abs(p.x), p.y, p.z), vector_zero, vector_dirZ, vector_dirX), 0.03)), -(builtin_axis(p, vec3<f32>(0.0, 0.0, 3.5), vector_dirX) - 0.5), 0.06);
}

fn gsl_remap_5(p: vec3<f32>, vector_zero: vec3<f32>, vector_dirZ: vec3<f32>, vector_dirX: vec3<f32>) -> f32 {
  return gsl_remap_4(vec3<f32>(p.x, p.y - 5.0, -p.z), vector_zero, vector_dirZ, vector_dirX);
}

fn builtin_intersectSharp_2(d1: f32, d2: f32) -> f32 {
  return max(d1, d2);
}

fn gsl_remap_6(p: vec3<f32>, vector_zero: vec3<f32>, vector_dirZ: vec3<f32>, vector_dirX: vec3<f32>) -> f32 {
  return min(min(builtin_intersectSharp_2(builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 5.0, 0.5), vec3<f32>(3.5, 2.5, 1.0)), 1.25 - abs(p.x)), builtin_gcl_closure_cylinderCapped(p, vec3<f32>(0.0, 5.0, 1.0), vec3<f32>(0.0, 5.0, 2.0), 2.5)), builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, 1.5), vec3<f32>(3.0, 10.0, 1.0)));
}

fn builtin_circleXY(p: vec3<f32>, center: vec2<f32>, radius: f32) -> f32 {
  return length(p.xy - center) - radius;
}

fn builtin_linearMap(gsl_var_value: f32, inMin: f32, inMax: f32, outMin: f32, outMax: f32) -> f32 {
  let slope = (outMax - outMin) / (inMax - inMin);
  return (gsl_var_value - inMin) * slope + outMin;
}

fn builtin_intersectSharp_4(d1: f32, d2: f32, d3: f32) -> f32 {
  return max(max(d1, d2), d3);
}

fn gsl_remap_7(p: vec3<f32>, vector_zero: vec3<f32>, vector_dirZ: vec3<f32>, vector_dirX: vec3<f32>) -> f32 {
  return builtin_intersectSharp_4(builtin_linearMap(p.z, -4.5, -7.5, builtin_gcl_closure_rectangleXYCentered(p, vec2<f32>(1.5, 5.0), vec2<f32>(0.5, 1.0)), builtin_circleXY(p, vec2<f32>(1.5, 5.0), 0.75)), p.z - -4.5, -p.z + -7.5);
}

fn gsl_remap_8(p: vec3<f32>, vector_zero: vec3<f32>, vector_dirZ: vec3<f32>, vector_dirX: vec3<f32>) -> f32 {
  return gsl_remap_7(ease_gcl_closure_interp(p.z, -4.5, -7.5, p, p - vec3<f32>(-1.5, -2.75, 0.0)), vector_zero, vector_dirZ, vector_dirX) * 0.5;
}

fn builtin_ramp(gsl_var_value: f32, inMin: f32, inMax: f32, outMin: f32, outMax: f32) -> f32 {
  let mapped = builtin_linearMap(gsl_var_value, inMin, inMax, outMin, outMax);
  let low = min(outMin, outMax);
  let high = max(outMin, outMax);
  return clamp(mapped, low, high);
}

fn builtin_cylinder(p: vec3<f32>, center: vec3<f32>, direction: vec3<f32>, radius: f32) -> f32 {
  return builtin_axis(p, center, direction) - radius;
}

fn builtin_cylinderZ(p: vec3<f32>, vector_dirZ: vec3<f32>, center: vec3<f32>, radius: f32) -> f32 {
  return builtin_cylinder(p, center, vector_dirZ, radius);
}

fn builtin_bezierQuadPointRound(p: vec3<f32>, start: vec3<f32>, control: vec3<f32>, end: vec3<f32>) -> gsl_struct {
  let d = start - p;
  let e = control - start;
  let g = e * 2.0;
  let f = start - control * 2.0 + end;
  let ff = dot(f, f);
  let ee = dot(e, e);
  let kk = 1.0 / max(ff, 0.000001 * ee + 0.00000000000000000001);
  let ef = dot(e, f);
  let kx = kk * ef;
  let df = dot(d, f);
  let ky = kk * (2.0 * ee + df) / 3.0;
  let de = dot(d, e);
  let kz = kk * de;
  let qq = kx * (2.0 * kx * kx - 3.0 * ky) + kz;
  let pp = ky - kx * kx;
  let zz = sqrt(max(-pp, 0.0));
  let v = acos(clamp(qq / (2.0 * pp * zz - 0.00000000000000000001), -1.0, 1.0)) / 3.0;
  let m = cos(v);
  let ta = clamp((m + m) * zz - kx, 0.0, 1.0);
  let eva = d + (g + f * ta) * ta;
  let da = dot(eva, eva);
  let n = sin(v) * 1.7320508;
  let tb = clamp((-n - m) * zz - kx, 0.0, 1.0);
  let evb = d + (g + f * tb) * tb;
  let dbb = dot(evb, evb);
  let tc = clamp((n - m) * zz - kx, 0.0, 1.0);
  let evc = d + (g + f * tc) * tc;
  let dc = dot(evc, evc);
  let d3 = min(min(da, dbb), dc);
  let h = qq * qq + 4.0 * pp * pp * pp;
  let sh = sqrt(max(h, 0.0));
  let w1 = (sh - qq) / 2.0;
  let u1 = sign(w1) * exp(log(max(abs(w1), 0.000000000001)) / 3.0);
  let w2 = (-sh - qq) / 2.0;
  let u2 = sign(w2) * exp(log(max(abs(w2), 0.000000000001)) / 3.0);
  let t1 = clamp(u1 + u2 - kx, 0.0, 1.0);
  let ev1 = d + (g + f * t1) * t1;
  let d1 = dot(ev1, ev1);
  let tLine = clamp(-dot(d, g) / (dot(g, g) + 0.00000000000000000001), 0.0, 1.0);
  let evLine = d + (g + f * tLine) * tLine;
  let dLine = dot(evLine, evLine);
  let straight = ff < 0.000001 * ee;
  let dist2 = select(select(d3, d1, h >= 0.0), dLine, straight);
  let distanceToCurve = sqrt(dist2);
  let tBest3 = select(select(tc, tb, dbb <= dc), select(tc, ta, da <= dc), da <= dbb);
  let tBest = select(select(tBest3, t1, h >= 0.0), tLine, straight);
  let closestPoint = start + g * tBest + f * tBest * tBest;
  var gsl_result : gsl_struct;
  gsl_result.v_1 = distanceToCurve;
  gsl_result.v_2 = tBest;
  gsl_result.v_3 = closestPoint;
  return gsl_result;
}

fn builtin_bezierQuadRound(p: vec3<f32>, start: vec3<f32>, control: vec3<f32>, end: vec3<f32>) -> f32 {
  let gsl_result = builtin_bezierQuadPointRound(p, start, control, end);
  let distanceToCurve = gsl_result.v_1;
  return distanceToCurve;
}

fn gsl_remap_9(p: vec3<f32>, vector_zero: vec3<f32>, vector_dirZ: vec3<f32>, vector_dirX: vec3<f32>) -> f32 {
  return builtin_bezierQuadRound(p, vec3<f32>(0.0, 5.0, 2.0), vec3<f32>(0.0, 5.0, 6.125), vec3<f32>(0.0, 3.0, 7.5)) - 0.3;
}

fn builtin_axis(p: vec3<f32>, center: vec3<f32>, direction: vec3<f32>) -> f32 {
  let offset = p - center;
  let unitDirection = normalize(direction);
  let along = dot(offset, unitDirection);
  let perp = offset - unitDirection * along;
  return length(perp);
}

fn builtin_intersectSharp(fields: vec3<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_intersectRound_2(fields: vec3<f32>, radius: f32) -> f32 {
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return min(-radius, builtin_intersectSharp(fields)) + sqrt(sumOfSquares);
}

fn builtin_gcl_closure_intersectRound(fields: vec3<f32>) -> f32 {
  return builtin_intersectRound_2(fields, 0.0);
}

fn builtin_gcl_closure_boxCentered(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound(vec3<f32>(d.x, d.y, d.z));
  return solid;
}

fn builtin_intersectSharp_3(fields: vec2<f32>) -> f32 {
  return max(fields.x, fields.y);
}

fn builtin_intersectRound_3(fields: vec2<f32>, radius: f32) -> f32 {
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0);
  return min(-radius, builtin_intersectSharp_3(fields)) + sqrt(sumOfSquares);
}

fn builtin_gcl_closure_intersectRound_2(fields: vec2<f32>) -> f32 {
  return builtin_intersectRound_3(fields, 0.0);
}

fn builtin_gcl_closure_cylinderCapped(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>, radius: f32) -> f32 {
  let pa = p - pointA;
  let ba = pointB - pointA;
  let paba = dot(pa, ba);
  let baba = dot(ba, ba);
  let t = clamp(paba / baba, 0.0, 1.0);
  let radialDistance = length(pa - ba * t) - radius;
  let axialDistance = (abs(paba - baba * 0.5) - baba * 0.5) / sqrt(baba);
  let form = builtin_gcl_closure_intersectRound_2(vec2<f32>(radialDistance, axialDistance));
  return form;
}

fn builtin_gcl_closure_intersectRound_3(fields: vec2<f32>) -> f32 {
  return builtin_intersectRound_3(fields, 0.0);
}

fn builtin_gcl_closure_rectangleXYCentered(p: vec3<f32>, center: vec2<f32>, size: vec2<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p.xy - center) - halfSize;
  return builtin_gcl_closure_intersectRound_3(vec2<f32>(d.x, d.y));
}

fn ease_quadratic(t: f32) -> f32 {
  return t * t * (3.0 - 2.0 * t);
}

fn ease_gcl_closure_interp(gsl_var_value: f32, inMin: f32, inMax: f32, outMin: vec3<f32>, outMax: vec3<f32>) -> vec3<f32> {
  let t = clamp((gsl_var_value - inMin) / (inMax - inMin), 0.0, 1.0);
  let s = ease_quadratic(t);
  return outMin + (outMax - outMin) * s;
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let vector_zero = vec3<f32>(0.0, 0.0, 0.0);
  let vector_dirZ = vec3<f32>(0.0, 0.0, 1.0);
  let vector_dirX = vec3<f32>(1.0, 0.0, 0.0);
  let placedParts = gsl_remap_5(vec3<f32>(p.x, abs(p.y), p.z), vector_zero, vector_dirZ, vector_dirX);
  let gates = gsl_remap_6(vec3<f32>(p.x, abs(p.y), p.z), vector_zero, vector_dirZ, vector_dirX);
  let drain = gsl_remap_8(vec3<f32>(abs(p.x), abs(p.y), p.z), vector_zero, vector_dirZ, vector_dirX);
  let gsl_partLo = vec3<f32>(-3.0, -3.0, -11.5);
  let gsl_partHi = vec3<f32>(3.0, 3.0, -4.5);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = builtin_gcl_closure_cylinderCapped(p, vec3<f32>(0.0, 0.0, -8.5), vec3<f32>(0.0, 0.0, -7.5), 3.0) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let sprueBase = gsl_part;
  let sprueShaftBot = max(builtin_cylinderZ(p, vector_dirZ, vector_zero, builtin_ramp(p.z, 2.5, 8.0, 1.5, 4.0)), -p.z + -7.5);
  let sprueShaft = max(builtin_intersectChamfer(sprueShaftBot, p.z - 8.0, 0.5), -(length(p.xy - vec2<f32>(0.0, 0.0)) - 0.375));
  let sprue = min(sprueBase, sprueShaft);
  let vent = gsl_remap_9(vec3<f32>(p.x, abs(p.y), p.z), vector_zero, vector_dirZ, vector_dirX);
  let gsl_s1 = gsl_shape(placedParts, 1u);
  let gsl_s2 = gsl_shape(gates, 2u);
  let gsl_s3 = gsl_shape(drain, 3u);
  let gsl_s4 = gsl_shape(sprue, 4u);
  let gsl_s5 = gsl_shape(vent, 5u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5);
}

var<private> gsl_vanishEye: vec4<f32>;
var<private> gsl_vanishView: vec4<f32>;
var<private> gsl_partNear: vec3<f32>;

fn gsl_vanish(lo: vec3<f32>, hi: vec3<f32>) -> vec4<f32> {
  if (gsl_vanishEye.w <= 0.0) { return vec4<f32>(1.0, 0.0, 1e10, 1e10); }
  let size = hi - lo;
  let dist = length(gsl_vanishEye.xyz - clamp(gsl_vanishEye.xyz, lo, hi));
  let pixel = gsl_vanishEye.w * mix(max(dist, 1.0), 1.0, gsl_vanishView.x);
  let edge = min(size.x, min(size.y, size.z));
  let px = edge / pixel;
  let w = smoothstep(gsl_vanishView.y, max(gsl_vanishView.z, gsl_vanishView.y + 1e-3), px);
  return vec4<f32>(w, (1.0 - w) * 0.5 * edge, px, edge);
}

fn gsl_boxDistance(q: vec3<f32>, lo: vec3<f32>, hi: vec3<f32>) -> f32 {
  return length(max(max(lo - q, q - hi), vec3<f32>(0.0)));
}

struct gsl_struct {
  v_1 : f32,
  v_2 : f32,
  v_3 : vec3<f32>,
}
