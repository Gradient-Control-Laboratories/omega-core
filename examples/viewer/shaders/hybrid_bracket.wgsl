fn sdShape(p: vec3<f32>, time: f32, gsl_param_2_18: f32, gsl_param_6_20: f32) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let d_base = max(max(abs(p.x - 1.0) - 1.0, abs(p.y) - 1.0), abs(p.z - 0.25) - 0.25);
  let d_upright = max(max(abs(p.x - 0.25) - 0.25, abs(p.y) - 1.0), abs(p.z - 1.0) - 1.0);
  let d_shell = min(d_base, d_upright);
  let gsl_partLo = vec3<f32>(0.0, -1.0, 0.0);
  let gsl_partHi = vec3<f32>(2.0, 1.0, 2.0);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  let infill_density = gsl_param_2_18;
  let sp = p * infill_density;
  let gyroid = sin(sp.x) * cos(sp.y) + sin(sp.y) * cos(sp.z) + sin(sp.z) * cos(sp.x);
  let infill_thickness = gsl_param_6_20;
  let d_gyroid = (abs(gyroid) - infill_thickness) / infill_density;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = max(d_gyroid, d_shell + 0.08) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let d_infill = gsl_part;
  let gsl_s1 = gsl_shape(min(d_shell, d_infill), 1u);
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