fn builtin_cropToBox(p: vec3<f32>, shape: f32, low: vec3<f32>, high: vec3<f32>) -> f32 {
  let lowCorner = min(low, high) - 0.01;
  let highCorner = max(low, high) + 0.01;
  let d = abs(p - (lowCorner + highCorner) * 0.5) - (highCorner - lowCorner) * 0.5;
  let outside = length(vec3<f32>(max(d.x, 0.0), max(d.y, 0.0), max(d.z, 0.0)));
  let inside = min(max(d.x, max(d.y, d.z)), 0.0);
  let gsl_capped = max(shape, outside + inside);
  if (gsl_boxCap) { return gsl_capped; }
  if (gsl_capped <= max(shape, 0.0)) { return shape; }
  return select(-1.0, 1.0, shape >= 0.0) * max(abs(shape), gsl_capped);
}

var<private> gsl_boxCap: bool = false;

fn builtin_torus(p: vec3<f32>, majorRadius: f32, minorRadius: f32) -> f32 {
  let radial = length(p.xy) - majorRadius;
  return length(vec2<f32>(radial, p.z)) - minorRadius;
}

fn gsl_remap(p: vec3<f32>) -> f32 {
  return builtin_cropToBox(p, builtin_torus(p, 3.0, 0.25), vec3<f32>(-3.25, -3.25, -0.25), vec3<f32>(3.25, 3.25, 0.25));
}

fn gsl_remap_2(p: vec3<f32>) -> f32 {
  return builtin_cropToBox(p, builtin_torus(p, 2.0, 0.25), vec3<f32>(-2.25, -2.25, -0.25), vec3<f32>(2.25, 2.25, 0.25));
}

fn gsl_remap_3(p: vec3<f32>) -> f32 {
  let ball = builtin_cropToBox(p, length(p) - 1.0, vec3<f32>(-1.0, -1.0, -1.0), vec3<f32>(1.0, 1.0, 1.0));
  return ball;
}

fn gsl_remap_4(p: vec3<f32>) -> f32 {
  return builtin_cropToBox(p, min(builtin_cropToBox(p, max(max(length(p.xy) - 2.0, -p.z), p.z - 1.0), vec3<f32>(-2.0, -2.0, 0.0), vec3<f32>(2.0, 2.0, 1.0)), builtin_cropToBox(p, max(max(length(p.xy) - 1.0, -3.0 - p.z), p.z), vec3<f32>(-1.0, -1.0, -3.0), vec3<f32>(1.0, 1.0, 0.0))), vec3<f32>(-2.0, -2.0, -3.0), vec3<f32>(2.0, 2.0, 1.0));
}

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let a = gsl_remap(vec3<f32>(p.x - 5.0, p.y, p.z));
  let gsl_partVanish = gsl_vanish(vec3<f32>(-2.26, 2.74, -0.26), vec3<f32>(2.26, 7.26, 0.26));
  var gsl_part = 1e10;
  if (gsl_partVanish.x > 0.0) { gsl_part = gsl_remap_2(p - vec3<f32>(0.0, 5.0, 0.0)) + gsl_partVanish.y; }
  if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  let b = gsl_part;
  let c = gsl_remap_3(vec3<f32>(p.x, p.y, p.z - 4.0));
  let d = gsl_remap_4(vec3<f32>(p.x + 8.0, p.y, p.z));
  let gsl_s1 = gsl_shape(a, 1u);
  let gsl_s2 = gsl_shape(b, 2u);
  let gsl_s3 = gsl_shape(c, 3u);
  let gsl_s4 = gsl_shape(d, 4u);
  return gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4);
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
