fn gsl_remap(p: vec3<f32>) -> f32 {
  let s = builtin_gcl_closure_boxCentered(p, vec3<f32>(10.0, 20.0, 30.0) * 0.5, abs(vec3<f32>(10.0, 20.0, 30.0)));
  return s;
}

fn gsl_remap_2(p: vec3<f32>) -> f32 {
  return gsl_remap(p - vec3<f32>(40.0, 0.0, 0.0));
}

fn gsl_remap_3(p: vec3<f32>) -> f32 {
  return gsl_remap(p - vec3<f32>(80.0, 0.0, 0.0));
}

fn gsl_remap_4(p: vec3<f32>) -> f32 {
  return gsl_remap(p - vec3<f32>(120.0, 0.0, 0.0));
}

fn gsl_remap_5(p: vec3<f32>) -> vec3<f32> {
  return vec3<f32>(clamp(p.x / 10.0, 0.0, 1.0), clamp(p.y / 20.0, 0.0, 1.0), 1.0);
}

fn builtin_intersectSharp(fields: vec3<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_intersectRound(fields: vec3<f32>, radius: f32) -> f32 {
  let sharp = builtin_intersectSharp(fields);
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return sharp - max(sharp + radius, 0.0) + sqrt(max(sumOfSquares, 0.000000000000000000000000000001));
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

fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(0.0, -14.382766, 0.0);
  let gsl_partHi = vec3<f32>(10.0, 17.551651, 35.91599);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = gsl_remap(vec3<f32>(p.x, p.y * 0.87758255 + p.z * 0.47942555, -p.y * 0.47942555 + p.z * 0.87758255)) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let a = gsl_part;
  let gsl_partLo_2 = vec3<f32>(35.103302, 0.0, -23.971277);
  let gsl_partHi_2 = vec3<f32>(58.261894, 20.0, 7.1504555);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = gsl_remap_2(vec3<f32>(p.x * 0.87758255 - p.z * 0.47942555, p.y, p.z * 0.87758255 + p.x * 0.47942555)) + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let b = gsl_part_2;
  let gsl_partLo_3 = vec3<f32>(60.618095, 38.354042, 0.0);
  let gsl_partHi_3 = vec3<f32>(78.98243, 60.69995, 30.0);
  let gsl_partBox_3 = gsl_boxDistance(p, gsl_partLo_3, gsl_partHi_3);
  var gsl_part_3 = gsl_partBox_3;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_3 <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo_3, gsl_partHi_3);
    gsl_part_3 = 1e10;
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = gsl_remap_3(vec3<f32>(p.x * 0.87758255 + p.y * 0.47942555, -p.x * 0.47942555 + p.y * 0.87758255, p.z)) + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let c = gsl_part_3;
  let gsl_partLo_4 = vec3<f32>(120.0, -10.170152, -3.3900504);
  let gsl_partHi_4 = vec3<f32>(140.78224, 19.387913, 33.10758);
  let gsl_partBox_4 = gsl_boxDistance(p, gsl_partLo_4, gsl_partHi_4);
  var gsl_part_4 = gsl_partBox_4;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_4 <= gsl_vanishView.w) {
    let gsl_partVanish_4 = gsl_vanish(gsl_partLo_4, gsl_partHi_4);
    gsl_part_4 = 1e10;
    if (gsl_partVanish_4.x > 0.0) { gsl_part_4 = gsl_remap_4(vec3<f32>(120.0, 0.0, 0.0) + (p - vec3<f32>(120.0, 0.0, 0.0)) * 0.87758255 - cross(vec3<f32>(0.70710677, 0.70710677, 0.0), p - vec3<f32>(120.0, 0.0, 0.0)) * 0.47942555 + vec3<f32>(0.70710677, 0.70710677, 0.0) * dot(vec3<f32>(0.70710677, 0.70710677, 0.0), p - vec3<f32>(120.0, 0.0, 0.0)) * 0.122417435) + gsl_partVanish_4.y; }
    if (abs(gsl_part_4) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_4), gsl_partVanish_4.zw); }
  }
  let d = gsl_part_4;
  let gsl_s1 = gsl_shape(a, 1u);
  let gsl_s2 = gsl_shape(b, 2u);
  let gsl_s3 = gsl_shape(c, 3u);
  let gsl_s4 = gsl_shape(d, 4u);
  return gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4);
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32) -> vec4<f32> {
  if (gsl_colorIdx == 1u) { return vec4<f32>(clamp((gsl_remap_5(vec3<f32>(p.x, p.y * 0.87758255 + p.z * 0.47942555, -p.y * 0.47942555 + p.z * 0.87758255)) * 1.0 + gsl_remap_5(vec3<f32>(p.x, p.y * 0.87758255 + p.z * 0.47942555, -p.y * 0.47942555 + p.z * 0.87758255)) * 0.0) / 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), 1.0); }
  if (gsl_colorIdx == 2u) { return vec4<f32>(clamp((gsl_remap_5(vec3<f32>(p.x * 0.87758255 - p.z * 0.47942555, p.y, p.z * 0.87758255 + p.x * 0.47942555) - vec3<f32>(40.0, 0.0, 0.0)) * 1.0 + gsl_remap_5(vec3<f32>(p.x * 0.87758255 - p.z * 0.47942555, p.y, p.z * 0.87758255 + p.x * 0.47942555) - vec3<f32>(40.0, 0.0, 0.0)) * 0.0) / 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), 1.0); }
  if (gsl_colorIdx == 3u) { return vec4<f32>(clamp((gsl_remap_5(vec3<f32>(p.x * 0.87758255 + p.y * 0.47942555, -p.x * 0.47942555 + p.y * 0.87758255, p.z) - vec3<f32>(80.0, 0.0, 0.0)) * 1.0 + gsl_remap_5(vec3<f32>(p.x * 0.87758255 + p.y * 0.47942555, -p.x * 0.47942555 + p.y * 0.87758255, p.z) - vec3<f32>(80.0, 0.0, 0.0)) * 0.0) / 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), 1.0); }
  if (gsl_colorIdx == 4u) { return vec4<f32>(clamp((gsl_remap_5(vec3<f32>(120.0, 0.0, 0.0) + (p - vec3<f32>(120.0, 0.0, 0.0)) * 0.87758255 - cross(vec3<f32>(0.70710677, 0.70710677, 0.0), p - vec3<f32>(120.0, 0.0, 0.0)) * 0.47942555 + vec3<f32>(0.70710677, 0.70710677, 0.0) * dot(vec3<f32>(0.70710677, 0.70710677, 0.0), p - vec3<f32>(120.0, 0.0, 0.0)) * 0.122417435 - vec3<f32>(120.0, 0.0, 0.0)) * 1.0 + gsl_remap_5(vec3<f32>(120.0, 0.0, 0.0) + (p - vec3<f32>(120.0, 0.0, 0.0)) * 0.87758255 - cross(vec3<f32>(0.70710677, 0.70710677, 0.0), p - vec3<f32>(120.0, 0.0, 0.0)) * 0.47942555 + vec3<f32>(0.70710677, 0.70710677, 0.0) * dot(vec3<f32>(0.70710677, 0.70710677, 0.0), p - vec3<f32>(120.0, 0.0, 0.0)) * 0.122417435 - vec3<f32>(120.0, 0.0, 0.0)) * 0.0) / 1.0, vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 1.0)), 1.0); }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
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
