fn builtin_torus(p: vec3<f32>, majorRadius: f32, minorRadius: f32) -> f32 {
  let radial = length(p.xy) - majorRadius;
  return length(vec2<f32>(radial, p.z)) - minorRadius;
}

fn builtin_lineSegment(p: vec3<f32>, pointA: vec3<f32>, pointB: vec3<f32>) -> f32 {
  let pa = p - pointA;
  let ba = pointB - pointA;
  let h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
  return length(pa - ba * h);
}

fn builtin_shell(d: f32, thickness: f32) -> f32 {
  return abs(d) - thickness * 0.5;
}

fn builtin_shell_2(d: f32, thickness: f32, bias: f32) -> f32 {
  return abs(d + bias * thickness * 0.5) - thickness * 0.5;
}

fn builtin_twoBodyField(d1: f32, d2: f32) -> f32 {
  return (d1 - d2) / (d1 + d2);
}

fn builtin_mixAngle(a: f32, b: f32, angle: f32) -> f32 {
  return a * cos(angle) + b * sin(angle);
}

fn builtin_intersectSharp_2(fields: vec2<f32>) -> f32 {
  return max(fields.x, fields.y);
}

fn builtin_intersectRvachev(fields: vec2<f32>, band: f32) -> f32 {
  let sharp = builtin_intersectSharp_2(fields);
  let safeBand = max(band, 0.000001);
  let progress = clamp((sharp + safeBand) / safeBand, 0.0, 1.0);
  let weight = progress * progress * (3.0 - 2.0 * progress);
  let sum = 0.0 + fields.x + fields.y;
  let sumOfSquares = 0.0 + fields.x * fields.x + fields.y * fields.y;
  let rvachev = sum + sqrt(sumOfSquares);
  let blended = (1.0 - weight) * sharp + weight * rvachev;
  return select(blended, sharp, sharp < -safeBand);
}

fn builtin_unionRvachev(fields: vec2<f32>, band: f32) -> f32 {
  let negated = vec2<f32>(-fields.x, -fields.y);
  return -builtin_intersectRvachev(negated, band);
}

fn builtin_differenceRvachev(gsl_var_target: f32, tool: f32, band: f32) -> f32 {
  return builtin_intersectRvachev(vec2<f32>(gsl_var_target, -tool), band);
}

fn gsl_remap(p: vec3<f32>) -> f32 {
  let t = builtin_torus(p, 20.0, 5.0);
  return t;
}

fn builtin_intersectSharp(fields: vec3<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_intersectRound(fields: vec3<f32>, radius: f32) -> f32 {
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return min(-radius, builtin_intersectSharp(fields)) + sqrt(sumOfSquares);
}

fn builtin_gcl_closure_intersectRound(fields: vec3<f32>) -> f32 {
  return builtin_intersectRound(fields, 0.0);
}

fn builtin_gcl_closure_boxCentered(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound(vec3<f32>(d.x, d.y, d.z));
  return solid;
}

fn map_rotateAxis(v: vec3<f32>, axisOrigin: vec3<f32>, axisDir: vec3<f32>, angle: f32) -> vec3<f32> {
  let rel = v - axisOrigin;
  let ca = cos(angle);
  let kn = normalize(axisDir);
  let sa = sin(angle);
  return axisOrigin + rel * ca + cross(kn, rel) * sa + kn * dot(kn, rel) * (1.0 - ca);
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let t = builtin_torus(p, 20.0, 5.0);
  let gsl_partLo = vec3<f32>(-12.0, -2.0, -2.0);
  let gsl_partHi = vec3<f32>(12.0, 2.0, 2.0);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = builtin_lineSegment(p, vec3<f32>(-10.0, 0.0, 0.0), vec3<f32>(10.0, 0.0, 0.0)) - 2.0 + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let seg = gsl_part;
  let gsl_partLo_2 = vec3<f32>(-8.75, -8.75, -8.75);
  let gsl_partHi_2 = vec3<f32>(8.75, 8.75, 8.75);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = builtin_shell(length(p) - 8.0, 1.5) + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let hollow = gsl_part_2;
  let gsl_partLo_3 = vec3<f32>(-8.75, -8.75, -8.75);
  let gsl_partHi_3 = vec3<f32>(8.75, 8.75, 8.75);
  let gsl_partBox_3 = gsl_boxDistance(p, gsl_partLo_3, gsl_partHi_3);
  var gsl_part_3 = gsl_partBox_3;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_3 <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo_3, gsl_partHi_3);
    gsl_part_3 = 1e10;
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = builtin_shell_2(length(p) - 8.0, 1.5, 0.5) + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let hollowBias = gsl_part_3;
  let cay = builtin_twoBodyField(length(p) - 8.0, builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(12.0, 12.0, 12.0)));
  let pol = builtin_mixAngle(length(p) - 8.0, builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(12.0, 12.0, 12.0)), 0.6);
  let rv = builtin_intersectRvachev(vec2<f32>(length(p) - 8.0, builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(12.0, 12.0, 12.0))), 1.0);
  let ru = builtin_unionRvachev(vec2<f32>(length(p) - 8.0, builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(12.0, 12.0, 12.0))), 1.0);
  let rd = builtin_differenceRvachev(length(p) - 8.0, builtin_gcl_closure_boxCentered(p, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(12.0, 12.0, 12.0)), 1.0);
  let spun = gsl_remap(map_rotateAxis(p, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 0.0), 0.5));
  let cyl = builtin_gcl_closure_boxCentered(vec3<f32>(length(p.xy), atan2(p.y, p.x) * 40.0 / 6.2831855, p.z), vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(3.0, 3.0, 3.0));
  let gsl_s1 = gsl_shape(min(min(min(t, seg), min(hollow, hollowBias)), min(min(cay, pol), min(min(rv, ru), min(rd, min(spun, cyl))))), 1u);
  return gsl_s1;
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
